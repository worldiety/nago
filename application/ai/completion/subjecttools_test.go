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
	"iter"
	"slices"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// testSubject is a stand-in for the acting user. Only the identity matters here: the point of the subject
// tools is that whatever the use case receives is the very same subject the run was started with.
type testSubject struct {
	auth.Subject
	id user.ID
}

func (s testSubject) ID() user.ID                      { return s.id }
func (s testSubject) Valid() bool                      { return s.id != "" }
func (s testSubject) HasPermission(permission.ID) bool { return true }
func (s testSubject) Audit(permission.ID) error        { return nil }
func (s testSubject) HasGroup(group.ID) bool           { return false }
func (s testSubject) HasResourcePermission(rebac.Namespace, rebac.Instance, permission.ID) bool {
	return true
}
func (s testSubject) AuditResource(rebac.Namespace, rebac.Instance, permission.ID) error { return nil }

type echoIn struct {
	Text string `json:"text" desc:"anything"`
}

type echoOut struct {
	Actor string `json:"actor"`
	Text  string `json:"text"`
}

// TestNewSubjectTool_PassesSubject is the whole point of the constructor: without this, every tool has to be
// rebuilt per turn just to close over the subject.
func TestNewSubjectTool_PassesSubject(t *testing.T) {
	tool := NewSubjectTool("echo", "echoes the input", func(subject auth.Subject, in echoIn) (echoOut, error) {
		return echoOut{Actor: string(subject.ID()), Text: in.Text}, nil
	})

	raw, err := tool.Invoke(testSubject{id: "alice"}, json.RawMessage(`{"text":"hi"}`))
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}

	var got echoOut
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("cannot decode result: %v", err)
	}

	if got.Actor != "alice" || got.Text != "hi" {
		t.Errorf("got %+v, want actor alice and text hi", got)
	}
}

// TestNewSubjectTool_IsReusableAcrossSubjects guards the property that makes package level tools safe: one
// tool value, many actors, no leakage between them.
func TestNewSubjectTool_IsReusableAcrossSubjects(t *testing.T) {
	tool := NewSubjectTool("echo", "echoes the input", func(subject auth.Subject, in echoIn) (echoOut, error) {
		return echoOut{Actor: string(subject.ID())}, nil
	})

	for _, want := range []user.ID{"alice", "bob", "alice"} {
		raw, err := tool.Invoke(testSubject{id: want}, json.RawMessage(`{"text":""}`))
		if err != nil {
			t.Fatalf("invoke failed: %v", err)
		}

		var got echoOut
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("cannot decode result: %v", err)
		}

		if got.Actor != string(want) {
			t.Errorf("got actor %q, want %q", got.Actor, want)
		}
	}
}

// TestNewUseCaseTool_TakesAUseCaseUnchanged demonstrates that a nago use case needs no adapter at all.
func TestNewUseCaseTool_TakesAUseCaseUnchanged(t *testing.T) {
	// This is deliberately written as a plain use case signature, not as a literal shaped for the test.
	var useCase func(auth.Subject, echoIn) (echoOut, error) = func(subject auth.Subject, in echoIn) (echoOut, error) {
		return echoOut{Actor: string(subject.ID()), Text: in.Text}, nil
	}

	tool := NewUseCaseTool("echo", "echoes the input", useCase)

	raw, err := tool.Invoke(testSubject{id: "carol"}, json.RawMessage(`{"text":"x"}`))
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}

	if !strings.Contains(string(raw), `"carol"`) {
		t.Errorf("the use case did not receive the acting subject: %s", raw)
	}
}

// TestNewUseCaseTool_PropagatesErrors makes sure a permission denial reaches the model as an error rather
// than being swallowed into an empty but successful looking result.
func TestNewUseCaseTool_PropagatesErrors(t *testing.T) {
	denied := errors.New("permission denied")

	tool := NewUseCaseTool("echo", "echoes the input", func(auth.Subject, echoIn) (echoOut, error) {
		return echoOut{}, denied
	})

	if _, err := tool.Invoke(testSubject{id: "dave"}, json.RawMessage(`{}`)); !errors.Is(err, denied) {
		t.Errorf("got %v, want the use case error", err)
	}
}

