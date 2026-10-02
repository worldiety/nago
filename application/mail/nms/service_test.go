// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nms_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/mail/nms"
	"go.wdy.de/nago/application/mail/nms/nmstest"
	datamem "go.wdy.de/nago/pkg/data/mem"
)

const origin = "https://wokoda.apps.example.com"

func message() nms.Message {
	return nms.Message{
		From:     nms.Address{Email: "noreply@wokoda.apps.example.com"},
		To:       []nms.Address{{Email: "torben@example.com"}},
		Subject:  "hello",
		TextPart: "world",
	}
}

// callBack lets the fake service fetch the nonce from the handler of the instance, like the real one does over https.
func callBack(t *testing.T, nonces *nms.Nonces) func(origin, nonce string) bool {
	return func(o, nonce string) bool {
		if o != origin {
			t.Errorf("unexpected origin %q", o)
			return false
		}

		rec := httptest.NewRecorder()
		nonces.Handler()(rec, httptest.NewRequest(http.MethodGet, o+nms.NoncePath+nonce, nil))
		if rec.Code != http.StatusOK {
			return false
		}

		var body struct{ Nonce string }
		return json.NewDecoder(rec.Body).Decode(&body) == nil && body.Nonce == nonce
	}
}

func newService(t *testing.T, srv *nmstest.Server, states nms.StateRepository, token string) *nms.Service {
	nonces := nms.NewNonces()
	srv.VerifyOrigin = callBack(t, nonces)
	return nms.NewService(nms.Options{
		Endpoint: srv.URL,
		Token:    token,
		Origin:   func() string { return origin },
		Nonces:   nonces,
		States:   states,
	})
}

func TestExchangeAndSend(t *testing.T) {
	srv := nmstest.NewServer(t)
	states := &datamem.Repository[nms.State, string]{}
	svc := newService(t, srv, states, "")

	if !svc.Available() {
		t.Fatal("the service must be available for an exchange")
	}

	res, err := svc.Send(context.Background(), message())
	if err != nil || res.Status != nms.StatusQueued {
		t.Fatalf("send failed: %+v %v", res, err)
	}

	if _, err := svc.Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	exchanges, tokens, sends := srv.Calls()
	if exchanges != 1 || tokens != 1 || sends != 2 {
		t.Fatalf("expected one exchange, one token and two sends, got %d %d %d", exchanges, tokens, sends)
	}

	st := option.Must(states.FindByID("default")).Unwrap()
	if st.Tenant != "wokoda.apps.example.com" || st.Refresh == "" || len(st.SenderDomains) != 1 {
		t.Fatalf("enrollment not persisted: %+v", st)
	}

	// a restart reuses the persisted, rotated refresh token
	again := newService(t, srv, states, "")
	if _, err := again.Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	if exchanges, _, _ := srv.Calls(); exchanges != 1 {
		t.Fatalf("the restart enrolled again: %d exchanges", exchanges)
	}
}

func TestRevokedTokensEnrollAgain(t *testing.T) {
	srv := nmstest.NewServer(t)
	svc := newService(t, srv, nil, "")

	if _, err := svc.Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	srv.RevokeAll()

	if _, err := svc.Send(context.Background(), message()); err != nil {
		t.Fatalf("the revocation was not recovered: %v", err)
	}

	if exchanges, _, _ := srv.Calls(); exchanges != 2 {
		t.Fatalf("expected a second exchange, got %d", exchanges)
	}

	if len(srv.Sent()) != 2 {
		t.Fatalf("expected two messages, got %d", len(srv.Sent()))
	}
}

func TestConfiguredTokenIsTakenOverOnce(t *testing.T) {
	srv := nmstest.NewServer(t)
	token := srv.IssueRefresh("custom", "wokoda.apps.example.com")
	states := &datamem.Repository[nms.State, string]{}
	svc := newService(t, srv, states, token)

	if _, err := svc.Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	if exchanges, _, _ := srv.Calls(); exchanges != 0 {
		t.Fatal("a configured token must not be exchanged")
	}

	// after a restart, the rotated token is used and not the configured one, which is no longer valid
	again := newService(t, srv, states, token)
	if _, err := again.Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	if exchanges, _, _ := srv.Calls(); exchanges != 0 {
		t.Fatalf("the configured token has been taken over again: %d exchanges", exchanges)
	}
}

