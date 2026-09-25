// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

type addIn struct {
	A int `json:"a" desc:"first summand"`
	B int `json:"b"`
}

type addOut struct {
	Sum int `json:"sum"`
}

func TestNewTool_Schema(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	if tool.Def.Name != "add" {
		t.Fatalf("unexpected name: %q", tool.Def.Name)
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
		t.Fatalf("schema not valid json: %v", err)
	}

	if schema["type"] != "object" {
		t.Fatalf("expected object schema, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("missing properties: %v", schema)
	}

	a, ok := props["a"].(map[string]any)
	if !ok || a["type"] != "integer" || a["description"] != "first summand" {
		t.Fatalf("unexpected schema for a: %v", props["a"])
	}

	required, ok := schema["required"].([]any)
	if !ok || len(required) != 2 {
		t.Fatalf("expected both fields required, got %v", schema["required"])
	}
}

func TestNewTool_Invoke(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	out, err := tool.Invoke(nil, json.RawMessage(`{"a":2,"b":40}`))
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}

	if string(out) != `{"sum":42}` {
		t.Fatalf("unexpected result: %s", out)
	}
}

// fakeCompletions returns the queued results in order, so we can simulate a tool-use turn followed by a
// final answer.
type fakeCompletions struct {
	results []Result
	calls   int
	// reqs records every request, so tests can inspect e.g. the escalated MaxTokens.
	reqs []Options
}

func (f *fakeCompletions) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

// Complete returns the queued results in order and repeats the last one once the queue is exhausted.
func (f *fakeCompletions) Complete(_ auth.Subject, opts Options) (Result, error) {
	f.reqs = append(f.reqs, opts)
	r := f.results[min(f.calls, len(f.results)-1)]
	f.calls++
	return r, nil
}

func (f *fakeCompletions) Stream(auth.Subject, Options) iter.Seq2[Delta, error] { return nil }

func TestRun_ExecutesToolLoop(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":2,"b":40}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message: Message{Role: Assistant, Content: []Content{
					Text{Text: "the sum is 42"},
				}},
				StopReason: StopEndTurn,
			},
		},
	}

	res, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "add 2 and 40"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.StopReason != StopEndTurn {
		t.Fatalf("unexpected stop reason: %v", res.StopReason)
	}

	// user prompt + assistant tool_use + user tool_result + assistant final = 4 messages
	if len(history) != 4 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}

	toolResultMsg := history[2]
	if toolResultMsg.Role != User || len(toolResultMsg.Content) != 1 {
		t.Fatalf("expected tool result user message, got %+v", toolResultMsg)
	}

	tr, ok := toolResultMsg.Content[0].(ToolResult)
	if !ok || tr.ToolCallID != "1" || tr.IsError {
		t.Fatalf("unexpected tool result: %+v", toolResultMsg.Content[0])
	}
}

// hasDanglingToolUse reports whether any assistant message contains a ToolCall that is not answered by a
// ToolResult with the same ID in the immediately following message. This is exactly the condition the
// Anthropic API rejects with a 400.
func hasDanglingToolUse(history []Message) bool {
	for i, m := range history {
		var ids []string
		for _, c := range m.Content {
			if call, ok := c.(ToolCall); ok {
				ids = append(ids, call.ID)
			}
		}
		if len(ids) == 0 {
			continue
		}

		answered := map[string]bool{}
		if i+1 < len(history) {
			for _, c := range history[i+1].Content {
				if tr, ok := c.(ToolResult); ok {
					answered[tr.ToolCallID] = true
				}
			}
		}
		for _, id := range ids {
			if !answered[id] {
				return true
			}
		}
	}
	return false
}

