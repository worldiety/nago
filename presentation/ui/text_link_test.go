// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "testing"

func TestLinkHref(t *testing.T) {
	tests := []struct {
		name   string
		href   string
		target string
		want   string
	}{
		{name: "external", href: "https://www.nago-docs.com/a?b=c", target: "_blank", want: "https://www.nago-docs.com/a?b=c"},
		{name: "mailto", href: "mailto:info@worldiety.de", want: "mailto:info@worldiety.de"},
		{name: "root view", href: "/admin/users", want: "/admin/users"},
		{name: "query keeps the path", href: "/admin/user?id=1234", target: "_self", want: "/admin/user?id=1234"},
		{name: "query is escaped and sorted", href: "/search?q=a+%26+b&a=1", want: "/search?a=1&q=a+%26+b"},
		{name: "first value of a key wins", href: "/x?a=1&a=2", want: "/x?a=1"},
		{name: "fragment is kept", href: "/docs?page=2#install", want: "/docs?page=2#install"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Link(nil, "text", tt.href, tt.target)
			if got.url != tt.want {
				t.Errorf("href: got %q, want %q", got.url, tt.want)
			}

			if got.target != tt.target {
				t.Errorf("target: got %q, want %q", got.target, tt.target)
			}
		})
	}
}
