// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package push

import "testing"

// TestVAPIDSubject checks the contact of the VAPID token. The push service of Apple has answered BadJwtToken for
// http://localhost:3000 and https://localhost:3000 and has accepted mailto: and no contact at all.
func TestVAPIDSubject(t *testing.T) {
	for _, tt := range []struct {
		configured, origin, want string
	}{
		{"mailto:ops@example.com", "http://localhost:3000", "mailto:ops@example.com"},
		{"", "https://app.example.com", "https://app.example.com"},
		{"", "https://app.example.com:8443", "https://app.example.com:8443"},
		{"", "http://app.example.com", ""},
		{"", "http://localhost:3000", ""},
		{"", "https://localhost:3000", ""},
		{"", "https://LOCALHOST.", ""},
		{"", "https://app.localhost", ""},
		{"", "https://printer.local", ""},
		{"", "https://127.0.0.1:3000", ""},
		{"", "https://[::1]:3000", ""},
		{"", "https://192.168.1.10", ""},
		{"", "", ""},
		{"", "not a url", ""},
	} {
		if got := vapidSubject(tt.configured, tt.origin); got != tt.want {
			t.Errorf("vapidSubject(%q, %q) = %q, want %q", tt.configured, tt.origin, got, tt.want)
		}
	}
}