func TestNoExchangeWithoutOrigin(t *testing.T) {
	srv := nmstest.NewServer(t)
	for _, o := range []string{"", "ftp://app.example.com", "https://app.example.com/sub", "localhost:3000"} {
		svc := nms.NewService(nms.Options{Endpoint: srv.URL, Origin: func() string { return o }, Nonces: nms.NewNonces()})
		if svc.Available() {
			t.Fatalf("%q: the service must not be available", o)
		}

		if _, err := svc.Send(context.Background(), message()); !errors.Is(err, nms.ErrNotEnrolled) {
			t.Fatalf("%q: expected not enrolled, got %v", o, err)
		}
	}

	if exchanges, tokens, sends := srv.Calls(); exchanges+tokens+sends != 0 {
		t.Fatal("the service has been contacted")
	}

	if nms.NewService(nms.Options{}).Enabled() {
		t.Fatal("an empty endpoint must disable the service")
	}
}

func TestFailedExchangeIsThrottled(t *testing.T) {
	srv := nmstest.NewServer(t)
	svc := newService(t, srv, nil, "")
	srv.VerifyOrigin = func(origin, nonce string) bool { return false }

	_, err := svc.Send(context.Background(), message())
	var e *nms.Error
	if !errors.Is(err, nms.ErrNotEnrolled) || !errors.As(err, &e) || e.Code != nms.CodeOriginUnverified {
		t.Fatalf("expected an unverified origin, got %v", err)
	}

	if svc.Available() {
		t.Fatal("the service must not be available until the next exchange is due")
	}

	if _, err := svc.Send(context.Background(), message()); !errors.Is(err, nms.ErrNotEnrolled) {
		t.Fatalf("expected not enrolled, got %v", err)
	}

	if exchanges, _, _ := srv.Calls(); exchanges != 1 {
		t.Fatalf("the failed exchange has been repeated: %d", exchanges)
	}

	if st := svc.Status(); st.LastError == "" || st.Enrolled {
		t.Fatalf("unexpected status %+v", st)
	}
}

func TestNonceIsAnsweredOnce(t *testing.T) {
	nonces := nms.NewNonces()
	nonce := nonces.Issue()

	get := func(path string) int {
		rec := httptest.NewRecorder()
		nonces.Handler()(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Code
	}

	if code := get(nms.NoncePath + "unknown"); code != http.StatusNotFound {
		t.Fatalf("unknown nonce answered with %d", code)
	}

	if code := get(nms.NoncePath + nonce); code != http.StatusOK {
		t.Fatalf("nonce answered with %d", code)
	}

	if code := get(nms.NoncePath + nonce); code != http.StatusNotFound {
		t.Fatalf("nonce answered twice with %d", code)
	}
}

// A refresh token revoked while the instance was down is replaced by a new exchange.
func TestRevokedRefreshAfterRestartEnrollsAgain(t *testing.T) {
	srv := nmstest.NewServer(t)
	states := &datamem.Repository[nms.State, string]{}
	if _, err := newService(t, srv, states, "").Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	srv.RevokeAll()

	if _, err := newService(t, srv, states, "").Send(context.Background(), message()); err != nil {
		t.Fatalf("the revoked refresh token was not replaced: %v", err)
	}

	if exchanges, tokens, _ := srv.Calls(); exchanges != 2 || tokens != 3 {
		t.Fatalf("expected a second exchange after the rejected refresh, got %d exchanges and %d tokens", exchanges, tokens)
	}
}

// A developer on localhost cannot be called back, but the service may admit the instance by its address. The client
// therefore asks, and leaves the decision to the service.
func TestLocalOriginAsksForExchange(t *testing.T) {
	srv := nmstest.NewServer(t)
	var asked []string
	srv.VerifyOrigin = func(origin, nonce string) bool {
		asked = append(asked, origin)
		return true
	}

	svc := nms.NewService(nms.Options{Endpoint: srv.URL, Origin: func() string { return "http://localhost:3000/" }, Nonces: nms.NewNonces()})
	if !svc.Available() {
		t.Fatal("a local origin may ask for an exchange")
	}

	if _, err := svc.Send(context.Background(), message()); err != nil {
		t.Fatal(err)
	}

	if len(asked) != 1 || asked[0] != "http://localhost:3000" {
		t.Fatalf("expected one exchange for the normalized origin, got %v", asked)
	}
}
