// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gorilla

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// A client answers the pings of the server by itself, like a browser does in a background tab whose timers are
// throttled, and every pong keeps the connection alive without any message of the client.
func TestKeepAliveCountsPongs(t *testing.T) {
	var pongs atomic.Int32
	stopped := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		channel := NewWebsocketChannel(conn)
		stop := channel.KeepAlive(10*time.Millisecond, func() { pongs.Add(1) })
		defer stop()
		_ = channel.Loop()
		close(stopped)
	}))
	defer srv.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}

	// the client only reads, its default ping handler answers with a pong
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	deadline := time.Now().Add(5 * time.Second)
	for pongs.Load() < 3 {
		if time.Now().After(deadline) {
			t.Fatalf("expected pongs, got %d", pongs.Load())
		}
		time.Sleep(5 * time.Millisecond)
	}

	_ = conn.Close()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("the loop must end with the connection")
	}
}
