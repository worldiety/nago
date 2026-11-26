// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

// TestPreview_Display: the preview and the label of a session show the text for the user of the first user message, not
// the text for the model; a text without it shows its text, shortened as before.
func TestPreview_Display(t *testing.T) {
	const sentence = "Großes PDF abgelegt: attachment=a1, file=handbuch.pdf, 120 Seiten, nicht an den Anbieter gegeben"
	s := Session{ID: "s1", Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{
		completion.Text{Text: sentence, Display: "📎 handbuch.pdf"},
	}}}}

	if got := s.Preview(); got != "📎 handbuch.pdf" {
		t.Fatalf("Preview = %q", got)
	}

	if got := s.String(); got != "📎 handbuch.pdf" {
		t.Fatalf("String = %q", got)
	}

	s.Messages[0].Content[0] = completion.Text{Text: sentence}
	if got, want := s.Preview(), string([]rune(sentence)[:48])+"…"; got != want {
		t.Fatalf("Preview without Display = %q, want %q", got, want)
	}
}