// TestRun_DropsTruncatedToolUse reproduces the 400 from the bug report: the model emitted tool_use blocks
// but the turn was cut off (stop_reason == max_tokens). The loop must not persist a dangling tool_use.
func TestRun_DropsTruncatedToolUse(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					Text{Text: "let me compute"},
					ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":2,"b":40}`)},
					ToolCall{ID: "2", Name: "add", Arguments: json.RawMessage(`{"a":1,`)}, // truncated args
				}},
				StopReason: StopMaxTokens,
			},
		},
	}

	res, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "add some numbers"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.StopReason != StopMaxTokens {
		t.Fatalf("unexpected stop reason: %v", res.StopReason)
	}

	if hasDanglingToolUse(history) {
		t.Fatalf("history contains a tool_use without a matching tool_result: %+v", history)
	}

	// initial attempt + maxTokenRetries escalations with a growing budget
	if fake.calls != 1+maxTokenRetries {
		t.Fatalf("expected %d attempts, got %d", 1+maxTokenRetries, fake.calls)
	}
	if fake.reqs[1].MaxTokens <= fake.reqs[0].MaxTokens || fake.reqs[2].MaxTokens <= fake.reqs[1].MaxTokens {
		t.Fatalf("expected escalating budgets, got %d, %d, %d", fake.reqs[0].MaxTokens, fake.reqs[1].MaxTokens, fake.reqs[2].MaxTokens)
	}

	// user prompt + assistant text (tool_use stripped) = 2 messages; no tool was executed.
	if len(history) != 2 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}

	last := history[1]
	if last.Role != Assistant || len(last.Content) != 1 {
		t.Fatalf("expected single-block assistant message, got %+v", last)
	}
	if _, ok := last.Content[0].(Text); !ok {
		t.Fatalf("expected remaining text block, got %T", last.Content[0])
	}
}

// TestRun_TruncatedToolUseOnlyDropsMessage verifies that an aborted turn consisting solely of tool_use
// blocks (no text/thinking) is dropped entirely rather than persisted as an empty assistant message.
func TestRun_TruncatedToolUseOnlyDropsMessage(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":2,`)},
				}},
				StopReason: StopMaxTokens,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "add"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if hasDanglingToolUse(history) {
		t.Fatalf("history contains a dangling tool_use: %+v", history)
	}

	// Only the original user prompt remains.
	if len(history) != 1 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}
}

type openIn struct {
	FID string `json:"fid" desc:"the id of the file to open"`
}

// TestNewOpenFileTool_Schema verifies the file-providing tool advertises OpenFile (not Invoke) and derives its
// input schema from In just like NewTool.
func TestNewOpenFileTool_Schema(t *testing.T) {
	tool := NewOpenFileTool("open_file", "opens a file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{}, nil
	})

	if tool.OpenFile == nil {
		t.Fatal("expected OpenFile to be set")
	}
	if tool.Invoke != nil {
		t.Fatal("expected Invoke to be nil for a file tool")
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
		t.Fatalf("schema not valid json: %v", err)
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok || props["fid"] == nil {
		t.Fatalf("expected fid property in schema, got %v", schema)
	}
}

// TestRun_OpenFileTool_AttachesMediaOnUserTurn is the core test: an OpenFile tool must upload its file and add
// a Media block to the SAME user turn as the tool_result, and the tool_result itself must NOT contain the
// media (a file-id source is invalid inside a tool_result).
func TestRun_OpenFileTool_AttachesMediaOnUserTurn(t *testing.T) {
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{
			Name:     "report.pdf",
			MimeType: file.PDF,
			Open:     func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader([]byte("%PDF-1.7"))), nil },
		}, nil
	})

	var uploadedName string
	uploader := func(_ auth.Subject, f OpenedFile) (file.ID, error) {
		uploadedName = f.Name
		// drain the reader to prove it is readable
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		_, _ = io.Copy(io.Discard, rc)
		return file.ID("file-xyz"), nil
	}

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "the pdf says hello"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "look at file abc"}}}},
		},
		Tools:        []Tool{tool},
		FileUploader: uploader,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if uploadedName != "report.pdf" {
		t.Fatalf("expected uploader to receive the file, got %q", uploadedName)
	}

	// history: user prompt + assistant tool_use + user(tool_result + media) + assistant final = 4
	if len(history) != 4 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}

	userTurn := history[2]
	if userTurn.Role != User {
		t.Fatalf("expected user turn at index 2, got %s", userTurn.Role)
	}
	if len(userTurn.Content) != 2 {
		t.Fatalf("expected tool_result + media on the user turn, got %d blocks: %#v", len(userTurn.Content), userTurn.Content)
	}

	tr, ok := userTurn.Content[0].(ToolResult)
	if !ok || tr.ToolCallID != "1" || tr.IsError {
		t.Fatalf("expected a successful tool_result first, got %#v", userTurn.Content[0])
	}
	// The tool_result must NOT carry the media itself.
	for _, c := range tr.Content {
		if _, isMedia := c.(Media); isMedia {
			t.Fatalf("tool_result must not contain a Media block: %#v", tr.Content)
		}
	}

	media, ok := userTurn.Content[1].(Media)
	if !ok {
		t.Fatalf("expected a Media block after the tool_result, got %T", userTurn.Content[1])
	}
	if media.MimeType != file.PDF || media.Source.FileID.UnwrapOr("") != file.ID("file-xyz") {
		t.Fatalf("unexpected media: %#v", media)
	}
}

