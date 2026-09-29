// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// blockingServer answers only once the client gave up (or after a generous safety timeout), and reports
// whether the request was aborted by the client.
func blockingServer(t *testing.T) (*httptest.Server, <-chan struct{}) {
	t.Helper()
	aborted := make(chan struct{}, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The server notices a closed connection only once the request body was consumed.
		_, _ = io.Copy(io.Discard, r.Body)

		select {
		case <-r.Context().Done():
			aborted <- struct{}{}
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, aborted
}

func TestClientCancellationAbortsRequest(t *testing.T) {
	calls := map[string]func(ctx context.Context, c *Client) error{
		"message": func(ctx context.Context, c *Client) error {
			_, err := c.CreateMessage(ctx, apiRequest{Model: "m", MaxTokens: 1})
			return err
		},
		"stream": func(ctx context.Context, c *Client) error {
			return c.CreateMessageStream(ctx, apiRequest{Model: "m", MaxTokens: 1}, func(string, []byte) error { return nil })
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			srv, aborted := blockingServer(t)

			c := NewClient("token", "", 0, false)
			c.base = srv.URL + "/"

			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(50*time.Millisecond, cancel)

			start := time.Now()
			err := call(ctx, c)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected context.Canceled, got %v", err)
			}
			if d := time.Since(start); d > 5*time.Second {
				t.Fatalf("cancellation took %s; the request was not aborted", d)
			}

			select {
			case <-aborted:
			case <-time.After(5 * time.Second):
				t.Fatal("the server never saw the request being aborted")
			}
		})
	}
}