type listFilter struct {
	Team string `json:"team" desc:"optional team filter"`
}

// listOf builds a listing use case yielding n numbered entries.
func listOf(n int) func(auth.Subject, listFilter) iter.Seq2[string, error] {
	return func(_ auth.Subject, _ listFilter) iter.Seq2[string, error] {
		return func(yield func(string, error) bool) {
			for i := range n {
				if !yield(fmt.Sprintf("item-%d", i), nil) {
					return
				}
			}
		}
	}
}

func invokeSeq(t *testing.T, tool Tool, args string) SeqToolResult[string] {
	t.Helper()

	raw, err := tool.Invoke(testSubject{id: "eve"}, json.RawMessage(args))
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}

	var got SeqToolResult[string]
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("cannot decode result: %v", err)
	}

	return got
}

// TestNewSeqTool_CapsAndReportsTruncation covers the reason NewSeqTool exists at all: an unbounded listing
// handed to a model exhausts the context window, and a silently partial list is worse than an explicit one.
func TestNewSeqTool_CapsAndReportsTruncation(t *testing.T) {
	tool := NewSeqTool("list_items", "lists items", listOf(DefaultSeqToolLimit+5))

	got := invokeSeq(t, tool, `{}`)

	if len(got.Items) != DefaultSeqToolLimit {
		t.Errorf("got %d items, want the default limit of %d", len(got.Items), DefaultSeqToolLimit)
	}

	if got.Count != len(got.Items) {
		t.Errorf("count %d does not match the %d items reported", got.Count, len(got.Items))
	}

	if !got.Truncated {
		t.Error("a listing that was cut short did not say so")
	}

	if got.Note == "" {
		t.Error("a truncated listing carries no explanation, so a model ignoring the flag is misled")
	}
}

// TestNewSeqTool_ExactlyAtTheLimitIsNotTruncated is the off-by-one that would otherwise make a model chase a
// non-existent remainder forever.
func TestNewSeqTool_ExactlyAtTheLimitIsNotTruncated(t *testing.T) {
	tool := NewSeqTool("list_items", "lists items", listOf(DefaultSeqToolLimit))

	got := invokeSeq(t, tool, `{}`)

	if got.Truncated {
		t.Error("a complete listing of exactly the limit was reported as truncated")
	}

	if len(got.Items) != DefaultSeqToolLimit {
		t.Errorf("got %d items, want %d", len(got.Items), DefaultSeqToolLimit)
	}
}

// TestNewSeqTool_HonoursAndBoundsTheRequestedLimit checks both directions: the model may ask for less, but it
// cannot talk itself past the safeguard.
func TestNewSeqTool_HonoursAndBoundsTheRequestedLimit(t *testing.T) {
	tool := NewSeqTool("list_items", "lists items", listOf(MaxSeqToolLimit+50))

	if got := invokeSeq(t, tool, `{"limit":5}`); len(got.Items) != 5 {
		t.Errorf("got %d items, want the requested 5", len(got.Items))
	}

	got := invokeSeq(t, tool, `{"limit":100000}`)
	if len(got.Items) != MaxSeqToolLimit {
		t.Errorf("got %d items, want the hard cap of %d", len(got.Items), MaxSeqToolLimit)
	}
}

// TestNewSeqTool_PassesTheFilterAndSubject makes sure the added limit field does not swallow the caller's own
// arguments.
func TestNewSeqTool_PassesTheFilterAndSubject(t *testing.T) {
	var gotTeam string
	var gotActor user.ID

	tool := NewSeqTool("list_items", "lists items", func(subject auth.Subject, f listFilter) iter.Seq2[string, error] {
		gotTeam = f.Team
		gotActor = subject.ID()
		return func(yield func(string, error) bool) {}
	})

	invokeSeq(t, tool, `{"team":"platform","limit":3}`)

	if gotTeam != "platform" {
		t.Errorf("got team %q, want platform", gotTeam)
	}

	if gotActor != "eve" {
		t.Errorf("got actor %q, want eve", gotActor)
	}
}

