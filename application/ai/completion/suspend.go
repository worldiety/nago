// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.wdy.de/nago/auth"
)

// AskUserToolName is the name of the built-in clarification tool created by [NewAskUserTool].
const AskUserToolName = "ask_user"

// ErrSuspendUnsupported is returned (wrapped) by [Run] when the model hit a user decision (see
// [Tool.AwaitsUser], [RunOptions.ConfirmMutating]). Use [Start] and [Continue] to support suspension.
var ErrSuspendUnsupported = errors.New("run suspension not supported by caller")

// ErrContinuationMismatch is returned (wrapped) by [Continue] when the continuation or the resolutions do not
// fit the supplied history, e.g. because a pending call was answered twice or not at all.
var ErrContinuationMismatch = errors.New("continuation does not match history")

// deferredText is the tool result of a mutating call the model requested in the same turn as a clarifying
// question. It is not executed, because the answer may change what should happen.
const deferredText = "not executed: you asked the user a question in the same turn. Wait for the answer and request this call again if it is still needed."

// dismissedText is the tool result of a pending call the user dismissed without deciding.
const dismissedText = "the user dismissed this without answering and moved on to something else. Do not assume an answer; the call was not executed."

// PendingKind classifies a [PendingCall].
type PendingKind string

const (
	// PendingQuestion is a clarifying question of a [Tool.AwaitsUser] tool, answered with free text.
	PendingQuestion PendingKind = "question"
	// PendingApproval is a mutating call held for the user's approval (see [RunOptions.ConfirmMutating]).
	PendingApproval PendingKind = "approval"
)

// PendingCall is a tool call a suspended run waits on.
type PendingCall struct {
	Kind PendingKind `json:"kind"`
	// Call is the exact call of the model. An approved call is executed with exactly these arguments.
	Call ToolCall `json:"call"`
	// Question and Options are set for [PendingQuestion].
	Question string   `json:"question,omitempty"`
	Options  []string `json:"options,omitempty"`
	// Effect is the human-readable description of the change for [PendingApproval] (see [Tool.Confirm]).
	Effect string `json:"effect,omitempty"`
}

// Continuation is the resumable state of a suspended run. It belongs to the history returned together with
// it, which ends with the assistant turn carrying the pending calls. It is JSON-serializable, so it can be
// persisted and resumed after a restart; see [Continue].
type Continuation struct {
	// Pending are the calls waiting on a user decision.
	Pending []PendingCall
	// Completed are the results of the calls of the same turn which already ran. They are never executed
	// again on resume.
	Completed []ToolResult
	// Attachments are media blocks produced by already executed file tools of the same turn.
	Attachments []Content
}

// Resolution is the user's decision on one [PendingCall].
type Resolution struct {
	// CallID identifies the pending call ([ToolCall.ID]).
	CallID string
	// Answer is the free-text answer to a [PendingQuestion].
	Answer string
	// Approved executes a [PendingApproval] call. When false, the call is rejected.
	Approved bool
	// Dismissed resolves the call without an answer or execution, e.g. because the user moved on.
	Dismissed bool
	// Message is an optional explanation sent to the model for a rejected approval.
	Message string
}

// Outcome is the result of [Start] and [Continue].
type Outcome struct {
	// Result is the last model answer.
	Result Result
	// History is the full trace, including all tool calls and results. It is valid to persist at any time.
	History []Message
	// Suspended is set when the run waits on a user decision. Persist it together with History and resume
	// via [Continue].
	Suspended *Continuation
	// Progressed reports whether History moved beyond the input (model turns, tool results, a resumed
	// decision or a compaction). A failed run with Progressed set left a history worth persisting.
	Progressed bool
}

type resumeInput struct {
	cont        Continuation
	resolutions []Resolution
}

// NewAskUserTool creates the clarification tool. It never runs a Go function: calling it suspends the run
// (see [Start]) until the user answered via [Continue]. The answer reaches the model as {"answer": "..."}.
func NewAskUserTool() Tool {
	schema := json.RawMessage(`{"type":"object","properties":{"question":{"type":"string","description":"the clarifying question to ask the user"},"options":{"type":"array","items":{"type":"string"},"description":"optional predefined answers the user may pick from"}},"required":["question"]}`)
	return Tool{
		Def: ToolDef{
			Name:        AskUserToolName,
			Description: "asks the user a clarifying question and waits for their answer before continuing. Use this whenever you need a decision or missing information from the user.",
			Schema:      schema,
		},
		AwaitsUser: true,
	}
}

// questionOf extracts a pending question from an ask call.
func questionOf(call ToolCall) (PendingCall, bool) {
	var in struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
	}
	if err := json.Unmarshal(call.Arguments, &in); err != nil || strings.TrimSpace(in.Question) == "" {
		return PendingCall{}, false
	}
	return PendingCall{Kind: PendingQuestion, Call: call, Question: in.Question, Options: in.Options}, true
}

// answerResult encodes the answer the way [NewAskUserTool] promises it to the model.
func answerResult(callID, answer string) ToolResult {
	buf, _ := json.Marshal(struct {
		Answer string `json:"answer"`
	}{answer})
	return ToolResult{ToolCallID: callID, Content: []Content{Text{Text: string(buf)}}}
}

