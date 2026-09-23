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
	"io"
	"iter"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/xerror"
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

// ToolFuncS is [ToolFunc] with the acting subject, which is the shape of every nago use case. See
// [NewSubjectTool] and [NewUseCaseTool].
type ToolFuncS[In, Out any] func(auth.Subject, In) (Out, error)

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
	//
	// The subject is the one [Run] was called with, i.e. the human whose question started the turn. A tool
	// that reaches into the domain must pass it on to its use cases; that is what bounds the assistant to
	// exactly what the acting person could do by hand. See [NewSubjectTool] and [NewUseCaseTool].
	Invoke func(subject auth.Subject, args json.RawMessage) (json.RawMessage, error)

	// OpenFile, when set, marks this tool as a file-providing tool: instead of (or in addition to) a textual
	// result, the tool yields a file that [Run] uploads to the active provider and attaches to the
	// conversation as a [Media] block referencing it by id. This lets the model ask to "look at" a file
	// (e.g. a PDF from a drive) without the caller pre-attaching it.
	//
	// [Run] requires a [RunOptions.FileUploader] to perform the upload. The attached [Media] is added to the
	// same user turn that carries the tool_result blocks (never inside the tool_result itself, which the
	// provider would reject for file-id sources). When set, OpenFile takes precedence over [Invoke]. May be
	// nil. Create such a tool with [NewOpenFileTool] or [NewSubjectOpenFileTool].
	OpenFile func(subject auth.Subject, args json.RawMessage) (OpenedFile, error)

	// InvokeContent, when set, lets the tool return content blocks instead of a JSON document. They are placed
	// verbatim into the tool_result, which is how a tool hands an image to the model: a [Media] block with
	// inline [Source.Data] of an image type is accepted inside a tool_result by providers with vision support.
	// Providers which cannot embed media there replace it by a textual note, so a tool should always add a
	// [Text] block that is useful on its own. When set, InvokeContent takes precedence over [Invoke]; OpenFile
	// still takes precedence over both. Create such a tool with [NewSubjectContentTool].
	InvokeContent func(subject auth.Subject, args json.RawMessage) ([]Content, error)

	// Mutating marks a tool that changes state rather than merely reading it. It is a property of the tool,
	// not a phrase in its description, so it can actually be enforced: [RunOptions.OnBeforeToolCall] sees it
	// before the call happens, and a UI can refuse or confirm it.
	//
	// Relying on the model to honour a "this writes, ask first" instruction in the prompt is not a control -
	// it is a request. Set this flag instead.
	Mutating bool

	// Confirm is a short, human-readable sentence describing what this tool will change, shown when a
	// confirmation is requested. Optional; the tool name and its arguments are shown regardless. Only
	// meaningful together with Mutating.
	Confirm string

	// resultDoc is the rendered description of the return type, filled in by the constructors and moved into
	// the advertised description by [Tool.WithResultDoc]. It is unexported because it is derived, not
	// configured; a Tool built as a struct literal simply has none and WithResultDoc then does nothing.
	resultDoc string
}

// AsMutating marks the tool as state-changing and attaches a human-readable description of the effect. See
// [Tool.Mutating].
func (t Tool) AsMutating(confirm string) Tool {
	t.Mutating = true
	t.Confirm = confirm
	return t
}

// WithResultDoc appends a description of what the tool returns to its advertised description, so the model
// knows the shape of the answer before it calls.
//
// This is opt-in rather than automatic, and the reason is cost. No provider accepts an output schema for a
// tool - the definition carries a name, a description and an input schema, and nothing else - so the only
// place this can go is the description text, which is sent on every request for every tool for the lifetime
// of the conversation. A result whose field names already say what they are teaches the model its shape for
// free, with the first actual result.
//
// Use it where the shape is genuinely unobvious: aggregated rows, results with a truncation flag the model
// must react to, or fields whose meaning needs a sentence. The `desc` tags on the return type are carried
// over, and they only ever reach the model through this call.
func (t Tool) WithResultDoc() Tool {
	if t.resultDoc == "" {
		return t
	}

	if t.Def.Description == "" {
		t.Def.Description = t.resultDoc
		return t
	}

	t.Def.Description = t.Def.Description + "\n\n" + t.resultDoc

	return t
}