// TestNewSeqTool_SchemaCarriesFilterAndLimit pins down that the model is told about both.
func TestNewSeqTool_SchemaCarriesFilterAndLimit(t *testing.T) {
	tool := NewSeqTool("list_items", "lists items", listOf(1))

	var schema map[string]any
	if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
		t.Fatalf("cannot decode schema: %v", err)
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no properties: %v", schema)
	}

	for _, want := range []string{"team", "limit"} {
		if _, ok := props[want]; !ok {
			t.Errorf("schema is missing property %q: %v", want, props)
		}
	}
}

// TestNewSeqTool_PropagatesErrors ensures a failing listing is not reported as an empty one.
func TestNewSeqTool_PropagatesErrors(t *testing.T) {
	boom := errors.New("backend unavailable")

	tool := NewSeqTool("list_items", "lists items", func(auth.Subject, listFilter) iter.Seq2[string, error] {
		return func(yield func(string, error) bool) { yield("", boom) }
	})

	if _, err := tool.Invoke(testSubject{id: "eve"}, json.RawMessage(`{}`)); !errors.Is(err, boom) {
		t.Errorf("got %v, want the listing error", err)
	}
}

// TestValidateToolName covers the names a provider accepts and the ones it does not.
func TestValidateToolName(t *testing.T) {
	valid := []string{"a", "add", "list_employees", "tool_2"}
	for _, name := range valid {
		if err := ValidateToolName(name); err != nil {
			t.Errorf("%q should be valid: %v", name, err)
		}
	}

	invalid := []string{"", "Add", "list-employees", "2fast", "_leading", "with space", strings.Repeat("a", 65)}
	for _, name := range invalid {
		if err := ValidateToolName(name); err == nil {
			t.Errorf("%q should be rejected", name)
		}
	}
}

// TestNewToolRejectsMalformedDeclarations fails the build of a tool at construction rather than leaving a
// confused model to discover the problem at runtime.
func TestNewToolRejectsMalformedDeclarations(t *testing.T) {
	cases := map[string]func(){
		"bad name":          func() { NewTool("Bad Name", "does something", func(echoIn) (echoOut, error) { return echoOut{}, nil }) },
		"empty description": func() { NewTool("fine_name", "  ", func(echoIn) (echoOut, error) { return echoOut{}, nil }) },
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected a panic")
				}
			}()

			build()
		})
	}
}

// TestAsMutating is the structural replacement for writing "THIS WRITES" into the description and hoping the
// model complies.
func TestAsMutating(t *testing.T) {
	tool := NewTool("save_thing", "saves a thing", func(echoIn) (echoOut, error) { return echoOut{}, nil })

	if tool.Mutating {
		t.Error("a plain tool must not be marked as mutating")
	}

	mutating := tool.AsMutating("overwrites the stored thing")

	if !mutating.Mutating || mutating.Confirm != "overwrites the stored thing" {
		t.Errorf("got %+v, want a mutating tool with a confirmation text", mutating)
	}

	if tool.Mutating {
		t.Error("AsMutating modified the receiver instead of returning a copy")
	}
}

// toolCallThen builds a fake provider that first requests the named tool and then answers with plain text.
func toolCallThen(name, args string) *fakeCompletions {
	return &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: name, Arguments: json.RawMessage(args)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "done"}}},
				StopReason: StopEndTurn,
			},
		},
	}
}

// runWith drives one tool round-trip and returns the tool result the model was given.
func runWith(t *testing.T, tool Tool, before BeforeToolCallFunc) ToolResult {
	t.Helper()

	_, history, err := Run(testSubject{id: "frank"}, toolCallThen(tool.Def.Name, `{"text":"x"}`), RunOptions{
		Options:          Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "go"}}}}},
		Tools:            []Tool{tool},
		OnBeforeToolCall: before,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	for _, msg := range history {
		for _, c := range msg.Content {
			if tr, ok := c.(ToolResult); ok {
				return tr
			}
		}
	}

	t.Fatal("no tool result in history")
	return ToolResult{}
}

