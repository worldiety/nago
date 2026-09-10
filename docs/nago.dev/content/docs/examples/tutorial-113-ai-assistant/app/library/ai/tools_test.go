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

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library"
	"go.wdy.de/nago/pkg/blob/mem"
	jsonrepo "go.wdy.de/nago/pkg/data/json"
)

func testTools(t *testing.T) []completion.Tool {
	t.Helper()

	repo := library.BookRepository(jsonrepo.NewSloppyJSONRepository[library.Book, library.BookID](mem.NewBlobStore("books")))
	if err := library.Seed(repo); err != nil {
		t.Fatalf("cannot seed: %v", err)
	}

	return Tools(library.NewUseCases(repo))
}

// schemaOf decodes the advertised input schema of a named tool.
func schemaOf(t *testing.T, name string) map[string]any {
	t.Helper()

	for _, tool := range testTools(t) {
		if tool.Def.Name != name {
			continue
		}

		var schema map[string]any
		if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
			t.Fatalf("cannot decode schema of %q: %v", name, err)
		}

		return schema
	}

	t.Fatalf("no %s tool", name)
	return nil
}

// TestMutatingToolsAreMarked is the check that matters most when tools are written or extended by a language
// model: a tool that changes something must say so structurally, or the confirmation gate silently does not
// apply to it.
//
// The expectation is spelled out as literals on purpose. A test that derives its expectation from the code it
// checks agrees with every change, including the one that forgets the mark.
func TestMutatingToolsAreMarked(t *testing.T) {
	var mutating, reading []string
	for _, tool := range testTools(t) {
		if tool.Mutating {
			mutating = append(mutating, tool.Def.Name)
		} else {
			reading = append(reading, tool.Def.Name)
		}
	}

	slices.Sort(mutating)
	slices.Sort(reading)

	if want := []string{"lend_book", "return_book"}; !slices.Equal(mutating, want) {
		t.Errorf("mutating tools are %v, want %v", mutating, want)
	}

	if want := []string{"list_books"}; !slices.Equal(reading, want) {
		t.Errorf("read-only tools are %v, want %v", reading, want)
	}
}

// TestMutatingToolsExplainThemselves guards the text the user is shown before approving. An empty one turns
// the confirmation dialog into "run lend_book? yes/no", which nobody can judge.
func TestMutatingToolsExplainThemselves(t *testing.T) {
	for _, tool := range testTools(t) {
		if tool.Mutating && tool.Confirm == "" {
			t.Errorf("tool %q changes something but does not say what", tool.Def.Name)
		}
	}
}

// TestListingToolIsBounded makes sure the listing went through NewSeqTool rather than being hand-rolled: the
// limit in the advertised schema is what stops an unbounded list from exhausting the context window.
func TestListingToolIsBounded(t *testing.T) {
	props, ok := schemaOf(t, "list_books")["properties"].(map[string]any)
	if !ok {
		t.Fatal("the listing schema has no properties")
	}

	for _, want := range []string{"query", "onlyAvailable", "borrower", "limit"} {
		if _, ok := props[want]; !ok {
			t.Errorf("the listing schema is missing %q: %v", want, props)
		}
	}
}

// TestFilterFieldsAreOptional is the distinction between a filter and a request made checkable. A listing
// whose filter fields are advertised as mandatory forces the model to invent values for all of them on every
// call - and it will, because the schema told it to.
func TestFilterFieldsAreOptional(t *testing.T) {
	if required, ok := schemaOf(t, "list_books")["required"]; ok {
		t.Errorf("the listing advertises mandatory arguments: %v", required)
	}
}

// TestLendRequestFieldsAreMandatory is the other half: a command must state what it cannot work without, or
// the model discovers it through a runtime error instead.
func TestLendRequestFieldsAreMandatory(t *testing.T) {
	raw, _ := schemaOf(t, "lend_book")["required"].([]any)

	var got []string
	for _, v := range raw {
		got = append(got, v.(string))
	}
	slices.Sort(got)

	if want := []string{"book", "borrower"}; !slices.Equal(got, want) {
		t.Errorf("got required %v, want %v", got, want)
	}
}

// TestListingDocumentsItsTruncation guards the one thing about the listing's result the model cannot infer
// from the data: that a complete-looking list may not be complete.
func TestListingDocumentsItsTruncation(t *testing.T) {
	for _, tool := range testTools(t) {
		if tool.Def.Name != "list_books" {
			continue
		}

		if !strings.Contains(tool.Def.Description, "truncated") {
			t.Errorf("the listing does not tell the model about truncation:\n%s", tool.Def.Description)
		}

		return
	}

	t.Fatal("no list_books tool")
}