// OpenedFile is the file returned by an [Tool.OpenFile] invocation. The bytes are read lazily via Open so a
// tool can decline (return an error) cheaply before any data is transferred.
type OpenedFile struct {
	// Name is the human-readable file name (used as the upload name and shown to the model).
	Name string
	// MimeType classifies the content so the loop can pick the right handling: text types (see [file.IsText])
	// are injected inline as text, everything else is uploaded and attached as a Media block (image vs.
	// document is then decided by the provider).
	MimeType file.Type
	// Open yields the file content. It is called at most once by [Run].
	Open func() (io.ReadCloser, error)
}

// FileUploader uploads an [OpenedFile] to the active provider and returns the provider-native file id under
// which it can be referenced from message content (see [Source.FileID]). It decouples [Run] from the concrete
// provider.Files capability (which completion must not import); the caller wires it, typically to
// provider.Files().Put.
type FileUploader func(subject auth.Subject, f OpenedFile) (file.ID, error)

// NewTool wraps an arbitrary Go function of the form func(In) (Out, error) into a callable [Tool].
//
// The JSON schema describing In is derived automatically from the Go type via reflection (struct fields use
// their json tag for the property name and an optional `desc`/`description` struct tag for the property
// description). In and Out must be JSON marshalable.
//
// name must be a stable, unique identifier (the model references it by name). description should explain to
// the model when and how to use the tool. Both are validated; see [ValidateToolName].
//
// Use this only for tools that need no authorization, such as pure computation or static knowledge. Anything
// that touches domain data must go through [NewSubjectTool] or [NewUseCaseTool] so it is bounded by the
// acting subject.
func NewTool[In, Out any](name, description string, fn ToolFunc[In, Out]) Tool {
	return NewSubjectTool(name, description, func(_ auth.Subject, in In) (Out, error) {
		return fn(in)
	})
}