// TestRun_OnBeforeToolCallCanRefuse is the mechanism that makes [Tool.Mutating] enforceable rather than
// advisory: the gate sits in front of the function call, not in the prompt.
func TestRun_OnBeforeToolCallCanRefuse(t *testing.T) {
	executed := false

	tool := NewTool("save_thing", "saves a thing", func(echoIn) (echoOut, error) {
		executed = true
		return echoOut{}, nil
	}).AsMutating("stores the thing")

	result := runWith(t, tool, func(_ auth.Subject, tool Tool, _ ToolCall) error {
		if tool.Mutating {
			return errors.New("the user declined")
		}
		return nil
	})

	if executed {
		t.Error("the refused tool ran anyway")
	}

	if !result.IsError {
		t.Error("the refusal was not reported to the model as an error")
	}

	if !strings.Contains(fmt.Sprint(result.Content), "declined") {
		t.Errorf("the model was not told why: %v", result.Content)
	}
}

// TestRun_RefusalDoesNotAbortTheRun matters because a declined change is a normal outcome: the assistant must
// be able to say so and offer something else, not fail the whole turn.
func TestRun_RefusalDoesNotAbortTheRun(t *testing.T) {
	tool := NewTool("save_thing", "saves a thing", func(echoIn) (echoOut, error) {
		return echoOut{}, nil
	}).AsMutating("stores the thing")

	res, _, err := Run(testSubject{id: "frank"}, toolCallThen("save_thing", `{}`), RunOptions{
		Options:          Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "go"}}}}},
		Tools:            []Tool{tool},
		OnBeforeToolCall: func(auth.Subject, Tool, ToolCall) error { return errors.New("no") },
	})
	if err != nil {
		t.Fatalf("a refused tool call aborted the run: %v", err)
	}

	if res.StopReason != StopEndTurn {
		t.Errorf("got stop reason %v, want the model's final answer", res.StopReason)
	}
}

// TestRun_OnBeforeToolCallLetsReadsThrough keeps the gate narrow, so a conversation that only reads never
// stops to ask.
func TestRun_OnBeforeToolCallLetsReadsThrough(t *testing.T) {
	executed := false

	tool := NewTool("read_thing", "reads a thing", func(echoIn) (echoOut, error) {
		executed = true
		return echoOut{Text: "value"}, nil
	})

	result := runWith(t, tool, func(_ auth.Subject, tool Tool, _ ToolCall) error {
		if tool.Mutating {
			return errors.New("the user declined")
		}
		return nil
	})

	if !executed {
		t.Error("a read-only tool was blocked by the mutation gate")
	}

	if result.IsError {
		t.Errorf("a read-only tool was reported as an error: %v", result.Content)
	}
}

// TestRun_RejectsDuplicateToolNames turns a wiring mistake into a message instead of a coin flip over which
// of the two tools becomes unreachable.
func TestRun_RejectsDuplicateToolNames(t *testing.T) {
	a := NewTool("thing", "does one thing", func(echoIn) (echoOut, error) { return echoOut{}, nil })
	b := NewTool("thing", "does another thing", func(echoIn) (echoOut, error) { return echoOut{}, nil })

	_, _, err := Run(testSubject{id: "frank"}, &fakeCompletions{}, RunOptions{
		Options: Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "go"}}}}},
		Tools:   []Tool{a, b},
	})

	if err == nil || !strings.Contains(err.Error(), "duplicate tool name") {
		t.Errorf("got %v, want a duplicate tool name error", err)
	}
}

type requiredIn struct {
	Mandatory string  `json:"mandatory" desc:"must be given"`
	ByTag     string  `json:"byTag" optional:"true" desc:"opted out explicitly"`
	ByOmit    string  `json:"byOmit,omitempty" desc:"opted out via json"`
	ByPointer *string `json:"byPointer" desc:"opted out by being a pointer"`
	Typo      string  `json:"typo" optional:"maybe" desc:"a typo must not silently loosen the contract"`
}

func requiredOf(t *testing.T, tool Tool) []string {
	t.Helper()

	var schema map[string]any
	if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
		t.Fatalf("cannot decode schema: %v", err)
	}

	raw, _ := schema["required"].([]any)
	got := make([]string, 0, len(raw))
	for _, v := range raw {
		got = append(got, v.(string))
	}

	slices.Sort(got)
	return got
}

