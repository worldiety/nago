// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ailibrary

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library"
)

// invokeTool runs a tool by name with raw JSON arguments and returns the raw result.
func invokeTool(t *testing.T, name, args string) []byte {
	t.Helper()

	for _, tool := range testTools(t) {
		if tool.Def.Name != name {
			continue
		}

		out, err := tool.Invoke(nil, json.RawMessage(args))
		if err != nil {
			t.Fatalf("invoking %s failed: %v", name, err)
		}

		return out
	}

	t.Fatalf("no %s tool", name)
	return nil
}

// TestKnowledgeToolsAreOffered guards the point of the whole file: without them the model answers "warum ist
// das so" by inventing a rule, which sounds exactly like the system explaining itself.
func TestKnowledgeToolsAreOffered(t *testing.T) {
	var names []string
	for _, tool := range testTools(t) {
		names = append(names, tool.Def.Name)
	}

	for _, want := range []string{"read_decision", "read_capabilities"} {
		if !slices.Contains(names, want) {
			t.Errorf("the assistant cannot explain itself: %q is missing from %v", want, names)
		}
	}
}

// TestKnowledgeToolsDoNotChangeAnything keeps them out of the confirmation gate, where they would ask the
// user to approve reading a paragraph.
func TestKnowledgeToolsDoNotChangeAnything(t *testing.T) {
	for _, tool := range testTools(t) {
		if tool.Def.Name != "read_decision" && tool.Def.Name != "read_capabilities" {
			continue
		}

		if tool.Mutating {
			t.Errorf("tool %q is marked as changing something", tool.Def.Name)
		}
	}
}

// TestReadDecisionReturnsTheCost is the field that decides whether the answer is an account or advocacy. A
// decision that only states its justification lets the assistant defend the system rather than explain it.
func TestReadDecisionReturnsTheCost(t *testing.T) {
	raw := invokeTool(t, "read_decision", `{"id":"DEC-AVAILABILITY-DERIVED"}`)

	var got library.Decision
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("cannot decode: %v", err)
	}

	if got.Rationale == "" {
		t.Error("the decision does not say why")
	}

	if got.Consequences == "" {
		t.Error("the decision does not say what it costs")
	}
}

// TestReadDecisionNamesWhatExists turns a dead end into the next step. A bare "not found" invites the model
// to guess another identifier, and it will.
func TestReadDecisionNamesWhatExists(t *testing.T) {
	for _, tool := range testTools(t) {
		if tool.Def.Name != "read_decision" {
			continue
		}

		_, err := tool.Invoke(nil, json.RawMessage(`{"id":"DEC-NOPE"}`))
		if err == nil {
			t.Fatal("an unknown decision resolved to something")
		}

		if !strings.Contains(err.Error(), "DEC-AVAILABILITY-DERIVED") {
			t.Errorf("the error does not say what does exist: %v", err)
		}

		return
	}

	t.Fatal("no read_decision tool")
}

// TestEveryDecisionIsComplete applies the rule to all of them, not just the one the test above happens to
// read. speclink refuses to build a Decision that lacks either field; this is the hand-maintained stand-in.
func TestEveryDecisionIsComplete(t *testing.T) {
	for _, d := range library.Decisions() {
		if d.ID == "" || d.Title == "" || d.Text == "" {
			t.Errorf("decision %+v is missing its identity or statement", d)
		}

		if d.Rationale == "" {
			t.Errorf("decision %q does not say why", d.ID)
		}

		if d.Consequences == "" {
			t.Errorf("decision %q does not say what it costs", d.ID)
		}
	}
}

// TestCapabilitiesReferenceRealDecisions is the check that catches the drift this arrangement is exposed to.
// A capability naming a decision that no longer exists sends the assistant to read_decision for an
// identifier it will never find, and the user gets "warum" answered with an error.
//
// A project using speclink does not need this test: spec.Satisfies references the requirement by its Go
// identifier, so a deleted requirement is a compile error rather than a dangling string.
func TestCapabilitiesReferenceRealDecisions(t *testing.T) {
	for _, c := range library.Capabilities() {
		for _, id := range c.Decisions {
			if _, ok := library.FindDecision(id); !ok {
				t.Errorf("capability %q rests on %q, which does not exist", c.UseCase, id)
			}
		}
	}
}

// TestCapabilitiesMatchTheOfferedTools is the other half of the same drift: a capability describing a tool
// that was renamed or removed tells the user about something they cannot reach.
func TestCapabilitiesMatchTheOfferedTools(t *testing.T) {
	var offered []string
	for _, tool := range testTools(t) {
		offered = append(offered, tool.Def.Name)
	}

	for _, c := range library.Capabilities() {
		if !slices.Contains(offered, c.Tool) {
			t.Errorf("capability %q names tool %q, which is not offered: %v", c.UseCase, c.Tool, offered)
		}

		if c.Help == "" {
			t.Errorf("capability %q has no explanation for a human", c.UseCase)
		}
	}
}

// TestSystemPromptCarriesTheIndexNotTheText pins down the split that keeps the prompt from growing with the
// project: the model learns what exists from the prompt and fetches the reasoning only when it needs it.
func TestSystemPromptCarriesTheIndexNotTheText(t *testing.T) {
	prompt := SystemPrompt()

	for _, d := range library.Decisions() {
		if !strings.Contains(prompt, d.ID) {
			t.Errorf("the prompt does not mention %q, so the model cannot know it can be looked up", d.ID)
		}

		if strings.Contains(prompt, d.Consequences) {
			t.Errorf("the prompt carries the full text of %q; that belongs behind read_decision", d.ID)
		}
	}
}

// TestSystemPromptDerivesTheIndex catches the copy that would otherwise be made by hand and then forgotten.
func TestSystemPromptDerivesTheIndex(t *testing.T) {
	if !strings.Contains(SystemPrompt(), library.DecisionIndex()) {
		t.Error("the prompt does not use the derived index, so it is a second list that can drift")
	}
}