// NewSubjectTool is [NewTool] for a function that also receives the acting subject.
//
// This is the shape to reach for whenever a tool touches domain data: pass the subject on to the use case and
// the assistant is bounded by exactly the permissions of the person using it. It cannot read a record they
// may not read, and it cannot write one they may not write - not because the prompt says so, but because the
// use case audits the very same subject it would audit for a click in the UI.
//
// Because the subject arrives per call rather than per construction, a tool can be built once at start-up and
// shared by every window and every user.
func NewSubjectTool[In, Out any](name, description string, fn ToolFuncS[In, Out]) Tool {
	mustValidateTool(name, description)

	var zeroOut Out

	return Tool{
		Def:       newToolDef[In](name, description),
		resultDoc: renderResultDoc(reflect.TypeOf(&zeroOut).Elem()),
		Invoke: func(subject auth.Subject, args json.RawMessage) (json.RawMessage, error) {
			in, err := decodeToolArgs[In](name, args)
			if err != nil {
				return nil, err
			}

			out, err := fn(subject, in)
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

// NewUseCaseTool exposes a nago use case as a tool without any adapter in between.
//
// Use cases in nago already have the shape func(auth.Subject, Request) (Response, error), which is exactly
// what a tool needs, so the wrapper that would otherwise be written by hand for every single tool is not
// required:
//
//	tool := completion.NewUseCaseTool("assign_duty",
//		"Assigns a training requirement to the given employees.",
//		dutyUseCases.AssignDuty).AsMutating("creates training obligations for real people")
//
// The request type must be JSON marshalable and should carry `desc` struct tags, since it doubles as the
// documentation the model reads.
func NewUseCaseTool[In, Out any](name, description string, uc func(auth.Subject, In) (Out, error)) Tool {
	return NewSubjectTool(name, description, ToolFuncS[In, Out](uc))
}

// NewOpenFileTool wraps a Go function of the form func(In) (OpenedFile, error) into a file-providing [Tool].
// When the model calls it, [Run] executes fn, uploads the returned [OpenedFile] via [RunOptions.FileUploader]
// and attaches it to the conversation as a [Media] block (referenced by its provider file id) so the model
// can inspect its content — e.g. letting the model ask to "look at" a PDF stored in a drive.
//
// The JSON input schema for In is derived by reflection exactly like [NewTool]. The tool_result reported back
// to the model is a short textual confirmation; the actual file is added as a separate Media block on the same
// user turn (a file-id source is not valid inside a tool_result). If no [RunOptions.FileUploader] is
// configured, the call is reported to the model as an error.
//
// See [NewSubjectOpenFileTool] for the variant that receives the acting subject, which any file coming out of
// a permission-checked store needs.
func NewOpenFileTool[In any](name, description string, fn func(In) (OpenedFile, error)) Tool {
	return NewSubjectOpenFileTool(name, description, func(_ auth.Subject, in In) (OpenedFile, error) {
		return fn(in)
	})
}

// NewSubjectOpenFileTool is [NewOpenFileTool] for a function that also receives the acting subject, so the
// per-file permissions of the underlying store are enforced for the person actually asking.
func NewSubjectOpenFileTool[In any](name, description string, fn func(auth.Subject, In) (OpenedFile, error)) Tool {
	mustValidateTool(name, description)

	return Tool{
		Def: newToolDef[In](name, description),
		OpenFile: func(subject auth.Subject, args json.RawMessage) (OpenedFile, error) {
			in, err := decodeToolArgs[In](name, args)
			if err != nil {
				return OpenedFile{}, err
			}

			return fn(subject, in)
		},
	}
}

// NewSubjectContentTool wraps a function returning content blocks into a [Tool], see [Tool.InvokeContent].
// Use it for tools which must hand something other than JSON to the model, most notably images:
//
//	completion.NewSubjectContentTool("look", "…", func(s auth.Subject, in In) ([]completion.Content, error) {
//		return []completion.Content{
//			completion.Text{Text: "the current screen"},
//			completion.Media{MimeType: file.PNG, Source: completion.Source{Data: png}},
//		}, nil
//	})
func NewSubjectContentTool[In any](name, description string, fn func(auth.Subject, In) ([]Content, error)) Tool {
	mustValidateTool(name, description)

	return Tool{
		Def: newToolDef[In](name, description),
		InvokeContent: func(subject auth.Subject, args json.RawMessage) ([]Content, error) {
			in, err := decodeToolArgs[In](name, args)
			if err != nil {
				return nil, err
			}

			return fn(subject, in)
		},
	}
}

// newToolDef builds the advertised definition including the reflected JSON schema of In.
func newToolDef[In any](name, description string) ToolDef {
	var zeroIn In
	return newToolDefFromSchema(name, description, reflectSchema(reflect.TypeOf(&zeroIn).Elem()))
}

// newToolDefFromSchema builds the advertised definition from an already reflected (and possibly adjusted)
// schema.
func newToolDefFromSchema(name, description string, schema map[string]any) ToolDef {
	rawSchema, err := json.Marshal(schema)
	if err != nil {
		// A type whose schema cannot be marshalled is a programming error; encode it defensively as an
		// empty object so the tool stays usable.
		rawSchema = json.RawMessage(`{"type":"object"}`)
	}

	return ToolDef{
		Name:        name,
		Description: description,
		Schema:      rawSchema,
	}
}

// decodeToolArgs turns the raw arguments the model produced into the input type of a tool.
func decodeToolArgs[In any](name string, args json.RawMessage) (In, error) {
	var in In
	if len(args) > 0 {
		if err := json.Unmarshal(args, &in); err != nil {
			return in, fmt.Errorf("cannot decode arguments for tool %q: %w", name, err)
		}
	}

	return in, nil
}

// toolNamePattern is what every provider accepts and what a model can reliably reproduce: lower case ASCII,
// digits and underscores, starting with a letter.
var toolNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// ValidateToolName checks that name is usable as a tool identifier.
func ValidateToolName(name string) error {
	if !toolNamePattern.MatchString(name) {
		return fmt.Errorf("invalid tool name %q: expected lower case letters, digits and underscores, starting with a letter, at most 64 characters", name)
	}

	return nil
}

// mustValidateTool rejects a malformed tool at construction time.
//
// This panics rather than returning an error because both mistakes are programming errors that are always
// present or always absent - never data dependent - and both are otherwise diagnosed only by a confused
// model at runtime: a name the provider rejects fails the whole turn, and an empty description leaves the
// model guessing what the tool is for. Failing at start-up turns that into a stack trace pointing at the
// offending line.
func mustValidateTool(name, description string) {
	if err := ValidateToolName(name); err != nil {
		panic(err)
	}

	if strings.TrimSpace(description) == "" {
		panic(fmt.Errorf("tool %q has no description: the description is the only thing telling the model when to use it", name))
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

// BeforeToolCallFunc is consulted immediately before a tool is executed and may refuse the call by returning
// an error. The refusal is reported to the model as an error tool result, so the run continues and the model
// can explain itself or pick a different route.
//
// It is the enforcement point for [Tool.Mutating]: block writes outright in a read-only context, or hold the
// call until the user has confirmed it. Because it runs on the loop's goroutine it may block for as long as
// that takes.
type BeforeToolCallFunc func(subject auth.Subject, tool Tool, call ToolCall) error

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

	// FileUploader uploads binary files (images, PDFs, office documents) produced by [Tool.OpenFile] tools to
	// the active provider so they can be attached to the conversation by id. Text files (see [file.IsText])
	// are injected inline and never need an uploader. It is required only when a tool opens a binary file; a
	// nil uploader makes such calls fail (reported to the model as an error). Wire it to provider.Files().Put.
	FileUploader FileUploader

	// OnBeforeToolCall is consulted before each tool execution and may refuse it. Optional; see
	// [BeforeToolCallFunc].
	OnBeforeToolCall BeforeToolCallFunc
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
		// Two tools under one name means one of them is unreachable, and which one depends on slice order.
		// That is a wiring mistake worth reporting rather than a situation to silently pick a winner in.
		if _, dup := tools[t.Def.Name]; dup {
			return Result{}, opts.Messages, fmt.Errorf("duplicate tool name %q", t.Def.Name)
		}

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
		// attachments collects Media blocks contributed by OpenFile tools. They are appended to the SAME user
		// turn as the tool_result blocks (Anthropic combines consecutive user turns anyway, and a file-id
		// source is not valid inside a tool_result). Keeping them on the tool-result turn keeps the history
		// valid and lets the model see the file on the next turn.
		var attachments []Content
		for _, call := range calls {
			call := call
			notify(Progress{Phase: PhaseToolStarted, Turn: turn, ToolCall: &call})

			result, media := executeToolCall(subject, tools, call, opts.FileUploader, opts.OnBeforeToolCall)

			notify(Progress{Phase: PhaseToolCompleted, Turn: turn, ToolCall: &call, ToolResult: &result})

			results = append(results, result)
			attachments = append(attachments, media...)
		}

		content := append(results, attachments...)
		history = append(history, Message{Role: User, Content: content})
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
//
// For a file-providing tool (see [Tool.OpenFile]) it additionally uploads the produced file via uploader and
// returns the resulting [Media] block(s) to be attached to the user turn (never inside the tool_result). The
// tool_result itself is then a short textual confirmation. The returned media slice is empty for regular
// tools or when the file could not be provided/uploaded.
func executeToolCall(subject auth.Subject, tools map[string]Tool, call ToolCall, uploader FileUploader, before BeforeToolCallFunc) (ToolResult, []Content) {
	tool, ok := tools[call.Name]
	if !ok {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: fmt.Sprintf("unknown tool %q", call.Name)}},
			IsError:    true,
		}, nil
	}

	// Give the caller the chance to refuse this call - to ask the user first, or to block a mutating tool
	// outright. A refusal is reported to the model like any other tool error, so it can explain itself or
	// choose a different route instead of the whole turn collapsing.
	if before != nil {
		if err := before(subject, tool, call); err != nil {
			return ToolResult{
				ToolCallID: call.ID,
				Content:    []Content{Text{Text: toolErrorText(err)}},
				IsError:    true,
			}, nil
		}
	}

	// File-providing tool: inject text files inline or upload binary files and attach them as a Media block.
	if tool.OpenFile != nil {
		return executeOpenFileCall(subject, call, tool, uploader)
	}

	if tool.InvokeContent != nil {
		content, err := tool.InvokeContent(subject, call.Arguments)
		if err != nil {
			return ToolResult{
				ToolCallID: call.ID,
				Content:    []Content{Text{Text: toolErrorText(err)}},
				IsError:    true,
			}, nil
		}

		if len(content) == 0 {
			// an empty tool_result is rejected by some providers and tells the model nothing
			content = []Content{Text{Text: "(no content)"}}
		}

		return ToolResult{ToolCallID: call.ID, Content: content}, nil
	}

	if tool.Invoke == nil {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: fmt.Sprintf("tool %q cannot be invoked", call.Name)}},
			IsError:    true,
		}, nil
	}

	out, err := tool.Invoke(subject, call.Arguments)
	if err != nil {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: toolErrorText(err)}},
			IsError:    true,
		}, nil
	}

	return ToolResult{
		ToolCallID: call.ID,
		Content:    []Content{Text{Text: string(out)}},
	}, nil
}