// TestRequiredIsTheDefault pins the semantics down as they always were. Inverting them would have made every
// argument of every existing tool optional at once - including the ones a write tool cannot work without,
// where the model would then be free to omit them and hit a runtime error instead of a schema that told it
// so up front.
func TestRequiredIsTheDefault(t *testing.T) {
	tool := NewTool("check", "checks something", func(requiredIn) (echoOut, error) {
		return echoOut{}, nil
	})

	want := []string{"mandatory", "typo"}
	if got := requiredOf(t, tool); !slices.Equal(got, want) {
		t.Errorf("got required %v, want %v", got, want)
	}
}

// TestOptionalTagOptsOut covers the addition: a domain type exposed to the model directly carries neither
// omitempty nor pointers, yet nearly every field of a filter is optional by nature.
func TestOptionalTagOptsOut(t *testing.T) {
	type filter struct {
		Team  string `json:"team" optional:"true"`
		Unit  string `json:"unit" optional:"yes"`
		Role  string `json:"role" optional:"1"`
		Query string `json:"query"`
	}

	tool := NewTool("search", "searches something", func(filter) (echoOut, error) {
		return echoOut{}, nil
	})

	want := []string{"query"}
	if got := requiredOf(t, tool); !slices.Equal(got, want) {
		t.Errorf("got required %v, want %v", got, want)
	}
}

// TestSeqToolLimitIsOptional guards the framework's own addition: a listing that demands a limit would force
// the model to invent one on every call.
func TestSeqToolLimitIsOptional(t *testing.T) {
	tool := NewSeqTool("list_items", "lists items", listOf(1))

	if got := requiredOf(t, tool); slices.Contains(got, "limit") {
		t.Errorf("limit is advertised as required: %v", got)
	}
}

type docOut struct {
	Person    string `json:"person" desc:"full name of the employee"`
	Overdue   int    `json:"overdue"`
	Truncated bool   `json:"truncated" desc:"more rows exist; narrow the filter"`
}

// TestWithResultDocDescribesTheReturnType covers the only path by which a return type reaches the model at
// all: no provider accepts an output schema, so it has to go into the description text.
func TestWithResultDocDescribesTheReturnType(t *testing.T) {
	tool := NewTool("team_duties", "reports the team", func(echoIn) ([]docOut, error) {
		return nil, nil
	}).WithResultDoc()

	desc := tool.Def.Description

	if !strings.HasPrefix(desc, "reports the team") {
		t.Errorf("the original description was lost: %q", desc)
	}

	for _, want := range []string{
		"a list of objects",
		"person (string, always present): full name of the employee",
		"overdue (integer, always present)",
		"truncated (boolean, always present): more rows exist",
	} {
		if !strings.Contains(desc, want) {
			t.Errorf("description is missing %q:\n%s", want, desc)
		}
	}
}

// TestResultDocIsOptIn is the cost decision made visible: the text is sent on every request for every tool,
// so a tool that did not ask for it must not carry it.
func TestResultDocIsOptIn(t *testing.T) {
	tool := NewTool("team_duties", "reports the team", func(echoIn) ([]docOut, error) {
		return nil, nil
	})

	if tool.Def.Description != "reports the team" {
		t.Errorf("a tool that did not ask for a result doc carries one: %q", tool.Def.Description)
	}
}

// TestResultDocIsStable protects provider-side prompt caching: a description that reorders itself between
// two builds would invalidate the cache for no benefit.
func TestResultDocIsStable(t *testing.T) {
	build := func() string {
		return NewTool("team_duties", "reports the team", func(echoIn) ([]docOut, error) {
			return nil, nil
		}).WithResultDoc().Def.Description
	}

	first := build()
	for range 20 {
		if got := build(); got != first {
			t.Fatalf("the rendered description is unstable:\n%s\n---\n%s", first, got)
		}
	}
}

