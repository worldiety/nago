// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xhttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// A retried request sends its body again.
func TestRetryResendsBody(t *testing.T) {
	var attempts atomic.Int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	var out struct{ OK bool }
	err := NewRequest().URL(srv.URL).Assert2xx(true).Retry(1).RetryWait(time.Millisecond).BodyJSON(map[string]string{"q": "hello"}).ToJSON(&out).Post()
	if err != nil {
		t.Fatal(err)
	}

	if len(bodies) != 2 || bodies[1] != `{"q":"hello"}` || !out.OK {
		t.Fatalf("the retry did not resend the body: %q", bodies)
	}
}

// A cancelled request does not wait for its next attempt.
func TestRetryWaitGivesUpOnCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := NewRequest().Context(ctx).URL(srv.URL).Assert2xx(true).Retry(3).RetryWait(10 * time.Second).Get()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected the deadline, got %v", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("the cancelled request waited %v for its retry", d)
	}
}

// Closing a handed over body releases the request, so the server sees the client go away.
func TestClosingHandedOverBodyReleasesRequest(t *testing.T) {
	gone := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(gone)
	}))
	defer srv.Close()

	err := NewRequest().URL(srv.URL).Timeout(50 * time.Millisecond).ToCloser(func(rc io.ReadCloser) {
		// the stream outlives the timeout of the request
		time.Sleep(150 * time.Millisecond)
		_ = rc.Close()
	}).Get()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-gone:
	case <-time.After(5 * time.Second):
		t.Fatal("the server still serves the closed stream")
	}
}

// A handed over body of an unexpected status is closed, and the callback never sees it.
func TestUnexpectedStatusClosesHandedOverBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	defer srv.Close()

	called := false
	err := NewRequest().URL(srv.URL).Assert2xx(true).ToCloser(func(rc io.ReadCloser) { called = true }).Get()
	var status UnexpectedStatusCodeError
	if !errors.As(err, &status) || status.StatusCode != http.StatusTeapot || called {
		t.Fatalf("expected the status error without a callback, got %v (called=%v)", err, called)
	}
}

// The limit applies whatever the order of To and ToLimit is.
func TestToLimitInAnyOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer srv.Close()

	for _, order := range []string{"limit first", "limit last"} {
		var got int
		read := func(r io.Reader) error {
			b, err := io.ReadAll(r)
			got = len(b)
			return err
		}

		req := NewRequest().URL(srv.URL)
		if order == "limit first" {
			req = req.ToLimit(10).To(read)
		} else {
			req = req.To(read).ToLimit(10)
		}

		if err := req.Get(); err != nil {
			t.Fatal(err)
		}
		if got != 10 {
			t.Fatalf("%s: expected 10 bytes, got %d", order, got)
		}
	}
}

// A 503 is reported with its status and body like any other status, also after the retries are exhausted.
func TestServiceUnavailableIsTyped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"overloaded"}`))
	}))
	defer srv.Close()

	for _, retry := range []int{0, 1} {
		err := NewRequest().URL(srv.URL).Assert2xx(true).Retry(retry).RetryWait(time.Millisecond).Get()
		var status UnexpectedStatusCodeError
		if !errors.As(err, &status) || status.StatusCode != http.StatusServiceUnavailable || !strings.Contains(string(status.Body), "overloaded") {
			t.Fatalf("retry %d: expected a typed 503 with body, got %v", retry, err)
		}
	}
}

// Every attempt counts against the rate limit of the group.
func TestRetriesAreThrottled(t *testing.T) {
	var arrivals []time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arrivals = append(arrivals, time.Now())
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	g := NewRequestGroup().RateLimit(4) // one request per 250ms
	_ = NewRequest().URL(srv.URL).Group(g).Assert2xx(true).Retry(2).RetryWait(time.Millisecond).Get()

	if len(arrivals) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(arrivals))
	}
	if d := arrivals[2].Sub(arrivals[0]); d < 400*time.Millisecond {
		t.Fatalf("the retries bypassed the rate limit: 3 attempts within %v", d)
	}
}

// An error of the request construction is reported right away, because no retry could fix it.
func TestConstructionErrorIsNotRetried(t *testing.T) {
	start := time.Now()
	err := NewRequest().URL("http://127.0.0.1:1").Retry(3).BodyJSON(map[string]any{"c": make(chan int)}).Post()
	if err == nil || !strings.Contains(err.Error(), "failed to create request body") {
		t.Fatalf("expected the construction error, got %v", err)
	}
	if d := time.Since(start); d > 100*time.Millisecond {
		t.Fatalf("the construction error has been retried for %v", d)
	}

	if err := NewRequest().URL("http://127.0.0.1:1").Retry(-1).Get(); err == nil {
		t.Fatal("expected a connection error")
	}
}

// A handed over body is read for as long as the caller likes; only the arrival of the response is bounded.
func TestHandedOverBodyIsBoundedByArrivalOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/late" {
			time.Sleep(300 * time.Millisecond)
		}
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		for range 10 {
			time.Sleep(20 * time.Millisecond)
			_, _ = w.Write([]byte("x"))
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	var got int
	err := NewRequest().URL(srv.URL).Timeout(50 * time.Millisecond).ToCloser(func(rc io.ReadCloser) {
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			t.Errorf("reading the stream failed: %v", err)
		}
		got = len(b)
	}).Get()
	if err != nil || got != 10 {
		t.Fatalf("expected the whole stream, got %d bytes, %v", got, err)
	}

	err = NewRequest().URL(srv.URL + "/late").Timeout(50 * time.Millisecond).ToCloser(func(rc io.ReadCloser) { _ = rc.Close() }).Get()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected the arrival timeout, got %v", err)
	}
}

// A JSON body cut by the limit names the limit in its error.
func TestToJSONMentionsTheLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"text":"` + strings.Repeat("x", 5000) + `"}`))
	}))
	defer srv.Close()

	var out map[string]any
	err := NewRequest().URL(srv.URL).ToLimit(100).ToJSON(&out).Get()
	var withBody ErrorWithBody
	if !errors.As(err, &withBody) || !strings.Contains(err.Error(), "limit of 100 bytes") || len(withBody.Body) != 100 {
		t.Fatalf("expected the limit in the error, got %v", err)
	}
}