// toolErrorText renders a tool failure for the model using the same classifier which drives
// the user facing banners, see [xerror.Present].
//
// A raw err.Error() is a poor signal: a permission denial arrives as a bare permission id with
// no indication that it is an authorization problem, so the model cannot tell "retry
// differently" apart from "you are not allowed to do this at all".
//
// Note on disclosure: unrecognized errors keep their original message. The tool already ran
// under the caller's own auth.Subject, so the model operates inside the user's authorization
// boundary, and withholding the detail would only prevent the model from recovering. Only the
// classified categories are rewritten, because for those a better phrasing exists.
func toolErrorText(err error) string {
	if err == nil {
		return ""
	}

	bundler := xerror.BundlerOrDefault(nil)

	p, ok := xerror.Present(bundler, err)
	if !ok {
		return err.Error()
	}

	var sb strings.Builder
	sb.WriteString(p.Title)
	sb.WriteString(": ")
	sb.WriteString(p.Message)

	if p.Denied() {
		sb.WriteString(" ")
		sb.WriteString(xerror.StrDeniedRetryHint.Get(bundler))
	}

	// Field bound validation messages are the actionable part: they tell the model exactly
	// which argument to correct.
	for _, key := range slices.Sorted(maps.Keys(p.Fields)) {
		sb.WriteString("\n- ")
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(p.Fields[key])
	}

	return sb.String()
}

