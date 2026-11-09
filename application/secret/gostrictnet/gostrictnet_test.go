// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrictnet_test

import (
	"iter"
	"net/http"
	"testing"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/secret/gostrictnet"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func TestResolverNeedsTheExactName(t *testing.T) {
	stored := []secret.Credentials{
		secret.HTTP{Name: "billing", BaseURL: "https://billing.example", Auth: gostrictnet.AuthBearer, Token: "b"},
		secret.SMTP{Name: "crm"},
		secret.HTTP{Name: "crm", BaseURL: "https://crm.example/v1", Auth: gostrictnet.AuthBearer, Token: "c"},
		secret.HTTP{Name: "twice", BaseURL: "https://a.example"},
		secret.HTTP{Name: "twice", BaseURL: "https://b.example"},
	}
	var gotGroup group.ID
	find := func(subject auth.Subject, gid group.ID) iter.Seq2[secret.Secret, error] {
		gotGroup = gid
		return func(yield func(secret.Secret, error) bool) {
			for _, c := range stored {
				if !yield(secret.Secret{Credentials: c}, nil) {
					return
				}
			}
		}
	}
	resolve := gostrictnet.Resolver(user.SU, find)

	conn, err := resolve("crm")
	if err != nil {
		t.Fatal(err)
	}
	if conn.BaseURL.String() != "https://crm.example/v1" || gotGroup != group.System {
		t.Errorf("unexpected connection %v of group %v", conn.BaseURL, gotGroup)
	}
	if _, err := resolve("nope"); err == nil {
		t.Error("an unknown name must not resolve to another connection")
	}
	if _, err := resolve("twice"); err == nil {
		t.Error("an ambiguous name must not resolve")
	}
}

func TestConnectionSetsTheCredentials(t *testing.T) {
	tests := []struct {
		name          string
		c             secret.HTTP
		header, value string
		secrets       int
	}{
		{"none", secret.HTTP{BaseURL: "https://x.example"}, "", "", 0},
		{"bearer", secret.HTTP{BaseURL: "https://x.example", Auth: gostrictnet.AuthBearer, Token: "t"}, "Authorization", "Bearer t", 1},
		{"basic", secret.HTTP{BaseURL: "https://x.example", Auth: gostrictnet.AuthBasic, Username: "u", Token: "t"}, "Authorization", "Basic dTp0", 2},
		{"header", secret.HTTP{BaseURL: "https://x.example", Auth: gostrictnet.AuthHeader, HeaderName: "X-Key", Token: "t"}, "X-Key", "t", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := gostrictnet.Connection(tt.c)
			if err != nil {
				t.Fatal(err)
			}
			if len(conn.Secrets) != tt.secrets {
				t.Errorf("unexpected secrets: %v", conn.Secrets)
			}
			if tt.header == "" {
				if conn.Decorate != nil {
					t.Error("a connection without a login must add nothing")
				}
				return
			}
			var got http.Header
			rt := conn.Decorate(roundTripper(func(r *http.Request) (*http.Response, error) {
				got = r.Header
				return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
			}))
			req, _ := http.NewRequest("GET", "https://x.example", nil)
			if _, err := rt.RoundTrip(req); err != nil {
				t.Fatal(err)
			}
			if got.Get(tt.header) != tt.value {
				t.Errorf("got %q", got.Get(tt.header))
			}
			if req.Header.Get(tt.header) != "" {
				t.Error("a decorator must not change the request of its caller")
			}
		})
	}
}

func TestConnectionRefusesInvalidAddresses(t *testing.T) {
	for _, c := range []secret.HTTP{
		{BaseURL: "ftp://x.example"},
		{BaseURL: "https://"},
		{BaseURL: "https://user:pw@x.example"},
		{BaseURL: "https://x.example", Auth: "magic"},
		{BaseURL: "https://x.example", Auth: gostrictnet.AuthHeader},
	} {
		if _, err := gostrictnet.Connection(c); err == nil {
			t.Errorf("%+v must be refused", c)
		}
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
