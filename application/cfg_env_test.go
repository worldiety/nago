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
		"INITIAL_USER_EMAIL=a@b.de":   "INITIAL_USER_EMAIL=***",
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

func TestLoadConfigFromEnvContextPath(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		hubURL   string
		want     string
	}{
		{name: "nothing set", want: ""},
		{name: "hostname", hostname: "my-app.example.com", want: "https://my-app.example.com"},
		{name: "hub public url", hubURL: "https://my-app.example.com/", want: "https://my-app.example.com"},
		{name: "hub public url wins over hostname", hostname: "a1b2c3d4e5f6", hubURL: "https://my-app.example.com", want: "https://my-app.example.com"},
		{name: "blank hub public url", hostname: "my-app.example.com", hubURL: " ", want: "https://my-app.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOSTNAME", tt.hostname)
			t.Setenv(EnvHubPublicURL, tt.hubURL)

			cfg := NewConfigurator()
			cfg.LoadConfigFromEnv()
			if got := cfg.ContextPath(); got != tt.want {
				t.Errorf("ContextPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