// executeOpenFileCall handles a [Tool.OpenFile] tool call. Text files (see [file.IsText]) are read and
// injected inline as the textual tool_result, so the model can read them directly without an upload - this
// works with every provider and needs no [FileUploader]. Binary files (images, PDFs, office documents) are
// uploaded via uploader and produce (a) a short textual tool_result confirming the attachment and (b) a
// [Media] block referencing the uploaded file by id, to be added to the user turn. Any failure (no uploader
// for a binary file, tool error, open/upload error) is reported back to the model as an error tool_result
// with no attachment.
func executeOpenFileCall(subject auth.Subject, call ToolCall, tool Tool, uploader FileUploader) (ToolResult, []Content) {
	toolErr := func(msg string) (ToolResult, []Content) {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: msg}},
			IsError:    true,
		}, nil
	}

	opened, err := tool.OpenFile(subject, call.Arguments)
	if err != nil {
		return toolErr(err.Error())
	}

	// Text files are fed to the model inline instead of being uploaded as a binary attachment.
	if file.IsText(opened.MimeType) {
		return injectFileAsText(call, opened)
	}

	if uploader == nil {
		return toolErr("cannot attach file: no file uploader configured for this run")
	}

	fileID, err := uploader(subject, opened)
	if err != nil {
		return toolErr(fmt.Sprintf("cannot attach file %q: %v", opened.Name, err))
	}

	media := Media{MimeType: opened.MimeType, Source: Source{FileID: option.Some(fileID)}}
	confirm := ToolResult{
		ToolCallID: call.ID,
		Content: []Content{Text{Text: fmt.Sprintf(
			"Attached file %q (%s) to the conversation; its content follows as an attachment.",
			opened.Name, opened.MimeType)}},
	}

	return confirm, []Content{media}
}

// maxInlineTextBytes caps how much of a text file is injected inline by [injectFileAsText]. Larger files are
// truncated with a note so a single tool call cannot blow the model's context window.
const maxInlineTextBytes = 256 * 1024