// resumeMessage builds the user turn which answers every tool call of the last assistant message: the
// already completed results plus the resolved pending calls, in call order, followed by attachments.
func resumeMessage(subject auth.Subject, tools map[string]Tool, opts RunOptions, history []Message, cont Continuation, resolutions []Resolution) (Message, error) {
	if len(history) == 0 || history[len(history)-1].Role != Assistant {
		return Message{}, fmt.Errorf("history does not end with the suspended assistant turn: %w", ErrContinuationMismatch)
	}

	byID := map[string]Resolution{}
	for _, r := range resolutions {
		if _, dup := byID[r.CallID]; dup {
			return Message{}, fmt.Errorf("call %q resolved twice: %w", r.CallID, ErrContinuationMismatch)
		}
		byID[r.CallID] = r
	}

	pending := map[string]PendingCall{}
	for _, pc := range cont.Pending {
		r, ok := byID[pc.Call.ID]
		if !ok {
			return Message{}, fmt.Errorf("pending call %q not resolved: %w", pc.Call.ID, ErrContinuationMismatch)
		}
		if pc.Kind == PendingQuestion && !r.Dismissed && strings.TrimSpace(r.Answer) == "" {
			return Message{}, fmt.Errorf("empty answer for question %q", pc.Call.ID)
		}
		pending[pc.Call.ID] = pc
	}
	if len(byID) != len(pending) {
		return Message{}, fmt.Errorf("resolution for an unknown call: %w", ErrContinuationMismatch)
	}

	completed := map[string]ToolResult{}
	for _, r := range cont.Completed {
		completed[r.ToolCallID] = r
	}

	var results, attachments []Content
	attachments = append(attachments, cont.Attachments...)
	for _, c := range history[len(history)-1].Content {
		call, ok := c.(ToolCall)
		if !ok {
			continue
		}

		if r, ok := completed[call.ID]; ok {
			results = append(results, r)
			continue
		}

		pc, ok := pending[call.ID]
		if !ok {
			return Message{}, fmt.Errorf("call %q has neither a result nor a pending decision: %w", call.ID, ErrContinuationMismatch)
		}

		r := byID[call.ID]
		switch {
		case r.Dismissed:
			results = append(results, ToolResult{ToolCallID: call.ID, IsError: true, Content: []Content{Text{Text: dismissedText}}})
		case pc.Kind == PendingQuestion:
			results = append(results, answerResult(call.ID, r.Answer))
		case r.Approved:
			// Execute exactly the arguments the user approved, not whatever the model might send later.
			result, media := executeToolCall(subject, tools, pc.Call, opts.FileUploader, opts.OnBeforeToolCall)
			results = append(results, result)
			attachments = append(attachments, media...)
		default:
			msg := r.Message
			if msg == "" {
				msg = "the user rejected this call; it was not executed"
			}
			results = append(results, ToolResult{ToolCallID: call.ID, IsError: true, Content: []Content{Text{Text: msg}}})
		}
	}

	if len(results) != len(completed)+len(pending) {
		return Message{}, fmt.Errorf("continuation holds results for calls outside the suspended turn: %w", ErrContinuationMismatch)
	}

	return Message{Role: User, Content: append(results, attachments...)}, nil
}

// Dismiss closes a suspended run without asking the model again: every pending call is answered as dismissed,
// already completed results are kept. The returned history is valid, so a new, unrelated user turn may follow.
func Dismiss(history []Message, cont Continuation) ([]Message, error) {
	resolutions := make([]Resolution, 0, len(cont.Pending))
	for _, pc := range cont.Pending {
		resolutions = append(resolutions, Resolution{CallID: pc.Call.ID, Dismissed: true})
	}

	msg, err := resumeMessage(nil, nil, RunOptions{}, history, cont, resolutions)
	if err != nil {
		return nil, err
	}

	out := make([]Message, 0, len(history)+1)
	out = append(out, history...)
	return append(out, msg), nil
}

// continuationVersion is the persisted format version of [Continuation].
const continuationVersion = 1

type continuationWire struct {
	Version     int               `json:"version"`
	Pending     []PendingCall     `json:"pending,omitempty"`
	Completed   []toolResultWire  `json:"completed,omitempty"`
	Attachments []contentEnvelope `json:"attachments,omitempty"`
}

// MarshalJSON encodes the continuation losslessly, including the heterogeneous attachment content.
func (c Continuation) MarshalJSON() ([]byte, error) {
	w := continuationWire{Version: continuationVersion, Pending: c.Pending}
	for _, r := range c.Completed {
		rw, err := toolResultToWire(r)
		if err != nil {
			return nil, err
		}
		w.Completed = append(w.Completed, rw)
	}

	att, err := marshalContentSlice(c.Attachments)
	if err != nil {
		return nil, err
	}
	w.Attachments = att

	return json.Marshal(w)
}

// UnmarshalJSON is the inverse of [Continuation.MarshalJSON].
func (c *Continuation) UnmarshalJSON(data []byte) error {
	var w continuationWire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	if w.Version != continuationVersion {
		return fmt.Errorf("unsupported continuation version %d", w.Version)
	}

	out := Continuation{Pending: w.Pending}
	for _, rw := range w.Completed {
		r, err := rw.toToolResult()
		if err != nil {
			return err
		}
		out.Completed = append(out.Completed, r)
	}

	att, err := unmarshalContentSlice(w.Attachments)
	if err != nil {
		return err
	}
	out.Attachments = att

	*c = out
	return nil
}