// TestRun_OpenFileTool_NoUploaderIsError verifies that without a FileUploader the OpenFile call is reported as
// an error tool_result and no media is attached.
func TestRun_OpenFileTool_NoUploaderIsError(t *testing.T) {
	opened := false
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		opened = true
		return OpenedFile{Name: "x.pdf", MimeType: file.PDF, Open: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("data")), nil
		}}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "done"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "open abc"}}}}},
		Tools:   []Tool{tool}, // no FileUploader
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	_ = opened // the tool may or may not be opened before the uploader check; not asserted

	userTurn := history[2]
	if len(userTurn.Content) != 1 {
		t.Fatalf("expected only the error tool_result, got %#v", userTurn.Content)
	}
	tr, ok := userTurn.Content[0].(ToolResult)
	if !ok || !tr.IsError {
		t.Fatalf("expected an error tool_result, got %#v", userTurn.Content[0])
	}
}

// TestRun_OpenFileTool_ToolErrorIsReported verifies that an error from the OpenFile function is reported as an
// error tool_result and no upload happens.
func TestRun_OpenFileTool_ToolErrorIsReported(t *testing.T) {
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{}, io.ErrUnexpectedEOF
	})

	uploaderCalled := false
	uploader := func(_ auth.Subject, f OpenedFile) (file.ID, error) {
		uploaderCalled = true
		return "", nil
	}

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "done"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options:      Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "open abc"}}}}},
		Tools:        []Tool{tool},
		FileUploader: uploader,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if uploaderCalled {
		t.Fatal("uploader must not be called when the tool itself errors")
	}

	tr, ok := history[2].Content[0].(ToolResult)
	if !ok || !tr.IsError {
		t.Fatalf("expected an error tool_result, got %#v", history[2].Content[0])
	}
}

// TestRun_OpenFileTool_TextInjectedInline verifies that a text file is injected inline as the tool_result text
// (no upload, no Media block) and works even without a FileUploader configured.
func TestRun_OpenFileTool_TextInjectedInline(t *testing.T) {
	const content = "# Requirements\n\nthe system must do X and Y"
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{
			Name:     "anforderungen.md",
			MimeType: file.Markdown,
			Open:     func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(content)), nil },
		}, nil
	})

	uploaderCalled := false
	uploader := func(_ auth.Subject, _ OpenedFile) (file.ID, error) {
		uploaderCalled = true
		return "", nil
	}

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "summary"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options:      Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "read abc"}}}}},
		Tools:        []Tool{tool},
		FileUploader: uploader,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if uploaderCalled {
		t.Fatal("uploader must not be called for a text file")
	}

	userTurn := history[2]
	// only the tool_result must be present; no Media attachment
	if len(userTurn.Content) != 1 {
		t.Fatalf("expected only the tool_result (no media), got %#v", userTurn.Content)
	}
	tr, ok := userTurn.Content[0].(ToolResult)
	if !ok || tr.IsError {
		t.Fatalf("expected a successful tool_result, got %#v", userTurn.Content[0])
	}
	if len(tr.Content) != 1 {
		t.Fatalf("expected a single text block in the tool_result, got %#v", tr.Content)
	}
	txt, ok := tr.Content[0].(Text)
	if !ok {
		t.Fatalf("expected a Text block, got %T", tr.Content[0])
	}
	if !strings.Contains(txt.Text, content) {
		t.Fatalf("tool_result text does not contain the file content: %q", txt.Text)
	}
	if !strings.Contains(txt.Text, "anforderungen.md") {
		t.Fatalf("expected the file name in the injected text header: %q", txt.Text)
	}
	if strings.Contains(txt.Text, "truncated") {
		t.Fatalf("small file must not be reported as truncated: %q", txt.Text)
	}
}

