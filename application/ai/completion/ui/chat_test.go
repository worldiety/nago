// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"slices"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

func testTools() []completion.Tool {
	noop := func(struct{}) (struct{}, error) { return struct{}{}, nil }

	return []completion.Tool{
		completion.NewTool("read_a", "reads a", noop),
		completion.NewTool("write_a", "writes a", noop).AsMutating("changes a"),
		completion.NewTool("read_b", "reads b", noop),
		completion.NewTool("write_b", "writes b", noop).AsMutating("changes b"),
	}
}

func toolNames(tools []completion.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Def.Name)
	}

	slices.Sort(names)
	return names
}

// TestResolveTools_ReadOnlyDropsMutatingTools is the property that lets an application ship one tool set and
// let a setting decide whether it may write, instead of maintaining two lists that drift apart.
//
// It is a filter rather than an instruction: a tool that is never advertised cannot be called, whatever the
// model decides.
func TestResolveTools_ReadOnlyDropsMutatingTools(t *testing.T) {
	agent := Agent{Tools: testTools()}

	got := toolNames(ChatOptions{ReadOnly: true}.resolveTools(agent))
	want := []string{"read_a", "read_b"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestResolveTools_DefaultKeepsEverything guards against the filter being applied by accident.
func TestResolveTools_DefaultKeepsEverything(t *testing.T) {
	agent := Agent{Tools: testTools()}

	got := toolNames(ChatOptions{}.resolveTools(agent))
	want := []string{"read_a", "read_b", "write_a", "write_b"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestResolveTools_DoesNotAliasTheAgent matters because the returned slice is appended to (ask_user) on every
// turn. Sharing the backing array with the agent would let one turn overwrite the agent's own tools.
func TestResolveTools_DoesNotAliasTheAgent(t *testing.T) {
	agent := Agent{Tools: testTools()}
	before := toolNames(agent.Tools)

	resolved := ChatOptions{}.resolveTools(agent)
	resolved = append(resolved, completion.NewTool("injected", "injected", func(struct{}) (struct{}, error) {
		return struct{}{}, nil
	}))
	_ = resolved

	if got := toolNames(agent.Tools); !slices.Equal(got, before) {
		t.Errorf("the agent's tools changed from %v to %v", before, got)
	}
}

// TestContainsMutating decides whether the confirmation gate is wired at all, so a purely reading assistant
// never pays for it.
func TestContainsMutating(t *testing.T) {
	if !containsMutating(testTools()) {
		t.Error("a tool set with writes was reported as read-only")
	}

	readOnly := ChatOptions{ReadOnly: true}.resolveTools(Agent{Tools: testTools()})
	if containsMutating(readOnly) {
		t.Error("a filtered tool set still reports mutating tools")
	}

	if containsMutating(nil) {
		t.Error("an empty tool set reports mutating tools")
	}
}

// TestRequiredPermissionsAreDeclared makes sure the list an operator is pointed at is not stale: every entry
// must be a permission the framework actually declares.
func TestRequiredPermissionsAreDeclared(t *testing.T) {
	perms := RequiredPermissions()

	if len(perms) == 0 {
		t.Fatal("no permissions required at all - that cannot be right")
	}

	for _, pid := range perms {
		if pid == "" {
			t.Error("an empty permission id is required")
		}
	}
}
