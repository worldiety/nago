// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// TestTextDisplay: the text for the user is kept with the message in both directions, and a text without it has no
// such key, so that every stored session keeps its form. Shown is Display if set, else Text.
func TestTextDisplay(t *testing.T) {
	const sentence = "Großes PDF abgelegt: attachment=a1, file=handbuch.pdf"
	history := []Message{{Role: User, Content: []Content{
		Text{Text: "Bitte lesen"},
		Text{Text: sentence, Display: "📎 handbuch.pdf"},
	}}}

	raw, err := json.Marshal(history)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(string(raw), `"display"`) != 1 || !strings.Contains(string(raw), `"display":"📎 handbuch.pdf"`) {
		t.Fatalf("expected exactly one display key, got %s", raw)
	}

	var got []Message
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(history, got) {
		t.Fatalf("roundtrip mismatch:\nwant %#v\n got %#v", history, got)
	}

	if s := got[0].Content[1].(Text).Shown(); s != "📎 handbuch.pdf" {
		t.Fatalf("Shown with Display = %q", s)
	}

	if s := got[0].Content[0].(Text).Shown(); s != "Bitte lesen" {
		t.Fatalf("Shown without Display = %q", s)
	}

	// the counterpart: a stored text of an earlier version has no key and reads back without Display
	var old []Message
	if err := json.Unmarshal([]byte(`[{"role":"user","content":[{"type":"text","text":"`+sentence+`"}]}]`), &old); err != nil {
		t.Fatal(err)
	}

	if tx := old[0].Content[0].(Text); tx.Display != "" || tx.Shown() != sentence {
		t.Fatalf("an earlier text must stay as it was, got %#v", tx)
	}
}
