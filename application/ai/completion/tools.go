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
	"reflect"
	"strings"
	"time"

	"go.wdy.de/nago/auth"
)

// ToolFunc is the canonical shape of a Go function that can be exposed as a callable tool to the model.
// Both the input In and the output Out must be JSON marshalable, because the model delivers the arguments
// as JSON and the (string) result is fed back to the model as a tool result.
//
// Typical usage:
//
//	type AddIn struct {
//		A int `json:"a" desc:"first summand"`
//		B int `json:"b" desc:"second summand"`
//	}
//	type AddOut struct {
//		Sum int `json:"sum"`
//	}
//
//	add := completion.NewTool("add", "adds two integers", func(in AddIn) (AddOut, error) {
//		return AddOut{Sum: in.A + in.B}, nil
//	})
type ToolFunc[In, Out any] func(In) (Out, error)

// Tool bundles the advertised [ToolDef] (name, description, JSON schema) with an executable invocation that
// knows how to unmarshal the JSON arguments, run the underlying Go function and marshal the result back.
//
// Create one with [NewTool]. Pass the resulting tools to [Run]/[RunStream], which drives the full agentic
// loop (call model -> execute requested tools -> feed results back -> repeat) until the model produces a
// final answer.
type Tool struct {
	// Def is the schema advertised to the provider via [Options.Tools].
	Def ToolDef

	// Invoke executes the underlying Go function for the given raw JSON arguments and returns the raw JSON
	// encoded result. The error is a transport/marshalling or business error; [Run] turns it into a tool
	// result flagged as error so the model may react to it.
	Invoke func(args json.RawMessage) (json.RawMessage, error)
}

// NewTool wraps an arbitrary Go function of the form func(In) (Out, error) into a callable [Tool].
//
// The JSON schema describing In is derived automatically from the Go type via reflection (struct fields use
// their json tag for the property name and an optional `desc`/`description` struct tag for the property
// description). In and Out must be JSON marshalable.
//
// name must be a stable, unique identifier (the model references it by name). description should explain to
// the model when and how to use the tool.
func NewTool[In, Out any](name, description string, fn ToolFunc[In, Out]) Tool {
	var zeroIn In
	schema := reflectSchema(reflect.TypeOf(&zeroIn).Elem())
	rawSchema, err := json.Marshal(schema)
	if err != nil {
		// A type whose schema cannot be marshalled is a programming error; encode it defensively as an
		// empty object so the tool stays usable.
		rawSchema = json.RawMessage(`{"type":"object"}`)
	}

	return Tool{
		Def: ToolDef{
			Name:        name,
			Description: description,
			Schema:      rawSchema,
		},
		Invoke: func(args json.RawMessage) (json.RawMessage, error) {
			var in In
			if len(args) > 0 {
				if err := json.Unmarshal(args, &in); err != nil {
					return nil, fmt.Errorf("cannot decode arguments for tool %q: %w", name, err)
				}
			}

			out, err := fn(in)
			if err != nil {
				return nil, err
			}

			raw, err := json.Marshal(out)
			if err != nil {
				return nil, fmt.Errorf("cannot encode result of tool %q: %w", name, err)
			}

			return raw, nil
		},
	}
}

// DefaultMaxToolTurns bounds the agentic loop in [Run] when [RunOptions.MaxTurns] is zero, protecting against
// models that keep requesting tools indefinitely.
const DefaultMaxToolTurns = 16

// DefaultMaxCompactions bounds how often [Run] may invoke the [Compactor] across the whole run when
// [RunOptions.MaxCompactions] is zero. Each compaction must strictly shrink the history, so a small budget is
// sufficient and guarantees the loop terminates.
const DefaultMaxCompactions = 4

// Compactor shrinks the conversation history so that a subsequent request fits into the model's context
// window. It is invoked by [Run] whenever a completion fails with [ContextWindowExceeded]. The returned
// history replaces the previous one and the failed turn is retried.
//
// A Compactor receives everything it needs to perform its own completion requests (e.g. to summarize older
// turns): the subject, the [Completions] capability and the in-flight [Options] (which carries the active
// model and system prompt). Implementations MUST return a history that is strictly smaller than the input
// (fewer runes), otherwise [Run] aborts to avoid an infinite loop.
//
// See [NewSummaryCompactor] for the default summarizing implementation.
type Compactor func(subject auth.Subject, c Completions, opts Options, history []Message) ([]Message, error)

