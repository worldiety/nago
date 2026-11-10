// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret_test

import (
	"testing"

	"go.wdy.de/nago/application/ai/provider/anthropic"
	"go.wdy.de/nago/application/ai/provider/gollama"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/chatbot/provider/mattermost"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/sms/provider/spryng"
)

// Every type with a write-only value and a configurable server must declare its address, otherwise an editor can
// redirect the stored token. The others have a fixed server.
func TestWriteOnlyTypesDeclareTheirAddress(t *testing.T) {
	tests := []struct {
		name        string
		credentials secret.Credentials
		want        bool
	}{
		{"smtp", secret.SMTP{}, true},
		{"openai", openai.Settings{}, true},
		{"mattermost", mattermost.Settings{}, true},
		{"anthropic", anthropic.Settings{}, false},
		{"gollama", gollama.Settings{}, false},
		{"spryng", spryng.Settings{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := secret.HasAddressFields(tt.credentials); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
