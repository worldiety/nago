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
		"PRUEF_PW=geheim":             "PRUEF_PW=***",
		"DB_PASS=geheim":              "DB_PASS=***",
		"BASIC_AUTH=a:b":              "BASIC_AUTH=***",
		"DATABASE_DSN=x":              "DATABASE_DSN=***",
		"TLS_CERT_FILE=/a/b":          "TLS_CERT_FILE=***",
		"SESSION_COOKIE=x":            "SESSION_COOKIE=***",
		"HOME=/Users/x":               "HOME=/Users/x",
		"HTTP_PORT=8080":              "HTTP_PORT=8080",
		"DATABASE_URL=postgres://user:geheim@db:5432/app?sslmode=off": "DATABASE_URL=postgres://user:***@db:5432/app?sslmode=off",
		"MYSQL=user:geheim@tcp(db:3306)/app":                          "MYSQL=user:***@tcp(db:3306)/app",
		"UPSTREAM=http://host:8080/path@x":                            "UPSTREAM=http://host:8080/path@x",
		"GIT_REMOTE=git@github.com:worldiety/nago.git":                "GIT_REMOTE=git@github.com:worldiety/nago.git",
	}

	for in, want := range tests {
		if got := maskEnv(in); got != want {
			t.Errorf("maskEnv(%q) = %q, want %q", in, got, want)
		}
	}
}
