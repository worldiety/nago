// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrictnet_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/secret/gostrictnet"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
)

const netProgram = `package p

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/task"
)

func Keys() string {
	var res string
	for k, v := range net.Values("crm").All() {
		res += k + "=" + v + ";"
	}
	return res
}

func Has(key string) bool {
	_, ok := net.Value("crm", key)
	return ok
}

func Call() task.ID {
	r := net.NewTextRequest("GET", "/echo", "")
	tenant, ok := net.Value("crm", "TENANT")
	if !ok {
		return net.Send("crm", r, report)
	}
	r.Query.Set("tenant", tenant)
	return net.Send("crm", r, report)
}

func report(r net.Result) error {
	return errors.New(string(r.Outcome) + "|" + r.Reason + "|" + r.Body.Text())
}
`

// TestProgramUsesKeyValues runs a program against a real server: it builds the values, which it may read, into its
// request, while the credentials reach the server without the program ever seeing them.
func TestProgramUsesKeyValues(t *testing.T) {
	var got http.Header
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, gotQuery = r.Header.Clone(), r.URL.RawQuery
		// a server, which mirrors everything, as an attacker could wish for
		fmt.Fprint(w, r.Header.Get("Authorization"), " ", r.Header.Get("X-Api-Key"), " ", r.URL.RawQuery)
	}))
	t.Cleanup(srv.Close)

	const token, apiKey, query = "t0ken-of-the-user", "k3y-of-the-user", "p4ram-of-the-user"
	kv := secret.KeyValues{Name: "crm", Sections: []secret.KeyValueSection{{Fields: []secret.KeyValueField{
		{Key: secret.KeyBaseURL, Kind: secret.FieldURL, Value: srv.URL},
		{Key: secret.KeyHeaderBearerAuth, Kind: secret.FieldString, Protection: secret.WriteOnly, Value: token},
		{Key: "HEADER_X_API_KEY", Kind: secret.FieldString, Protection: secret.Concealed, Value: apiKey},
		{Key: "QUERY_signature", Kind: secret.FieldString, Protection: secret.WriteOnly, Value: query},
		{Key: "SIGNING_KEY", Kind: secret.FieldMultiLine, Protection: secret.WriteOnly, Value: "never-readable"},
		{Key: "TENANT", Kind: secret.FieldString, Value: "t-7"},
		{Key: "REGION", Kind: secret.FieldString, Protection: secret.Concealed, Value: "eu"},
	}}}}

	prog, err := gostrict.LoadModules("llm.local/p", gostrict.Module{Path: "llm.local/p", FS: fstest.MapFS{"p.go": {Data: []byte(netProgram)}}})
	if err != nil {
		t.Fatal(err)
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	bundle, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}

	ends := make(chan string, 1)
	host := nethost.New(t.Context(), nethost.Options{AllowPrivateNetworks: true})
	var pool *gostrict.Pool
	binding := nethost.Binding{
		Resolve: func(name string) (nethost.Connection, error) {
			if name != kv.Name {
				return nethost.Connection{}, nethost.Refused("unknown")
			}
			return gostrictnet.KeyValuesConnection(kv)
		},
		Complete: func(ctx context.Context, done gostrict.Detached, args []any) error {
			res, err := pool.CallDetached(ctx, done, args...)
			if err != nil {
				ends <- "engine: " + err.Error()
				return err
			}
			err = gostrict.ErrorOf(res[0])
			ends <- err.Error()
			return err
		},
	}
	pool, err = gostrict.NewPoolFunc(t.Context(), &vm.Engine{}, bundle, 1, func(*gostrict.Slot) gostrict.Funcs { return host.Funcs(binding) })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	res, err := pool.Call(t.Context(), "Keys")
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "REGION=eu;TENANT=t-7;" {
		t.Errorf("the program reads %q", res[0])
	}
	for _, key := range []string{secret.KeyBaseURL, secret.KeyHeaderBearerAuth, "HEADER_X_API_KEY", "QUERY_signature", "SIGNING_KEY"} {
		res, err := pool.Call(t.Context(), "Has", key)
		if err != nil {
			t.Fatal(err)
		}
		if res[0] != false {
			t.Errorf("the program can read the key %s", key)
		}
	}

	if _, err := pool.Call(t.Context(), "Call"); err != nil {
		t.Fatal(err)
	}
	var end string
	select {
	case end = <-ends:
	case <-time.After(10 * time.Second):
		t.Fatal("the request did not end")
	}

	if got.Get("Authorization") != "Bearer "+token || got.Get("X-Api-Key") != apiKey {
		t.Errorf("the server got %v", got)
	}
	if !strings.Contains(gotQuery, "signature="+query) || !strings.Contains(gotQuery, "tenant=t-7") {
		t.Errorf("the server got the query %q", gotQuery)
	}
	for _, leak := range []string{token, apiKey, query} {
		if strings.Contains(end, leak) {
			t.Errorf("the program has seen the secret %q in %q", leak, end)
		}
	}
	if !strings.HasPrefix(end, "ok||Bearer *** *** ") || !strings.Contains(end, "signature=***") {
		t.Errorf("the mirrored answer is %q", end)
	}
}