// TestRun_OpenFileTool_TextWithoutUploader verifies a text file is injected even when no FileUploader is set.
func TestRun_OpenFileTool_TextWithoutUploader(t *testing.T) {
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{
			Name:     "notes.txt",
			MimeType: file.Text,
			Open:     func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("hello world")), nil },
		}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "ok"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "read abc"}}}}},
		Tools:   []Tool{tool}, // no FileUploader
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	tr, ok := history[2].Content[0].(ToolResult)
	if !ok || tr.IsError {
		t.Fatalf("expected a successful tool_result even without an uploader, got %#v", history[2].Content[0])
	}
	txt := tr.Content[0].(Text)
	if !strings.Contains(txt.Text, "hello world") {
		t.Fatalf("expected file content in tool_result, got %q", txt.Text)
	}
}

// TestExecuteOpenFileCall_TextTruncated verifies that a text file larger than the inline cap is truncated and
// carries a truncation note.
func TestExecuteOpenFileCall_TextTruncated(t *testing.T) {
	big := strings.Repeat("a", maxInlineTextBytes+1000)
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{
			Name:     "big.txt",
			MimeType: file.Text,
			Open:     func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(big)), nil },
		}, nil
	})

	call := ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)}
	tr, media := executeOpenFileCall(nil, call, tool, nil)

	if media != nil {
		t.Fatalf("text injection must not produce media, got %#v", media)
	}
	if tr.IsError {
		t.Fatalf("unexpected error tool_result: %#v", tr)
	}
	txt := tr.Content[0].(Text)
	if !strings.Contains(txt.Text, "truncated") {
		t.Fatalf("expected a truncation note, got tail: %q", txt.Text[len(txt.Text)-80:])
	}
	// the injected text is "<header>\n<body><note>"; the body is the capped run of 'a's. Verify the longest
	// run of 'a' equals exactly the cap (header/note contain only isolated 'a's).
	longestRun := 0
	run := 0
	for _, r := range txt.Text {
		if r == 'a' {
			run++
			if run > longestRun {
				longestRun = run
			}
		} else {
			run = 0
		}
	}
	if longestRun != maxInlineTextBytes {
		t.Fatalf("expected exactly %d body bytes, got a longest run of %d", maxInlineTextBytes, longestRun)
	}
}

// TestExecuteOpenFileCall_ImageStillUploads is a regression guard: a binary (image) file must still go through
// the uploader and produce a Media block.
func TestExecuteOpenFileCall_ImageStillUploads(t *testing.T) {
	tool := NewOpenFileTool("open_drive_file", "opens a drive file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{
			Name:     "pic.png",
			MimeType: file.PNG,
			Open:     func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("\x89PNG")), nil },
		}, nil
	})

	uploaderCalled := false
	uploader := func(_ auth.Subject, _ OpenedFile) (file.ID, error) {
		uploaderCalled = true
		return file.ID("file-png"), nil
	}

	call := ToolCall{ID: "1", Name: "open_drive_file", Arguments: json.RawMessage(`{"fid":"abc"}`)}
	tr, media := executeOpenFileCall(nil, call, tool, uploader)

	if !uploaderCalled {
		t.Fatal("expected the uploader to be called for an image")
	}
	if tr.IsError {
		t.Fatalf("unexpected error tool_result: %#v", tr)
	}
	if len(media) != 1 {
		t.Fatalf("expected exactly one media block, got %#v", media)
	}
	m, ok := media[0].(Media)
	if !ok || m.MimeType != file.PNG || m.Source.FileID.UnwrapOr("") != file.ID("file-png") {
		t.Fatalf("unexpected media block: %#v", media[0])
	}
}

func userMsg(text string) []Message {
	return []Message{{Role: User, Content: []Content{Text{Text: text}}}}
}

func assistantText(stop StopReason, blocks ...Content) Result {
	return Result{Message: Message{Role: Assistant, Content: blocks}, StopReason: stop}
}

