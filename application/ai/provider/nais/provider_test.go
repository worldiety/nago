// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nais_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider/nais"
	"go.wdy.de/nago/application/user"
	datamem "go.wdy.de/nago/pkg/data/mem"
)

const origin = "https://wokoda.apps.example.com"

// fakeService answers like the Nago AI Service: the token handling and the Messages API with its models and files.
type fakeService struct {
	t      *testing.T
	nonces *nais.Nonces

	mutex     sync.Mutex
	exchanges int
	refreshes int
	refresh   string
	access    string
	// revokeNext answers the next model call with token_revoked
	revokeNext bool
	keys       []string
	uploads    int
}

func (f *fakeService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	fail := func(status int, code string) {
		w.WriteHeader(status)
		_, _ = fmt.Fprintf(w, `{"type":"error","error":{"type":"authentication_error","code":%q,"message":"no"}}`, code)
	}

	switch r.URL.Path {
	case "/v1/exchange":
		var body struct{ Origin, Nonce string }
		_ = json.NewDecoder(r.Body).Decode(&body)

		// the call back of the origin, like the real service does over https
		rec := httptest.NewRecorder()
		f.nonces.Handler()(rec, httptest.NewRequest(http.MethodGet, body.Origin+nais.NoncePath+body.Nonce, nil))
		if body.Origin != origin || rec.Code != http.StatusOK {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":"origin_unverified","message":"no"}}`))
			return
		}

		f.exchanges++
		f.refresh = fmt.Sprintf("refresh-%d", f.exchanges)
		_, _ = fmt.Fprintf(w, `{"tenant":"hub","refresh":%q,"models":[{"id":"sonnet","name":"Claude Sonnet"}],"defaultModel":"sonnet","limits":{"tokensPerDay":1000}}`, f.refresh)
	case "/v1/token":
		var body struct{ Refresh string }
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Refresh != f.refresh {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"refresh_invalid","message":"no"}}`))
			return
		}

		f.refreshes++
		f.refresh = fmt.Sprintf("refresh-%d-%d", f.exchanges, f.refreshes)
		f.access = fmt.Sprintf("nai_%d", f.refreshes)
		_, _ = fmt.Fprintf(w, `{"access":%q,"expiresIn":3600,"refresh":%q}`, f.access, f.refresh)
	default:
		key := r.Header.Get("x-api-key")
		f.keys = append(f.keys, key)
		if key != f.access {
			fail(http.StatusUnauthorized, "unauthorized")
			return
		}

		if f.revokeNext {
			f.revokeNext = false
			f.refresh, f.access = "", ""
			fail(http.StatusUnauthorized, "token_revoked")
			return
		}

		switch r.URL.Path {
		case "/v1/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"sonnet","type":"model","display_name":"Claude Sonnet","created_at":"1970-01-01T00:00:00Z","default":true}],"has_more":false}`))
		case "/v1/messages":
			_, _ = w.Write([]byte(`{"id":"msg_1","type":"message","role":"assistant","model":"sonnet","content":[{"type":"text","text":"Moin"}],"stop_reason":"end_turn","usage":{"input_tokens":3,"output_tokens":1}}`))
		case "/v1/files":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			f.uploads++
			_, _ = w.Write([]byte(`{"id":"file_1","type":"file","filename":"a.txt","mime_type":"text/plain","size_bytes":5,"created_at":"2026-10-08T08:00:00Z"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func newWorld(t *testing.T) (*fakeService, *nais.Service, *httptest.Server) {
	f := &fakeService{t: t, nonces: nais.NewNonces()}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)

	svc := nais.NewService(nais.Options{
		Endpoint: srv.URL,
		Origin:   func() string { return origin },
		Nonces:   f.nonces,
		States:   &datamem.Repository[nais.State, string]{},
	})

	return f, svc, srv
}

func TestProviderEnrollsByItself(t *testing.T) {
	f, svc, _ := newWorld(t)
	prov := nais.NewProvider(svc, nil)
	subject := user.SU()

	if prov.Identity() != nais.ID || prov.Name() != "Nago AI Service" {
		t.Fatalf("unexpected provider %s %s", prov.Identity(), prov.Name())
	}

	// listing the models enrolls the instance, without any secret
	for m, err := range prov.Completions().Unwrap().Models(subject) {
		if err != nil || m.ID != "sonnet" {
			t.Fatalf("unexpected model %+v %v", m, err)
		}
	}

	status := svc.Status()
	if !status.Enrolled || status.Tenant != "hub" || status.DefaultModel != "sonnet" || f.exchanges != 1 {
		t.Fatalf("unexpected status %+v", status)
	}

	res, err := prov.Completions().Unwrap().Complete(context.Background(), subject, completion.Options{
		Model:    "sonnet",
		Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Hallo"}}}},
	})
	if err != nil || res.Usage.InputTokens != 3 {
		t.Fatalf("unexpected result %+v %v", res, err)
	}

	// the access token is reused, not renewed per call
	if f.refreshes != 1 {
		t.Fatalf("expected one refresh, got %d", f.refreshes)
	}

	// files go through the service as well
	up, err := prov.Files().Unwrap().Put(subject, file.CreateOptions{Name: "a.txt", MimeType: "text/plain", Open: func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("Hallo")), nil
	}})
	if err != nil || up.ID != "file_1" || f.uploads != 1 {
		t.Fatalf("unexpected upload %+v %v", up, err)
	}
}

func TestRevokedTokenEnrollsAgain(t *testing.T) {
	f, svc, _ := newWorld(t)
	prov := nais.NewProvider(svc, nil)

	call := func() error {
		_, err := prov.Completions().Unwrap().Complete(context.Background(), user.SU(), completion.Options{
			Model:    "sonnet",
			Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Hallo"}}}},
		})
		return err
	}

	if err := call(); err != nil {
		t.Fatal(err)
	}

	// the service revokes the token: the call is repeated after a new enrollment and succeeds
	f.revokeNext = true
	if err := call(); err != nil {
		t.Fatalf("the call must succeed after enrolling again: %v", err)
	}

	if f.exchanges != 2 {
		t.Fatalf("expected a second exchange, got %d", f.exchanges)
	}

	// an expired access token is renewed with the refresh token, without enrolling again
	f.mutex.Lock()
	f.access = "nai_expired_elsewhere"
	f.mutex.Unlock()
	if err := call(); err != nil {
		t.Fatalf("the call must succeed with a new access token: %v", err)
	}

	if f.exchanges != 2 || f.refreshes != 3 {
		t.Fatalf("expected a refresh only, got %d exchanges and %d refreshes", f.exchanges, f.refreshes)
	}
}

func TestDisabledWithoutEndpoint(t *testing.T) {
	svc := nais.NewService(nais.Options{})
	if svc.Enabled() {
		t.Fatal("without endpoint the service is disabled")
	}

	if _, err := svc.Access(context.Background()); err == nil {
		t.Fatal("a disabled service hands out no token")
	}
}
