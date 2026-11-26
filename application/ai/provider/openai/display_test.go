// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"encoding/json"
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

// TestBuildRequest_SendsTextNotDisplay: the provider reads only Text; the text for the user never reaches the model.
func TestBuildRequest_SendsTextNotDisplay(t *testing.T) {
	p := newProvider("id", Settings{BaseURL: "http://localhost:11434/v1"})
	req, err := p.buildRequest(completion.Options{Model: "m", Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{
		completion.Text{Text: "Bitte lesen"}, completion.Text{Text: "the sentence for the model", Display: "📎 a.pdf"},
	}}}})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(raw), "the sentence for the model") || !strings.Contains(string(raw), "Bitte lesen") || strings.Contains(string(raw), "📎") {
		t.Errorf("unexpected request %s", raw)
	}
}