func TestRun_PauseTurnResumes(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopPauseTurn, Text{Text: "working"}),
		assistantText(StopEndTurn, Text{Text: "done"}),
	}}

	res, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("go")}})
	if err != nil {
		t.Fatal(err)
	}
	if fake.calls != 2 || res.StopReason != StopEndTurn {
		t.Fatalf("expected the paused turn to be resumed, calls=%d stop=%s", fake.calls, res.StopReason)
	}
	// user, paused assistant turn (sent back as-is), final answer
	if len(history) != 3 || history[1].Role != Assistant {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestRun_MaxTokensRetriesWithLargerBudget(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopMaxTokens, Thinking{Text: "long reasoning"}),
		assistantText(StopEndTurn, Text{Text: "answer"}),
	}}

	res, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("q"), MaxTokens: 4096}})
	if err != nil {
		t.Fatal(err)
	}
	if res.StopReason != StopEndTurn {
		t.Fatalf("unexpected stop reason %s", res.StopReason)
	}
	if got := fake.reqs[1].MaxTokens; got != 8192 {
		t.Fatalf("expected retry with doubled budget, got %d", got)
	}
	// the truncated attempt is discarded
	if len(history) != 2 {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestRun_MaxTokensAtCapContinuesText(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopMaxTokens, Text{Text: "part 1"}),
		assistantText(StopEndTurn, Text{Text: "part 2"}),
	}}

	_, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("q"), MaxTokens: maxEscalatedTokens}})
	if err != nil {
		t.Fatal(err)
	}
	// user, part 1, injected continue prompt, part 2
	if len(history) != 4 || !IsLoopPrompt(history[2]) {
		t.Fatalf("unexpected history: %+v", history)
	}
	if IsLoopPrompt(history[0]) {
		t.Fatalf("a real user message must not count as loop prompt")
	}
}

func TestRun_ThinkingOnlyAnswerIsRequested(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopEndTurn, Thinking{Text: "hmm", Signature: "s"}),
		assistantText(StopEndTurn, Text{Text: "answer"}),
	}}

	_, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("q")}})
	if err != nil {
		t.Fatal(err)
	}
	// user, thinking-only turn, injected answer prompt, answer
	if len(history) != 4 || !IsLoopPrompt(history[2]) {
		t.Fatalf("unexpected history: %+v", history)
	}
	if _, ok := history[3].Content[0].(Text); !ok {
		t.Fatalf("expected a final text answer, got %+v", history[3])
	}
}

func TestRun_NeverPersistsEmptyMessages(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopEndTurn, Text{Text: " "}),
	}}

	_, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("q")}})
	if err != nil {
		t.Fatal(err)
	}
	if fake.calls != 1+maxContinuations {
		t.Fatalf("expected %d attempts, got %d", 1+maxContinuations, fake.calls)
	}
	for _, m := range history {
		if m.Role == Assistant && !hasContent(m) {
			t.Fatalf("history contains an empty assistant message: %+v", history)
		}
	}
}

func TestRun_TurnLimitKeepsValidHistory(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})
	fake := &fakeCompletions{results: []Result{
		assistantText(StopToolUse, ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":1,"b":2}`)}),
	}}

	_, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("q")}, Tools: []Tool{tool}, MaxTurns: 3})
	if !errors.Is(err, TurnLimitExceeded) {
		t.Fatalf("expected TurnLimitExceeded, got %v", err)
	}
	// user + 3 x (tool_use, tool_result)
	if len(history) != 7 || hasDanglingToolUse(history) {
		t.Fatalf("unexpected history (%d): %+v", len(history), history)
	}
}

func TestRun_DefaultTurnLimit(t *testing.T) {
	if DefaultMaxToolTurns != 128 {
		t.Fatalf("unexpected default turn limit %d", DefaultMaxToolTurns)
	}
}

func TestMessageJSON_RedactedThinkingRoundTrip(t *testing.T) {
	in := Message{Role: Assistant, Content: []Content{RedactedThinking{Data: "enc"}, Text{Text: "hi"}}}
	buf, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out Message
	if err := json.Unmarshal(buf, &out); err != nil {
		t.Fatal(err)
	}
	if rt, ok := out.Content[0].(RedactedThinking); !ok || rt.Data != "enc" {
		t.Fatalf("redacted thinking lost: %+v", out.Content)
	}
}
