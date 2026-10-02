// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import "testing"

func TestMailServiceOrigin(t *testing.T) {
	for _, c := range []struct {
		contextPath string
		want        string
	}{
		{"", "http://localhost:3000"},
		{"localhost:3000", "http://localhost:3000"},
		{"http://localhost", "http://localhost"},
		{"https://wokoda.apps.example.com", "https://wokoda.apps.example.com"},
		{"https://wokoda.apps.example.com/", "https://wokoda.apps.example.com"},
	} {
		if got := mailServiceOrigin(c.contextPath, 3000); got != c.want {
			t.Fatalf("%q: expected %q, got %q", c.contextPath, c.want, got)
		}
	}
}
