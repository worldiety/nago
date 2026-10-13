// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

func TestCleanTitle(t *testing.T) {
	tests := map[string]string{
		"Rechnung prüfen":                "Rechnung prüfen",
		"„Rechnung prüfen“.":             "Rechnung prüfen",
		"Titel: \"Rechnung prüfen\"":     "Rechnung prüfen",
		"## **Rechnung prüfen**\nweiter": "Rechnung prüfen",
		"\n\n  Urlaub planen  \n":        "Urlaub planen",
		"":                               "",
		strings.Repeat("Wort ", 30):      strings.TrimSpace(strings.Repeat("Wort ", 12)) + "…",
	}

	for in, want := range tests {
		if got := cleanTitle(in); got != want {
			t.Errorf("cleanTitle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMeaningful(t *testing.T) {
	user := func(s string) completion.Message {
		return completion.Message{Role: completion.User, Content: []completion.Content{completion.Text{Text: s}}}
	}
	answer := completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "ok"}}}

	if meaningful([]completion.Message{user("Hallo"), answer}) {
		t.Fatal("a greeting has no topic")
	}

	if meaningful([]completion.Message{user("Bitte prüfe die Rechnung")}) {
		t.Fatal("a conversation without an answer is not summarized")
	}

	if !meaningful([]completion.Message{user("Bitte prüfe die Rechnung"), answer}) {
		t.Fatal("expected a topic")
	}
}