// injectFileAsText reads a text [OpenedFile] and returns its content as the textual tool_result (no upload, no
// [Media] block). The content is capped at [maxInlineTextBytes] and a truncation note is appended when the
// file is larger. The returned media slice is always nil.
func injectFileAsText(call ToolCall, opened OpenedFile) (ToolResult, []Content) {
	toolErr := func(msg string) (ToolResult, []Content) {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    []Content{Text{Text: msg}},
			IsError:    true,
		}, nil
	}

	rc, err := opened.Open()
	if err != nil {
		return toolErr(fmt.Sprintf("cannot open file %q: %v", opened.Name, err))
	}
	defer rc.Close()

	// read one byte more than the cap so we can detect (and report) truncation
	buf, err := io.ReadAll(io.LimitReader(rc, maxInlineTextBytes+1))
	if err != nil {
		return toolErr(fmt.Sprintf("cannot read file %q: %v", opened.Name, err))
	}

	truncated := false
	if len(buf) > maxInlineTextBytes {
		buf = buf[:maxInlineTextBytes]
		truncated = true
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Contents of %q (%s):\n", opened.Name, opened.MimeType)
	sb.Write(buf)
	if truncated {
		fmt.Fprintf(&sb, "\n\n[truncated: showing the first %d bytes; the file is larger]", maxInlineTextBytes)
	}

	return ToolResult{
		ToolCallID: call.ID,
		Content:    []Content{Text{Text: sb.String()}},
	}, nil
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

		if isRequiredField(f, omitempty) {
			*required = append(*required, name)
		}
	}
}

// isRequiredField decides whether a struct field is advertised as mandatory.
//
// A field is required unless it says otherwise. The three ways to say otherwise are, in order of precedence:
//
//   - `optional:"true"`, which states it outright. This is the one to reach for on a domain type that is
//     exposed to the model directly: a filter type carries no omitempty and is not built from pointers, yet
//     nearly every field on it is optional by nature.
//   - a json `omitempty`, because a field that may be left out of the encoding may be left out of the call.
//   - a pointer, because its whole point is the absence of a value.
//
// The default is deliberately "required" rather than "optional": stating what a tool must be given is the
// part a model gets wrong, and an inverted default would silently weaken every existing tool at once.
func isRequiredField(f reflect.StructField, omitempty bool) bool {
	if isOptionalTag(f.Tag.Get("optional")) {
		return false
	}

	if omitempty {
		return false
	}

	return f.Type.Kind() != reflect.Pointer
}