// ProgressPhase classifies a [Progress] event emitted during the agentic loop in [Run].
type ProgressPhase string

const (
	// PhaseTurnStarted is emitted right before the model is asked to complete the current history.
	PhaseTurnStarted ProgressPhase = "turn_started"

	// PhaseModelResponded is emitted after the model returned an assistant turn. The [Progress.Result] field
	// carries the raw model answer (which may request tools).
	PhaseModelResponded ProgressPhase = "model_responded"

	// PhaseToolStarted is emitted just before a requested tool is executed. The [Progress.ToolCall] field
	// describes the call.
	PhaseToolStarted ProgressPhase = "tool_started"

	// PhaseToolCompleted is emitted after a tool finished executing. The [Progress.ToolCall] and
	// [Progress.ToolResult] fields describe the call and its outcome.
	PhaseToolCompleted ProgressPhase = "tool_completed"
)

// Progress is a single observability event handed to [RunOptions.OnProgress] while [Run] drives the agentic
// loop. It lets a caller surface what is happening (e.g. "calling tool X", "thinking ...") to a waiting user.
//
// Depending on Phase only a subset of the optional fields is populated:
//   - PhaseTurnStarted: Turn, MaxTurns
//   - PhaseModelResponded: Turn, MaxTurns, Result
//   - PhaseToolStarted: Turn, MaxTurns, ToolCall
//   - PhaseToolCompleted: Turn, MaxTurns, ToolCall, ToolResult
type Progress struct {
	// Phase identifies which step of the loop produced this event.
	Phase ProgressPhase

	// Turn is the zero-based index of the current loop iteration.
	Turn int

	// MaxTurns is the effective turn limit (see [RunOptions.MaxTurns]).
	MaxTurns int

	// Result is the model answer for PhaseModelResponded, nil otherwise.
	Result *Result

	// ToolCall is the tool invocation for PhaseToolStarted/PhaseToolCompleted, nil otherwise.
	ToolCall *ToolCall

	// ToolResult is the tool outcome for PhaseToolCompleted, nil otherwise.
	ToolResult *ToolResult
}

// ProgressFunc receives [Progress] events while [Run] executes. It must not block for long, as it is invoked
// synchronously inside the loop.
type ProgressFunc func(Progress)

// RunOptions configures the agentic loop executed by [Run]. It embeds the stateless [Options] and adds the
// executable [Tool]s plus a turn limit.
type RunOptions struct {
	// Options is the base stateless request. Its Tools field is overwritten with the schema of the supplied
	// Tools, so it does not need to be set by the caller.
	Options

	// Tools are the executable Go functions the model may call. Their [ToolDef]s are advertised to the
	// provider automatically.
	Tools []Tool

	// MaxTurns caps how many times the model may request (and we execute) tools before [Run] gives up. Zero
	// means [DefaultMaxToolTurns].
	MaxTurns int

	// OnProgress is an optional callback invoked synchronously for every [Progress] event of the loop. Use
	// it to keep a waiting user informed (e.g. show which tool is currently running). May be nil.
	OnProgress ProgressFunc

	// Compactor shrinks the history when a turn fails with [ContextWindowExceeded] so the request fits the
	// model's context window again. When nil, [Run] initializes and uses a default [NewSummaryCompactor] so
	// context overflows are recovered from out of the box.
	Compactor Compactor

	// MaxCompactions caps how often [Compactor] may run across the whole [Run]. Zero means
	// [DefaultMaxCompactions].
	MaxCompactions int
}

