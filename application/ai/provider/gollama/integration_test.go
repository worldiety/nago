// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	gollama "github.com/dianlight/gollama.cpp"
	"go.wdy.de/nago/application/ai/completion"
)

// These are real end-to-end tests that load a GGUF model through the llama.cpp dylib and run a greedy
// generation. They are skipped unless NAGO_GOLLAMA_MODEL points at a local *.gguf file.
//
// ensureBackend loads the llama.cpp dylib that ships embedded in the gollama.cpp module (build b6862),
// extracting it into the gollama cache directory on first use, so these tests run under a plain `go test`
// without any DYLD_LIBRARY_PATH setup:
//
//	NAGO_GOLLAMA_MODEL="$HOME/Downloads/Qwen3-1.7B-Q8_0.gguf" \
//		go test -run TestIntegration -v ./application/ai/provider/gollama/

// integrationModel returns the configured model path or skips the test.
func integrationModel(t *testing.T) string {
	t.Helper()
	path := os.Getenv("NAGO_GOLLAMA_MODEL")
	if path == "" {
		t.Skip("set NAGO_GOLLAMA_MODEL=/path/to/model.gguf to run the integration tests")
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("NAGO_GOLLAMA_MODEL %q not accessible: %v", path, err)
	}
	return path
}

// loadIntegrationModel initialises the backend and loads the model, fully offloaded to the GPU. It skips
// (rather than fails) when the backend cannot initialise, e.g. when the embedded llama.cpp library cannot be
// extracted or loaded on this platform.
func loadIntegrationModel(t *testing.T, e *engine, path string) *loadedModel {
	t.Helper()
	if err := e.ensureBackend(); err != nil {
		t.Skipf("backend init failed: %v", err)
	}

	meta, err := readGGUFMetadata(path)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	t.Logf("model: arch=%q tokenizer=%q ctxLen=%d eos=%d eot=%d adapter=%s",
		meta.architecture, meta.tokenizerModel, meta.contextLength, meta.eosTokenID, meta.eotTokenID,
		selectAdapter(familyAuto, meta).name())

	mp := gollama.Model_default_params()
	mp.NGpuLayers = 999
	handle, err := gollama.Model_load_from_file(path, mp)
	if err != nil {
		t.Fatalf("load model: %v", err)
	}
	t.Cleanup(func() { gollama.Model_free(handle) })

	return &loadedModel{
		path:    path,
		handle:  handle,
		meta:    meta,
		adapter: selectAdapter(familyAuto, meta),
	}
}

// TestIntegrationChat verifies a plain chat turn generates non-empty text and terminates on an end-of-turn
// token (i.e. it is not truncated by the max-token cap). This is the primary regression test for the
// context-params SIGSEGV: it exercises Init_from_model, Decode and the greedy sampling loop end to end.
func TestIntegrationChat(t *testing.T) {
	path := integrationModel(t)
	e := newEngine(Settings{GpuLayers: 999})
	lm := loadIntegrationModel(t, e, path)

	opts := completion.Options{
		System: "You are a terse assistant. Answer in a single short sentence.",
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{
				completion.Text{Text: "What is the capital of France? Answer with just the city name."},
			}},
		},
		MaxTokens: 64,
	}

	prompt, err := lm.adapter.renderPrompt(opts)
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	t.Logf("prompt:\n%s", prompt)

	nCtx := e.effectiveCtx(catalogEntry{}, lm.meta)
	out, err := e.generate(lm, prompt, lm.adapter.stopStrings(), lm.adapter.toolMarkers(), opts.MaxTokens, nCtx, nil)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	contents, parsed := lm.adapter.parse(out.text)
	t.Logf("raw output:   %q", out.text)
	t.Logf("stop=%v parsed=%v promptTokens=%d outputTokens=%d nCtx=%d",
		out.stop, parsed, out.promptTokens, out.outputTokens, nCtx)
	for _, c := range contents {
		t.Logf("content: %T %+v", c, c)
	}

	if strings.TrimSpace(out.text) == "" {
		t.Fatalf("expected non-empty output")
	}
	if !strings.Contains(strings.ToLower(out.text), "paris") {
		t.Errorf("expected the answer to mention Paris, got %q", out.text)
	}
	if out.stop == completion.StopMaxTokens {
		t.Errorf("generation hit the token cap instead of ending the turn naturally (output may be truncated)")
	}
}

