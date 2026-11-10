// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrictnet_test

import (
	"iter"
	"maps"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"testing"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/secret/gostrictnet"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func kv(name string, fields ...secret.KeyValueField) secret.KeyValues {
	return secret.KeyValues{Name: name, Sections: []secret.KeyValueSection{{ID: "s", Fields: fields}}}
}

func field(key string, protection secret.FieldProtection, value string) secret.KeyValueField {
	kind := secret.FieldString
	if key == secret.KeyBaseURL {
		kind = secret.FieldURL
	}
	return secret.KeyValueField{ID: key, Key: key, Kind: kind, Protection: protection, Value: value}
}

func systemSecrets(stored ...secret.Credentials) secret.FindGroupSecrets {
	return func(subject auth.Subject, gid group.ID) iter.Seq2[secret.Secret, error] {
		return func(yield func(secret.Secret, error) bool) {
			for _, c := range stored {
				if !yield(secret.Secret{Credentials: c}, nil) {
					return
				}
			}
		}
	}
}

func TestResolverNeedsTheExactName(t *testing.T) {
	stored := systemSecrets(
		kv("new", field(secret.KeyBaseURL, secret.Visible, "https://new.example/v1")),
		secret.SMTP{Name: "mail"},
		kv("doubled"),
		kv("doubled"),
	)

	var gotGroup group.ID
	resolve := gostrictnet.Resolver(user.SU, func(subject auth.Subject, gid group.ID) iter.Seq2[secret.Secret, error] {
		gotGroup = gid
		return stored(subject, gid)
	})

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"new", "https://new.example/v1", false},
		{"New", "", true},
		{"mail", "", true},
		{"doubled", "", true},
		{"nope", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := resolve(tt.name)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("%q must not resolve", tt.name)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if conn.BaseURL.String() != tt.address || gotGroup != group.System {
				t.Errorf("unexpected address %v of group %v", conn.BaseURL, gotGroup)
			}
		})
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// request sends a request of a program through the decorator and returns what reaches the server.
func request(t *testing.T, c secret.KeyValues, target string, header http.Header) *http.Request {
	t.Helper()
	conn, err := gostrictnet.KeyValuesConnection(c)
	if err != nil {
		t.Fatal(err)
	}
	if conn.Decorate == nil {
		t.Fatal("expected a decorator")
	}
	var got *http.Request
	rt := conn.Decorate(roundTripper(func(r *http.Request) (*http.Response, error) {
		got = r
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}))
	req, _ := http.NewRequest("GET", target, nil)
	maps.Copy(req.Header, header)
	urlBefore, headerBefore := req.URL.String(), req.Header.Clone()
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != urlBefore || !reflect.DeepEqual(req.Header, headerBefore) {
		t.Error("a decorator must not change the request of its caller")
	}
	return got
}

func TestKeyValuesDecoratorSetsHeaders(t *testing.T) {
	tests := []struct {
		name   string
		fields []secret.KeyValueField
		header string
		want   string
	}{
		{"bearer", []secret.KeyValueField{field(secret.KeyHeaderBearerAuth, secret.WriteOnly, "t")}, "Authorization", "Bearer t"},
		{"basic", []secret.KeyValueField{field(secret.KeyHeaderBasicAuthUser, secret.Visible, "u"), field(secret.KeyHeaderBasicAuthPassword, secret.WriteOnly, "t")}, "Authorization", "Basic dTp0"},
		{"basic without password", []secret.KeyValueField{field(secret.KeyHeaderBasicAuthUser, secret.Visible, "u")}, "Authorization", "Basic dTo="},
		{"authorization", []secret.KeyValueField{field("HEADER_AUTHORIZATION", secret.WriteOnly, "Token x")}, "Authorization", "Token x"},
		{"header", []secret.KeyValueField{field("HEADER_X_API_KEY", secret.WriteOnly, "k")}, "X-Api-Key", "k"},
		{"visible header", []secret.KeyValueField{field("HEADER_ACCEPT", secret.Visible, "application/json")}, "Accept", "application/json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request(t, kv("a", tt.fields...), "https://x.example/p", nil)
			if got.Header.Get(tt.header) != tt.want {
				t.Errorf("got %q", got.Header.Get(tt.header))
			}
		})
	}
}

func TestKeyValuesDecoratorCanNotBeOverwritten(t *testing.T) {
	c := kv("a",
		field(secret.KeyHeaderBearerAuth, secret.WriteOnly, "t"),
		field("HEADER_X_KEY", secret.Concealed, "k"),
		field("QUERY_api_key", secret.WriteOnly, "a b&c"),
		field("QUERY_v", secret.Visible, "1"),
	)
	got := request(t, c, "https://x.example/p?api_key=evil&page=2&api_key=evil2", http.Header{
		"Authorization": {"Bearer evil"},
		"X-Key":         {"evil"},
	})
	if got.Header.Get("Authorization") != "Bearer t" || got.Header.Get("X-Key") != "k" {
		t.Errorf("unexpected header %v", got.Header)
	}
	want := url.Values{"api_key": {"a b&c"}, "v": {"1"}, "page": {"2"}}
	if query := got.URL.Query(); !reflect.DeepEqual(query, want) {
		t.Errorf("unexpected query %v", query)
	}
}

