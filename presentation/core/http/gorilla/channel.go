// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gorilla

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketChannel struct {
	conn         *websocket.Conn
	observers    map[int]func(msg []byte) error
	mutex        sync.RWMutex
	hnd          int
	gorillaMutex sync.Mutex
}

func NewWebsocketChannel(conn *websocket.Conn) *WebsocketChannel {
	return &WebsocketChannel{conn: conn, observers: map[int]func(msg []byte) error{}}
}

// KeepAlive sends a websocket ping in the given interval and invokes alive for every pong, until stop is called or
// a ping fails. A browser answers a ping by itself, without any script. Thus, unlike a ping of the frontend, it is
// not delayed in a background tab, whose timers a browser throttles to once a minute after five minutes. It must be
// called before [WebsocketChannel.Loop], which reads the pongs.
func (w *WebsocketChannel) KeepAlive(interval time.Duration, alive func()) (stop func()) {
	w.conn.SetPongHandler(func(string) error {
		alive()
		return nil
	})

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// WriteControl may be called concurrently with the other write methods
				if err := w.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(interval)); err != nil {
					return
				}
			}
		}
	}()

	var once sync.Once
	return func() { once.Do(func() { close(done) }) }
}

func (w *WebsocketChannel) Loop() error {
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := w.dispatch(message); err != nil {
			return err
		}
	}
}

func (w *WebsocketChannel) dispatch(buf []byte) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, f := range w.observers {
		if err := f(buf); err != nil {
			return fmt.Errorf("%s: %w", string(buf), err)
		}
	}

	return nil
}

func (w *WebsocketChannel) Subscribe(f func(msg []byte) error) (destroy func()) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.hnd++
	myHandle := w.hnd
	w.observers[myHandle] = f

	return func() {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		delete(w.observers, myHandle)
	}
}

func (w *WebsocketChannel) Publish(msg []byte) error {
	// we need this lock, to mitigate a race condition within the gorilla lib in their socket connection (beginMessage)
	w.gorillaMutex.Lock()
	defer w.gorillaMutex.Unlock()

	return w.conn.WriteMessage(websocket.BinaryMessage, msg)
}

// PublishLocal dispatches the buffer directly to currently all registered subscribers.
func (w *WebsocketChannel) PublishLocal(buf []byte) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for _, f := range w.observers {
		if err := f(buf); err != nil {
			return err
		}
	}

	return nil
}