// TestIntegrationStreaming verifies the streaming emit path surfaces the same text incrementally and that the
// concatenated deltas equal the final text.
func TestIntegrationStreaming(t *testing.T) {
	path := integrationModel(t)
	e := newEngine(Settings{GpuLayers: 999})
	lm := loadIntegrationModel(t, e, path)

	opts := completion.Options{
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{
				completion.Text{Text: "Count from 1 to 5, separated by commas."},
			}},
		},
		MaxTokens: 64,
	}

	prompt, err := lm.adapter.renderPrompt(opts)
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}

	var streamed strings.Builder
	emit := func(s string) bool {
		streamed.WriteString(s)
		return true
	}

	nCtx := e.effectiveCtx(catalogEntry{}, lm.meta)
	out, err := e.generate(lm, prompt, lm.adapter.stopStrings(), lm.adapter.toolMarkers(), opts.MaxTokens, nCtx, emit)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	t.Logf("streamed=%q final=%q", streamed.String(), out.text)
	if streamed.String() != out.text {
		t.Errorf("streamed text %q != final text %q", streamed.String(), out.text)
	}
	if strings.TrimSpace(out.text) == "" {
		t.Fatalf("expected non-empty output")
	}
}

// TestIntegrationToolCall verifies the tool-call round trip: with a single tool advertised and a prompt that
// requires it, the model must emit a tool call that the adapter parses into a [completion.ToolCall] with the
// expected name and valid JSON arguments.
func TestIntegrationToolCall(t *testing.T) {
	path := integrationModel(t)
	e := newEngine(Settings{GpuLayers: 999})
	lm := loadIntegrationModel(t, e, path)

	opts := completion.Options{
		System: "You are a helpful assistant with access to tools. When a tool is relevant, call it.",
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{
				completion.Text{Text: "What is the weather in Paris right now? Use the get_weather tool."},
			}},
		},
		Tools: []completion.ToolDef{
			{
				Name:        "get_weather",
				Description: "Get the current weather for a given city.",
				Schema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"city": {"type": "string", "description": "The city name"}
					},
					"required": ["city"]
				}`),
			},
		},
		MaxTokens: 256,
	}

	prompt, err := lm.adapter.renderPrompt(opts)
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	t.Logf("prompt:\n%s", prompt)

	nCtx := e.effectiveCtx(catalogEntry{}, lm.meta)
	out, err := e.generate(lm, prompt, lm.adapter.stopStrings(), lm.adapter.toolMarkers(), opts.MaxTokens, nCtx, nil)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	contents, parsed := lm.adapter.parse(out.text)
	t.Logf("raw output: %q", out.text)
	t.Logf("stop=%v parsed=%v", out.stop, parsed)

	var calls []completion.ToolCall
	for _, c := range contents {
		t.Logf("content: %T %+v", c, c)
		if tc, ok := c.(completion.ToolCall); ok {
			calls = append(calls, tc)
		}
	}

	if len(calls) == 0 {
		t.Fatalf("expected at least one tool call, got none; raw output: %q", out.text)
	}
	tc := calls[0]
	if tc.Name != "get_weather" {
		t.Errorf("tool call name = %q, want get_weather", tc.Name)
	}
	if !json.Valid(tc.Arguments) {
		t.Errorf("tool call arguments are not valid JSON: %q", tc.Arguments)
	}
	if tc.ID == "" {
		t.Errorf("tool call id must not be empty")
	}
	if parsed != completion.StopToolUse {
		t.Errorf("parsed stop reason = %v, want StopToolUse", parsed)
	}
}
