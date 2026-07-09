// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package file

import "testing"

func TestIsText(t *testing.T) {
	cases := []struct {
		mime Type
		want bool
	}{
		{Text, true},
		{Markdown, true},
		{CSV, true},
		{JSON, true},
		{XML, true},
		{"text/html", true},
		{"text/x-go", true},
		{"application/vnd.api+json", true},
		{"image/svg+xml", true},
		{"application/x-yaml", true},
		{"application/toml", true},
		{"TEXT/PLAIN", true}, // case-insensitive
		{" text/plain ", true},

		{PNG, false},
		{JPEG, false},
		{GIF, false},
		{PDF, false},
		{DOCX, false},
		{Binary, false},
		{"application/zip", false},
		{"", false},
	}

	for _, c := range cases {
		if got := IsText(c.mime); got != c.want {
			t.Errorf("IsText(%q) = %v, want %v", c.mime, got, c.want)
		}
	}
}

func TestExtForTextTypes(t *testing.T) {
	cases := map[Type]string{
		Text:     ".txt",
		Markdown: ".md",
		CSV:      ".csv",
		JSON:     ".json",
		XML:      ".xml",
		PNG:      ".png",
		PDF:      ".pdf",
		Binary:   "",
	}

	for mime, want := range cases {
		if got := mime.Ext(); got != want {
			t.Errorf("%q.Ext() = %q, want %q", mime, got, want)
		}
	}
}