// Run drives the full agentic loop on top of [Completions.Complete]:
//
//  1. advertise the tool schemas and call the model,
//  2. if the model requested tool calls, execute the matching Go functions,
//  3. feed the results back as a follow-up user message,
//  4. repeat until the model returns a final (non tool_use) answer or the turn limit is hit.
//
// It returns the final assistant [Result] together with the complete message history (including all
// intermediate tool calls and tool results) so callers can inspect or persist the trace.
func Run(subject auth.Subject, c Completions, opts RunOptions) (Result, []Message, error) {
	maxTurns := opts.MaxTurns
	if maxTurns <= 0 {
		maxTurns = DefaultMaxToolTurns
	}

	maxCompactions := opts.MaxCompactions
	if maxCompactions <= 0 {
		maxCompactions = DefaultMaxCompactions
	}
	compactions := 0

	// Compaction is on by default: when the caller did not supply a strategy, fall back to the summarizing
	// compactor so context window overflows are recovered from automatically.
	compactor := opts.Compactor
	if compactor == nil {
		compactor = NewSummaryCompactor(SummaryCompactorConfig{})
	}

	// notify reports a progress event to the optional callback, guarding against a nil OnProgress.
	notify := func(p Progress) {
		if opts.OnProgress == nil {
			return
		}
		p.MaxTurns = maxTurns
		opts.OnProgress(p)
	}

	tools := make(map[string]Tool, len(opts.Tools))
	defs := make([]ToolDef, 0, len(opts.Tools))
	for _, t := range opts.Tools {
		tools[t.Def.Name] = t
		defs = append(defs, t.Def)
	}

	req := opts.Options
	req.Tools = defs

	// copy the initial history so we never mutate the caller's slice
	history := make([]Message, len(req.Messages))
	copy(history, req.Messages)

	for turn := 0; turn < maxTurns; turn++ {
		req.Messages = history

		notify(Progress{Phase: PhaseTurnStarted, Turn: turn})

		// complete the current turn, transparently compacting the history and retrying on a context window
		// overflow until either the request fits or the compaction budget is exhausted.
		var res Result
		for {
			var err error
			res, err = c.Complete(subject, req)
			if err == nil {
				break
			}

			if !errors.Is(err, ContextWindowExceeded) || compactions >= maxCompactions {
				return Result{}, history, err
			}

			before := runeLen(history)
			compacted, cerr := compactor(subject, c, req, history)
			if cerr != nil {
				return Result{}, history, fmt.Errorf("compaction failed: %w", cerr)
			}
			compactions++

			// A compactor must make progress; otherwise we would loop forever on the same overflow.
			if runeLen(compacted) >= before {
				return Result{}, history, fmt.Errorf("compaction did not shrink history (%d runes): %w", before, err)
			}

			history = compacted
			req.Messages = history
		}

		notify(Progress{Phase: PhaseModelResponded, Turn: turn, Result: &res})

		// Collect any tool calls in this assistant turn up front, decoupled from the stop reason.
		var calls []ToolCall
		for _, content := range res.Message.Content {
			if call, ok := content.(ToolCall); ok {
				calls = append(calls, call)
			}
		}

		if res.StopReason != StopToolUse {
			// The turn did not (cleanly) request tools. If it nonetheless carries tool_use blocks the
			// generation was cut off mid tool-call (e.g. stop_reason == max_tokens). Anthropic requires every
			// tool_use to be followed by a matching tool_result; a truncated call has invalid/partial
			// arguments and must not be executed. Drop those blocks so the persisted history stays valid.
			if len(calls) > 0 {
				cleaned := stripToolCalls(res.Message)
				if len(cleaned.Content) > 0 {
					history = append(history, cleaned)
				}
			} else {
				history = append(history, res.Message)
			}
			return res, history, nil
		}

		history = append(history, res.Message)

		if len(calls) == 0 {
			// The model signalled tool_use but emitted no actual call we understand; stop to avoid looping.
			return res, history, nil
		}

		results := make([]Content, 0, len(calls))
		for _, call := range calls {
			call := call
			notify(Progress{Phase: PhaseToolStarted, Turn: turn, ToolCall: &call})

			result := executeToolCall(tools, call)

			notify(Progress{Phase: PhaseToolCompleted, Turn: turn, ToolCall: &call, ToolResult: &result})

			results = append(results, result)
		}

		history = append(history, Message{Role: User, Content: results})
	}

	return Result{}, history, fmt.Errorf("tool loop exceeded %d turns", maxTurns)
}