// isOptionalTag reads the optional struct tag. Anything but an explicit truthy value keeps the field
// required, so a typo cannot quietly turn a mandatory argument into an omittable one.
func isOptionalTag(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "yes", "1":
		return true
	default:
		return false
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

// DefaultSeqToolLimit caps how many elements a [NewSeqTool] returns when the model does not ask for a
// specific limit.
//
// A listing use case happily yields every record it has. Handing all of them to a model is not merely
// wasteful: a few thousand rows exhaust the context window, and the turn fails with an error that looks like
// a provider problem rather than a missing limit. So listings are capped by default and say when they were.
const DefaultSeqToolLimit = 200

// MaxSeqToolLimit is the largest limit a model may request from a [NewSeqTool], regardless of what it asks
// for. It exists so a model cannot talk itself past the safeguard above.
const MaxSeqToolLimit = 1000

// SeqToolResult is what a [NewSeqTool] returns to the model.
type SeqToolResult[T any] struct {
	// Items are the elements that were read, at most Limit many.
	Items []T `json:"items" desc:"the entries that were read"`

	// Count is len(Items), stated explicitly because models are unreliable at counting.
	Count int `json:"count" desc:"number of entries in items"`

	// Truncated reports that the underlying listing had more to offer and the result was cut short. When
	// this is true, the model must not treat Items as the complete answer - it should narrow its filter
	// instead.
	Truncated bool `json:"truncated" desc:"true when there were more entries than the limit allowed; the list is then incomplete and the filter should be narrowed rather than the result treated as final"`

	// Note explains a truncation in words, so the model reacts to it even if it ignores the boolean.
	Note string `json:"note,omitempty" desc:"present only when truncated, explaining what to do about it"`
}

// seqLimitInput carries the limit the model may choose. It is decoded from the same argument object as the
// caller's filter type, which keeps the advertised schema flat: the model sees the filter fields and the
// limit side by side, exactly as it would expect from a listing.
type seqLimitInput struct {
	Limit int `json:"limit,omitempty" desc:"maximum number of entries to return; defaults to 200, at most 1000. Narrow the other filters instead of raising this."`
}

// NewSeqTool exposes a listing use case of the form func(auth.Subject, Filter) iter.Seq2[T, error] as a tool.
//
// This is the second shape nago use cases come in, and it needs more than a signature adapter: the result is
// bounded (see [DefaultSeqToolLimit]) and the model is told when it was cut short, so it narrows its filter
// rather than reasoning from a silently partial list.
//
//	tool := completion.NewSeqTool("list_employees",
//		"Lists employee profiles with their teams, units and roles.",
//		peopleUseCases.ListEmployees)
//
// The advertised input schema is the filter type plus a `limit` property.
func NewSeqTool[In, T any](name, description string, uc func(auth.Subject, In) iter.Seq2[T, error]) Tool {
	mustValidateTool(name, description)

	var zeroOut SeqToolResult[T]

	return Tool{
		Def:       newSeqToolDef[In](name, description),
		resultDoc: renderResultDoc(reflect.TypeOf(&zeroOut).Elem()),
		Invoke: func(subject auth.Subject, args json.RawMessage) (json.RawMessage, error) {
			filter, err := decodeToolArgs[In](name, args)
			if err != nil {
				return nil, err
			}

			bounds, err := decodeToolArgs[seqLimitInput](name, args)
			if err != nil {
				return nil, err
			}

			limit := bounds.Limit
			if limit <= 0 {
				limit = DefaultSeqToolLimit
			}
			if limit > MaxSeqToolLimit {
				limit = MaxSeqToolLimit
			}

			res := SeqToolResult[T]{Items: make([]T, 0, min(limit, 64))}

			for item, err := range uc(subject, filter) {
				if err != nil {
					return nil, err
				}

				// Stop at the limit and remember that there was more, so truncation is a fact rather than a
				// guess: a listing of exactly limit elements is complete and must not be reported as cut short.
				if len(res.Items) == limit {
					res.Truncated = true
					break
				}

				res.Items = append(res.Items, item)
			}

			res.Count = len(res.Items)

			if res.Truncated {
				res.Note = fmt.Sprintf("Only the first %d entries are shown; there are more. Narrow the filter instead of assuming this is the complete list.", limit)
			}

			raw, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("cannot encode result of tool %q: %w", name, err)
			}

			return raw, nil
		},
	}
}

// NewSliceTool exposes a listing use case of the form func(auth.Subject, Filter) ([]T, error) as a tool.
//
// It is [NewSeqTool] for the other shape a listing comes in, and it exists for the same reason: the result is
// bounded (see [DefaultSeqToolLimit]) and the model is told when it was cut short. A use case that reads
// everything into a slice is if anything more dangerous than an iterator, because the cost is already paid by
// the time the tool sees it - the only thing left to protect is the model's context window.
//
//	tool := completion.NewSliceTool("list_employees",
//		"Lists employee profiles with their teams, units and roles.",
//		peopleUseCases.ListEmployees)
//
// The advertised input schema is the filter type plus a `limit` property.
func NewSliceTool[In, T any](name, description string, uc func(auth.Subject, In) ([]T, error)) Tool {
	return NewSeqTool(name, description, func(subject auth.Subject, in In) iter.Seq2[T, error] {
		return func(yield func(T, error) bool) {
			items, err := uc(subject, in)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}

			for _, item := range items {
				if !yield(item, nil) {
					return
				}
			}
		}
	})
}

// newSeqToolDef reflects the filter type and merges in the limit property.
func newSeqToolDef[In any](name, description string) ToolDef {
	var zeroIn In
	schema := reflectSchema(reflect.TypeOf(&zeroIn).Elem())

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		// The filter is not an object (e.g. a bare string). Advertise the reflected schema unchanged rather
		// than forcing a limit into something that cannot carry it.
		return newToolDefFromSchema(name, description, schema)
	}

	limitSchema := reflectSchema(reflect.TypeOf(seqLimitInput{}))
	if limitProps, ok := limitSchema["properties"].(map[string]any); ok {
		for k, v := range limitProps {
			// A filter that already defines "limit" wins; it presumably means something specific.
			if _, taken := props[k]; !taken {
				props[k] = v
			}
		}
	}

	return newToolDefFromSchema(name, description, schema)
}
