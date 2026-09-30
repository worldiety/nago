// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build manual

package openai

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
)

// These are manual end-to-end tests against a real OpenAI-compatible server. They are excluded from a plain
// `go test ./...` by the "manual" build tag and skip themselves if the server is unreachable.
//
// Run them against a local Ollama in Docker:
//
//	docker run -d --name nago-openai-test -p 11434:11434 ollama/ollama
//	docker exec nago-openai-test ollama pull qwen2.5:0.5b
//	go test -tags manual -run TestManual -v ./application/ai/provider/openai/
//	docker rm -f nago-openai-test
//
// The target is configured by environment variables:
//
//	NAGO_OPENAI_BASE_URL  base URL of the API (default http://localhost:11434/v1)
//	NAGO_OPENAI_MODEL     model to use (default qwen2.5:0.5b)
//	NAGO_OPENAI_TOKEN     optional API token, e.g. to run against https://api.openai.com/v1

func manualProvider(t *testing.T) (*openaiProvider, model.ID) {
	t.Helper()

	baseURL := os.Getenv("NAGO_OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}

	m := os.Getenv("NAGO_OPENAI_MODEL")
	if m == "" {
		m = "qwen2.5:0.5b"
	}

	probe := &http.Client{Timeout: 3 * time.Second}
	resp, err := probe.Get(strings.TrimRight(baseURL, "/") + "/models")
	if err != nil {
		t.Skipf("OpenAI-compatible server at %s not reachable: %v", baseURL, err)
	}
	_ = resp.Body.Close()

	return newProvider("manual", Settings{BaseURL: baseURL, Token: os.Getenv("NAGO_OPENAI_TOKEN")}), model.ID(m)
}

func manualOpts(m model.ID, text string) completion.Options {
	return completion.Options{
		Model:       m,
		System:      "You are a terse assistant.",
		MaxTokens:   200,
		Temperature: option.Some(0.0),
		Messages:    []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: text}}}},
	}
}

func TestManualModels(t *testing.T) {
	p, m := manualProvider(t)

	found := false
	for mdl, err := range p.completions.Models(nil) {
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("model: %s (%s)", mdl.ID, mdl.Description)
		if mdl.ID == m {
			found = true
		}
	}

	if !found {
		t.Errorf("model %s not listed", m)
	}
}

func TestManualComplete(t *testing.T) {
	p, m := manualProvider(t)

	res, err := p.completions.Complete(context.Background(), nil, manualOpts(m, "What is the capital of France? Answer with one word."))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result: %+v", res)
	text := resultText(res)
	if !strings.Contains(strings.ToLower(text), "paris") {
		t.Errorf("unexpected answer %q", text)
	}
	if res.StopReason != completion.StopEndTurn {
		t.Errorf("unexpected stop reason %q", res.StopReason)
	}
	if res.Usage.InputTokens == 0 || res.Usage.OutputTokens == 0 {
		t.Errorf("expected usage, got %+v", res.Usage)
	}
}

func TestManualMaxTokens(t *testing.T) {
	p, m := manualProvider(t)

	opts := manualOpts(m, "Write a long story about a lighthouse keeper.")
	opts.MaxTokens = 5
	res, err := p.completions.Complete(context.Background(), nil, opts)
	if err != nil {
		t.Fatal(err)
	}

	if res.StopReason != completion.StopMaxTokens {
		t.Errorf("expected max tokens stop reason, got %q (%+v)", res.StopReason, res.Usage)
	}
}

func TestManualStream(t *testing.T) {
	p, m := manualProvider(t)

	var sb strings.Builder
	var done completion.Delta
	chunks := 0
	for d, err := range p.completions.Stream(context.Background(), nil, manualOpts(m, "Count from 1 to 5, separated by commas.")) {
		if err != nil {
			t.Fatal(err)
		}
		if d.TextDelta != "" {
			chunks++
		}
		sb.WriteString(d.TextDelta)
		if d.Done {
			done = d
		}
	}

	t.Logf("streamed %d chunks: %q, final: %+v", chunks, sb.String(), done)
	if chunks < 2 || !strings.Contains(sb.String(), "3") {
		t.Errorf("unexpected stream %q in %d chunks", sb.String(), chunks)
	}
	if !done.Done || done.StopReason != completion.StopEndTurn {
		t.Errorf("unexpected final delta %+v", done)
	}
	if done.Usage.IsNone() || done.Usage.Unwrap().OutputTokens == 0 {
		t.Errorf("expected usage in the stream, got %+v", done.Usage)
	}
}

type addIn struct {
	A int `json:"a" description:"first summand"`
	B int `json:"b" description:"second summand"`
}

type addOut struct {
	Sum int `json:"sum"`
}

func addTool(calls *int) completion.Tool {
	return completion.NewTool("add", "Adds two integers and returns their sum. Always use it for additions.", func(in addIn) (addOut, error) {
		*calls++
		return addOut{Sum: in.A + in.B}, nil
	})
}

func TestManualStreamToolCall(t *testing.T) {
	p, m := manualProvider(t)

	calls := 0
	opts := manualOpts(m, "Use the add tool to compute 1234 + 4321.")
	opts.Tools = []completion.ToolDef{addTool(&calls).Def}

	var got []completion.ToolCall
	var done completion.Delta
	for d, err := range p.completions.Stream(context.Background(), nil, opts) {
		if err != nil {
			t.Fatal(err)
		}
		if d.ToolCall.IsSome() {
			got = append(got, d.ToolCall.Unwrap())
		}
		if d.Done {
			done = d
		}
	}

	t.Logf("streamed tool calls: %+v, final: %+v", got, done)
	if len(got) == 0 || got[0].Name != "add" || got[0].ID == "" {
		t.Fatalf("expected a streamed add tool call, got %+v", got)
	}
	if done.StopReason != completion.StopToolUse {
		t.Errorf("expected tool use stop reason, got %q", done.StopReason)
	}
}

func TestManualToolRoundtrip(t *testing.T) {
	p, m := manualProvider(t)

	calls := 0
	out, err := completion.Start(nil, p.completions, completion.RunOptions{
		Options:  manualOpts(m, "Use the add tool to compute 1234 + 4321, then tell me the result."),
		Tools:    []completion.Tool{addTool(&calls)},
		MaxTurns: 4,
		Context:  context.Background(),
	})
	if err != nil {
		t.Fatal(err)
	}

	for i, msg := range out.History {
		t.Logf("history[%d] %s: %+v", i, msg.Role, msg.Content)
	}

	if calls == 0 {
		t.Fatalf("the model never called the tool")
	}
	text := resultText(out.Result)
	if !strings.Contains(strings.ReplaceAll(text, ",", ""), "5555") {
		t.Errorf("expected the sum in the answer, got %q", text)
	}
}

func TestManualCancel(t *testing.T) {
	p, m := manualProvider(t)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	opts := manualOpts(m, "Write a very long essay about the history of computing.")
	opts.MaxTokens = 4000
	start := time.Now()
	_, err := p.completions.Complete(ctx, nil, opts)
	if err == nil || ctx.Err() == nil {
		t.Fatalf("expected a cancelled request, got %v", err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("cancellation took %s", d)
	}
}

func resultText(res completion.Result) string {
	var sb strings.Builder
	for _, c := range res.Message.Content {
		if t, ok := c.(completion.Text); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String()
}
