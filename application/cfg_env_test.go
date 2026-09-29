// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import "testing"

func TestMaskEnv(t *testing.T) {
	tests := map[string]string{
		"PATH=/usr/bin":               "PATH=/usr/bin",
		"API_TOKEN=abc":               "API_TOKEN=***",
		"GitLab_Access_Token=abc=def": "GitLab_Access_Token=***",
		"CLIENT_SECRET=abc":           "CLIENT_SECRET=***",
		"secretfile=x":                "secretfile=***",
		"EMPTY_TOKEN=":                "EMPTY_TOKEN=",
		"NO_SEPARATOR_TOKEN":          "NO_SEPARATOR_TOKEN",
		"TOKENIZER_URL=https://a.b/c": "TOKENIZER_URL=***",
	}

	for in, want := range tests {
		if got := maskEnv(in); got != want {
			t.Errorf("maskEnv(%q) = %q, want %q", in, got, want)
		}
	}
}
