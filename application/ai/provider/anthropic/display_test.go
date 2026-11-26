// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"encoding/json"
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

// TestTextBlock_SendsTextNotDisplay: the provider reads only Text; the text for the user never reaches the model.
func TestTextBlock_SendsTextNotDisplay(t *testing.T) {
	for _, c := range []completion.Text{{Text: "the sentence for the model", Display: "📎 a.pdf"}, {Text: "the sentence for the model"}} {
		ac, err := toAPIContent(c)
		if err != nil {
			t.Fatal(err)
		}

		raw, err := json.Marshal(ac)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(raw), "the sentence for the model") || strings.Contains(string(raw), "📎") || strings.Contains(string(raw), "display") {
			t.Errorf("unexpected block %s", raw)
		}
	}
}