// stripToolCalls returns a copy of msg with all [ToolCall] content blocks removed. It is used to discard
// truncated tool_use blocks from an aborted turn (e.g. stop_reason == max_tokens) so the resulting history
// never contains a tool_use without a matching tool_result.
func stripToolCalls(msg Message) Message {
	out := make([]Content, 0, len(msg.Content))
	for _, c := range msg.Content {
		if _, ok := c.(ToolCall); ok {
			continue
		}
		out = append(out, c)
	}
	return Message{Role: msg.Role, Content: out}
}

// executeToolCall runs a single tool call and wraps the outcome into a [ToolResult] content block. Unknown
// tools and execution errors are reported back to the model as error results instead of aborting the loop.
func executeToolCall(tools map[string]Tool, call ToolCall) ToolResult {
	tool, ok := tools[call.Name]
	if !ok {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: fmt.Sprintf("unknown tool %q", call.Name)}},
			IsError:    true,
		}
	}

	out, err := tool.Invoke(call.Arguments)
	if err != nil {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: err.Error()}},
			IsError:    true,
		}
	}

	return ToolResult{
		ToolCallID: call.ID,
		Content:    []Content{Text{Text: string(out)}},
	}
}

// reflectSchema builds a (subset of) JSON Schema object for the given Go type. It supports the JSON
// marshalable primitives, slices/arrays, maps, pointers and (possibly nested/embedded) structs. Struct
// fields honour their json tag for the property name and the omitempty option, plus an optional
// `desc`/`description` struct tag used as the property description.
func reflectSchema(t reflect.Type) map[string]any {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return map[string]any{}
	}

	switch t.Kind() {
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Slice, reflect.Array:
		// []byte is JSON-encoded as a base64 string
		if t.Elem().Kind() == reflect.Uint8 {
			return map[string]any{"type": "string"}
		}
		return map[string]any{"type": "array", "items": reflectSchema(t.Elem())}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": reflectSchema(t.Elem())}
	case reflect.Interface:
		// "any" – no constraint
		return map[string]any{}
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return map[string]any{"type": "string", "format": "date-time"}
		}

		properties := map[string]any{}
		var required []string
		collectStructFields(t, properties, &required)

		schema := map[string]any{
			"type":                 "object",
			"properties":           properties,
			"additionalProperties": false,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	default:
		return map[string]any{}
	}
}

// collectStructFields fills properties/required for a struct type, recursing into embedded (anonymous)
// structs so their fields are promoted just like encoding/json does.
func collectStructFields(t reflect.Type, properties map[string]any, required *[]string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous { // unexported, non-embedded
			continue
		}

		name, omitempty, skip := parseJSONField(f)
		if skip {
			continue
		}

		// promote embedded struct fields when they have no explicit json name
		if f.Anonymous && name == "" {
			ft := f.Type
			for ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				collectStructFields(ft, properties, required)
				continue
			}
		}

		if name == "" {
			name = f.Name
		}

		sub := reflectSchema(f.Type)
		if desc := fieldDescription(f); desc != "" {
			sub["description"] = desc
		}

		properties[name] = sub

		if !omitempty && f.Type.Kind() != reflect.Pointer {
			*required = append(*required, name)
		}
	}
}

// parseJSONField returns the JSON property name and options for a struct field. skip is true when the field
// is explicitly ignored via `json:"-"`.
func parseJSONField(f reflect.StructField) (name string, omitempty bool, skip bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}

	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitempty = true
		}
	}

	if name == "" && !f.Anonymous {
		name = f.Name
	}

	return name, omitempty, false
}

// fieldDescription returns the human description for a struct field, honouring both `desc` and `description`
// struct tags.
func fieldDescription(f reflect.StructField) string {
	if d := f.Tag.Get("desc"); d != "" {
		return d
	}
	return f.Tag.Get("description")
}