// After a change of the address a write-only value is gone, and then nothing may be sent for it.
func TestKeyValuesDecoratorSkipsEmptyValues(t *testing.T) {
	conn, err := gostrictnet.KeyValuesConnection(kv("a",
		field(secret.KeyBaseURL, secret.Visible, "https://x.example"),
		field(secret.KeyHeaderBearerAuth, secret.WriteOnly, ""),
		field(secret.KeyHeaderBasicAuthUser, secret.Visible, ""),
		field(secret.KeyHeaderBasicAuthPassword, secret.WriteOnly, ""),
		field("HEADER_X_KEY", secret.WriteOnly, ""),
		field("QUERY_a", secret.WriteOnly, ""),
	))
	if err != nil {
		t.Fatal(err)
	}
	if conn.Decorate != nil || len(conn.Secrets) != 0 {
		t.Errorf("expected nothing to be sent or scrubbed, got %v", conn.Secrets)
	}
}

func TestKeyValuesSecrets(t *testing.T) {
	tests := []struct {
		name   string
		fields []secret.KeyValueField
		want   []string
	}{
		{"write-only of a plain key", []secret.KeyValueField{field("token", secret.WriteOnly, "w")}, []string{"w"}},
		{"bearer is scrubbed even if visible", []secret.KeyValueField{field(secret.KeyHeaderBearerAuth, secret.Visible, "b")}, []string{"b"}},
		{
			"basic password and base64, not the visible user",
			[]secret.KeyValueField{field(secret.KeyHeaderBasicAuthUser, secret.Visible, "u"), field(secret.KeyHeaderBasicAuthPassword, secret.Visible, "t")},
			[]string{"dTp0", "t"},
		},
		{
			"a concealed user",
			[]secret.KeyValueField{field(secret.KeyHeaderBasicAuthUser, secret.Concealed, "u"), field(secret.KeyHeaderBasicAuthPassword, secret.WriteOnly, "t")},
			[]string{"dTp0", "t", "u"},
		},
		{"concealed header", []secret.KeyValueField{field("HEADER_X_KEY", secret.Concealed, "c")}, []string{"c"}},
		{"concealed query", []secret.KeyValueField{field("QUERY_a", secret.Concealed, "c")}, []string{"c"}},
		{"the sent form of a query", []secret.KeyValueField{field("QUERY_a", secret.WriteOnly, "a+b")}, []string{"a%2Bb", "a+b"}},
		{"not a visible header", []secret.KeyValueField{field("HEADER_ACCEPT", secret.Visible, "application/json")}, nil},
		{"not a visible query", []secret.KeyValueField{field("QUERY_v", secret.Visible, "1")}, nil},
		{"not the address", []secret.KeyValueField{field(secret.KeyBaseURL, secret.Concealed, "https://x.example")}, nil},
		{"not a concealed value of a program", []secret.KeyValueField{field("pin", secret.Concealed, "1234")}, nil},
		{"duplicates once", []secret.KeyValueField{field("a", secret.WriteOnly, "same"), field("b", secret.WriteOnly, "same")}, []string{"same"}},
		{
			"the longest first, so that no rest stays behind",
			[]secret.KeyValueField{field("a", secret.WriteOnly, "abc"), field("b", secret.WriteOnly, "abcdef")},
			[]string{"abcdef", "abc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := gostrictnet.KeyValuesConnection(kv("a", tt.fields...))
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(conn.Secrets, tt.want) {
				t.Errorf("got %q, want %q", conn.Secrets, tt.want)
			}
		})
	}
}

func TestKeyValuesValues(t *testing.T) {
	conn, err := gostrictnet.KeyValuesConnection(kv("a",
		field(secret.KeyBaseURL, secret.Visible, "https://x.example"),
		field(secret.KeyHeaderBearerAuth, secret.Concealed, "b"),
		field("HEADER_ACCEPT", secret.Visible, "text/plain"),
		field("QUERY_v", secret.Visible, "1"),
		field("tenant", secret.Visible, "acme"),
		field("pin", secret.Concealed, "1234"),
		field("token", secret.WriteOnly, "w"),
	))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"tenant": "acme", "pin": "1234"}
	if !maps.Equal(conn.Values, want) {
		t.Errorf("got %v", conn.Values)
	}
}

func TestKeyValuesWithoutAddressOnlyOffersValues(t *testing.T) {
	conn, err := gostrictnet.KeyValuesConnection(kv("a", field("tenant", secret.Visible, "acme")))
	if err != nil {
		t.Fatal(err)
	}
	if conn.BaseURL != nil || conn.Decorate != nil || conn.Values["tenant"] != "acme" {
		t.Errorf("unexpected connection %+v", conn)
	}

	conn, err = gostrictnet.KeyValuesConnection(kv("a", field(secret.KeyBaseURL, secret.Visible, "")))
	if err != nil || conn.BaseURL != nil {
		t.Errorf("an empty address is no address: %v %v", conn.BaseURL, err)
	}
}

// A record, which has not been validated by the use cases, must not reach the transport either.
func TestKeyValuesRefusesInvalidRecords(t *testing.T) {
	for _, f := range []secret.KeyValueField{
		field(secret.KeyBaseURL, secret.Visible, "ftp://x.example"),
		field(secret.KeyBaseURL, secret.Visible, "https://u:p@x.example"),
		field("HEADER_HOST", secret.Visible, "evil.example"),
		field("HEADER_PROXY_AUTHORIZATION", secret.Visible, "x"),
		field("HEADER_", secret.Visible, "x"),
		field("QUERY_", secret.Visible, "x"),
	} {
		if _, err := gostrictnet.KeyValuesConnection(kv("a", f)); err == nil {
			t.Errorf("%s must be refused", f.Key)
		}
	}
}
