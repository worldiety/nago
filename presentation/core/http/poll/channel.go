// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package poll implements a [core.Channel] over plain http requests, which the frontend uses, when a proxy or
// firewall blocks the websocket. The client fetches the published messages by long polling and posts its own
// messages. Both directions number their messages, see [WriteFrames], thus a request which is repeated after a
// network error neither loses nor duplicates a message.
package poll

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// MaxQueued limits the bytes of the published messages, which the client has not acknowledged yet. A client, which
// does not poll anymore, exceeds it eventually and the channel is closed.
const MaxQueued = 32 * 1024 * 1024

var (
	// ErrClosed is returned by a closed channel.
	ErrClosed = errors.New("poll channel is closed")
	// ErrOverflow closes a channel, whose client does not fetch its messages.
	ErrOverflow = errors.New("poll channel exceeds its queue")
)

// Channel is a [core.Channel], whose messages are fetched by [Channel.Poll] and received by [Channel.Receive].
type Channel struct {
	token string

	mutex    sync.Mutex
	queue    []message // the published messages, which the client has not acknowledged yet
	queued   int       // the bytes of queue
	nextSeq  uint64    // the number of the next published message, starting at 1
	wake     chan struct{}
	closed   bool
	lastSeen time.Time

	recvMutex sync.Mutex // keeps the received messages in order
	recvSeq   uint64     // the number of the last received message

	obsMutex  sync.RWMutex
	observers map[int]func(msg []byte) error
	hnd       int
}

type message struct {
	seq  uint64
	data []byte
}

// New creates a channel with a random token, which identifies it to the client.
func New() *Channel {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	return &Channel{
		token:     hex.EncodeToString(buf[:]),
		nextSeq:   1,
		wake:      make(chan struct{}),
		lastSeen:  time.Now(),
		observers: map[int]func(msg []byte) error{},
	}
}

// Token identifies the channel. A client, which does not know it, is not connected to this channel.
func (c *Channel) Token() string {
	return c.token
}

// Subscribe registers f for the received messages.
func (c *Channel) Subscribe(f func(msg []byte) error) (destroy func()) {
	c.obsMutex.Lock()
	defer c.obsMutex.Unlock()

	c.hnd++
	handle := c.hnd
	c.observers[handle] = f

	return func() {
		c.obsMutex.Lock()
		defer c.obsMutex.Unlock()
		delete(c.observers, handle)
	}
}

// Publish queues the message for the client. The channel keeps msg until the client has acknowledged it, thus the
// caller must not modify it afterward.
func (c *Channel) Publish(msg []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return ErrClosed
	}

	if c.queued+len(msg) > MaxQueued {
		c.closeLocked()
		return ErrOverflow
	}

	c.queue = append(c.queue, message{seq: c.nextSeq, data: msg})
	c.queued += len(msg)
	c.nextSeq++
	c.signalLocked()
	return nil
}

// Poll acknowledges the messages up to ack and returns the following ones with the number of the first one. If there
// are none, it waits for them up to wait or until ctx is done, and returns none.
func (c *Channel) Poll(ctx context.Context, ack uint64, wait time.Duration) (first uint64, msgs [][]byte, err error) {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	for {
		c.mutex.Lock()
		c.lastSeen = time.Now()
		if c.closed {
			c.mutex.Unlock()
			return 0, nil, ErrClosed
		}

		if ack >= c.nextSeq {
			c.mutex.Unlock()
			return 0, nil, fmt.Errorf("acknowledged message %d has not been published", ack)
		}

		n := 0
		for n < len(c.queue) && c.queue[n].seq <= ack {
			c.queued -= len(c.queue[n].data)
			n++
		}

		// keep the backing array from growing forever, while the queue is mostly empty
		c.queue = append(c.queue[:0:0], c.queue[n:]...)
		if len(c.queue) > 0 {
			first = c.queue[0].seq
			msgs = make([][]byte, len(c.queue))
			for i, m := range c.queue {
				msgs[i] = m.data
			}

			c.mutex.Unlock()
			return first, msgs, nil
		}

		wake := c.wake
		c.mutex.Unlock()

		select {
		case <-wake:
		case <-timer.C:
			return c.nextSeqOf(), nil, nil
		case <-ctx.Done():
			return c.nextSeqOf(), nil, nil
		}
	}
}

func (c *Channel) nextSeqOf() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastSeen = time.Now()
	return c.nextSeq
}

// Receive dispatches the messages of the client to the observers, whose numbers start at first. It skips the messages,
// which it has received before, thus the client may repeat a request after a network error. A gap is an error.
func (c *Channel) Receive(first uint64, msgs [][]byte) error {
	c.recvMutex.Lock()
	defer c.recvMutex.Unlock()

	c.mutex.Lock()
	closed := c.closed
	c.lastSeen = time.Now()
	c.mutex.Unlock()
	if closed {
		return ErrClosed
	}

	if first == 0 || first > c.recvSeq+1 {
		return fmt.Errorf("expected message %d, got %d", c.recvSeq+1, first)
	}

	for i, msg := range msgs {
		seq := first + uint64(i)
		if seq <= c.recvSeq {
			continue
		}

		if err := c.dispatch(msg); err != nil {
			return err
		}

		c.recvSeq = seq
	}

	return nil
}

func (c *Channel) dispatch(msg []byte) error {
	c.obsMutex.RLock()
	defer c.obsMutex.RUnlock()

	for _, f := range c.observers {
		if err := f(msg); err != nil {
			return err
		}
	}

	return nil
}

// Idle returns the duration since the last request of the client.
func (c *Channel) Idle() time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return time.Since(c.lastSeen)
}

// Closed tells whether the channel is closed.
func (c *Channel) Closed() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.closed
}

// Close drops the queued messages and ends the waiting polls.
func (c *Channel) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closeLocked()
}

func (c *Channel) closeLocked() {
	if c.closed {
		return
	}

	c.closed = true
	c.queue = nil
	c.queued = 0
	c.signalLocked()
}

func (c *Channel) signalLocked() {
	close(c.wake)
	c.wake = make(chan struct{})
}
