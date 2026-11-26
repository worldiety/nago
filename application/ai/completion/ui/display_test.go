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

// displaySentence is a text, which an application gives the model instead of an attachment, with many words.
const displaySentence = "Großes PDF abgelegt: attachment=a1, file=handbuch.pdf, 120 Seiten, nicht an den Anbieter gegeben; lies es mit pdf.Read"

// TestDisplay_Bubble: the bubble shows the text for the user instead of the text for the model, beside the typed text;
// a text without it shows its text as before.
func TestDisplay_Bubble(t *testing.T) {
	got := joinTexts([]completion.Text{{Text: "Bitte lesen"}, {Text: displaySentence, Display: "📎 handbuch.pdf"}})
	if got != "Bitte lesen\n\n📎 handbuch.pdf" {
		t.Fatalf("unexpected bubble %q", got)
	}

	if got := joinTexts([]completion.Text{{Text: displaySentence}}); got != displaySentence {
		t.Fatalf("a text without Display must stay, got %q", got)
	}

	h := []completion.Message{{Role: completion.User, Content: []completion.Content{
		completion.Text{Text: "Bitte lesen"}, completion.Text{Text: displaySentence, Display: "📎 handbuch.pdf"},
	}}}
	if views := renderHistory(nil, h, historyView{}); len(views) != 1 {
		t.Fatalf("expected one bubble, got %d views", len(views))
	}
}

// TestDisplay_Title: a text with Display counts as an attachment, not by its words, and the transcript for the title
// carries the text for the user, never the one for the model.
func TestDisplay_Title(t *testing.T) {
	answer := completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "Das Handbuch beschreibt …"}}}
	attached := completion.Message{Role: completion.User, Content: []completion.Content{completion.Text{Text: displaySentence, Display: "📎 handbuch.pdf"}}}

	if !meaningful([]completion.Message{attached, answer}) {
		t.Fatal("an attachment without typed text has a topic")
	}

	if meaningful([]completion.Message{attached}) {
		t.Fatal("a conversation without an answer is not summarized")
	}

	// the counterpart of „not by its words“: the same sentence typed counts its words, a short one has no topic
	short := completion.Message{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Hallo"}}}
	shortShown := completion.Message{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Hallo", Display: "📎 a.pdf"}}}
	if meaningful([]completion.Message{short, answer}) || !meaningful([]completion.Message{shortShown, answer}) {
		t.Fatal("Display must count as an attachment, independent of the words of Text")
	}

	tr := transcript([]completion.Message{attached, answer})
	if !strings.Contains(tr, "Nutzer: 📎 handbuch.pdf") || strings.Contains(tr, "attachment=a1") {
		t.Fatalf("unexpected transcript %q", tr)
	}

	plain := completion.Message{Role: completion.User, Content: []completion.Content{completion.Text{Text: displaySentence}}}
	if tr := transcript([]completion.Message{plain, answer}); !strings.Contains(tr, "attachment=a1") {
		t.Fatalf("a text without Display stays in the transcript, got %q", tr)
	}
}