// TestWithResultDocOnAStructLiteralIsHarmless covers a Tool that was not built by a constructor and thus has
// no derived documentation to move.
func TestWithResultDocOnAStructLiteralIsHarmless(t *testing.T) {
	tool := Tool{Def: ToolDef{Name: "manual", Description: "hand built"}}

	if got := tool.WithResultDoc(); got.Def.Description != "hand built" {
		t.Errorf("got %q, want the description untouched", got.Def.Description)
	}
}

// TestResultDocNamesTypesWithoutArticles guards a detail that reads as sloppiness to whoever has to review
// the prompt: a field type is named "integer", not "an integer", because it already sits in parentheses.
func TestResultDocNamesTypesWithoutArticles(t *testing.T) {
	tool := NewTool("team_duties", "reports the team", func(echoIn) ([]docOut, error) {
		return nil, nil
	}).WithResultDoc()

	for _, bad := range []string{"(an ", "(a ", "list of an ", "list of a "} {
		if strings.Contains(tool.Def.Description, bad) {
			t.Errorf("description contains %q:\n%s", bad, tool.Def.Description)
		}
	}
}

// TestResultDocExpandsNestedStructs is the reason the renderer is not a one-liner: collapsing the most
// important field of a listing to "(array)" tells the model nothing at all.
func TestResultDocExpandsNestedStructs(t *testing.T) {
	type row struct {
		Person string   `json:"person"`
		Lines  []docOut `json:"lines" desc:"one entry per obligation"`
	}

	tool := NewTool("team_duties", "reports the team", func(echoIn) ([]row, error) {
		return nil, nil
	}).WithResultDoc()

	desc := tool.Def.Description

	if !strings.Contains(desc, "one entry per obligation — a list of objects with:") {
		t.Errorf("the nested list was not expanded:\n%s", desc)
	}

	// The nested fields must be indented, otherwise a reader cannot tell which bullet belongs to what.
	if !strings.Contains(desc, "\n  - overdue (integer") {
		t.Errorf("nested fields are not indented:\n%s", desc)
	}
}

// TestResultDocTerminatesOnRecursiveTypes makes sure a self-referential aggregate - a group containing
// groups, say - does not render forever.
func TestResultDocTerminatesOnRecursiveTypes(t *testing.T) {
	type node struct {
		Name     string  `json:"name"`
		Children []*node `json:"children,omitempty"`
	}

	done := make(chan string, 1)
	go func() {
		done <- NewTool("tree", "returns a tree", func(echoIn) (node, error) {
			return node{}, nil
		}).WithResultDoc().Def.Description
	}()

	select {
	case desc := <-done:
		if !strings.Contains(desc, "name") {
			t.Errorf("unexpected rendering:\n%s", desc)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("rendering a recursive type did not terminate")
	}
}

// TestNewSliceTool covers the other shape a listing use case comes in. It matters because most nago use cases
// return a slice rather than an iterator, and an unbounded slice handed to a model is the same hazard.
func TestNewSliceTool(t *testing.T) {
	uc := func(_ auth.Subject, _ listFilter) ([]string, error) {
		items := make([]string, DefaultSeqToolLimit+3)
		for i := range items {
			items[i] = fmt.Sprintf("item-%d", i)
		}
		return items, nil
	}

	got := invokeSeq(t, NewSliceTool("list_items", "lists items", uc), `{}`)

	if len(got.Items) != DefaultSeqToolLimit {
		t.Errorf("got %d items, want the default limit of %d", len(got.Items), DefaultSeqToolLimit)
	}

	if !got.Truncated {
		t.Error("a listing that was cut short did not say so")
	}
}

// TestNewSliceToolPropagatesErrors makes sure a failing listing is not reported as an empty one.
func TestNewSliceToolPropagatesErrors(t *testing.T) {
	boom := errors.New("backend unavailable")

	tool := NewSliceTool("list_items", "lists items", func(auth.Subject, listFilter) ([]string, error) {
		return nil, boom
	})

	if _, err := tool.Invoke(testSubject{id: "eve"}, json.RawMessage(`{}`)); !errors.Is(err, boom) {
		t.Errorf("got %v, want the listing error", err)
	}
}
