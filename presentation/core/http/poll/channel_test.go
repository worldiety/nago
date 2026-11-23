// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package poll

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func poll(t *testing.T, c *Channel, ack uint64, wait time.Duration) (uint64, []string) {
	t.Helper()
	first, msgs, err := c.Poll(context.Background(), ack, wait)
	if err != nil {
		t.Fatal(err)
	}

	var s []string
	for _, m := range msgs {
		s = append(s, string(m))
	}

	return first, s
}

// A poll returns the messages, which the client has not acknowledged, again, thus a lost response loses nothing.
func TestChannel_PollRepeatsUnacknowledged(t *testing.T) {
	c := New()
	for _, m := range []string{"a", "b", "c"} {
		if err := c.Publish([]byte(m)); err != nil {
			t.Fatal(err)
		}
	}

	for range 2 {
		if first, msgs := poll(t, c, 0, time.Second); first != 1 || fmt.Sprint(msgs) != "[a b c]" {
			t.Fatalf("got %d %v", first, msgs)
		}
	}

	if first, msgs := poll(t, c, 2, time.Second); first != 3 || fmt.Sprint(msgs) != "[c]" {
		t.Fatalf("got %d %v", first, msgs)
	}

	// an empty poll answers with the number of the next message
	if first, msgs := poll(t, c, 3, 10*time.Millisecond); first != 4 || len(msgs) != 0 {
		t.Fatalf("got %d %v", first, msgs)
	}

	if _, _, err := c.Poll(context.Background(), 4, time.Second); err == nil {
		t.Fatal("a message, which has not been published, cannot be acknowledged")
	}
}

// A waiting poll returns, as soon as a message is published, or the channel is closed.
func TestChannel_PollWaits(t *testing.T) {
	c := New()
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = c.Publish([]byte("late"))
	}()

	start := time.Now()
	first, msgs := poll(t, c, 0, 10*time.Second)
	if first != 1 || fmt.Sprint(msgs) != "[late]" || time.Since(start) > 5*time.Second {
		t.Fatalf("got %d %v after %v", first, msgs, time.Since(start))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if first, msgs, err := c.Poll(ctx, 1, 10*time.Second); err != nil || first != 2 || len(msgs) != 0 {
		t.Fatalf("a cancelled poll returns nothing: %d %v %v", first, msgs, err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		c.Close()
	}()

	if _, _, err := c.Poll(context.Background(), 1, 10*time.Second); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected the close to end the poll: %v", err)
	}

	if err := c.Publish([]byte("x")); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed: %v", err)
	}
}

// Receive dispatches each message once and in order, also when the client repeats a request.
func TestChannel_Receive(t *testing.T) {
	c := New()
	var got []string
	destroy := c.Subscribe(func(msg []byte) error {
		got = append(got, string(msg))
		return nil
	})

	msgs := func(s ...string) [][]byte {
		var b [][]byte
		for _, m := range s {
			b = append(b, []byte(m))
		}
		return b
	}

	for _, step := range []struct {
		first uint64
		msgs  [][]byte
		fails bool
	}{
		{1, msgs("a", "b"), false},
		{1, msgs("a", "b"), false}, // repeated after a lost response
		{2, msgs("b", "c"), false},
		{5, msgs("e"), true}, // a gap
		{0, msgs("x"), true},
		{4, msgs("d"), false},
	} {
		if err := c.Receive(step.first, step.msgs); (err != nil) != step.fails {
			t.Fatalf("receive %d: %v", step.first, err)
		}
	}

	if fmt.Sprint(got) != "[a b c d]" {
		t.Fatalf("got %v", got)
	}

	destroy()
	if err := c.Receive(5, msgs("e")); err != nil || len(got) != 4 {
		t.Fatalf("an unsubscribed observer must not receive: %v %v", got, err)
	}
}

// A client, which does not fetch its messages, closes the channel eventually.
func TestChannel_Overflow(t *testing.T) {
	c := New()
	big := make([]byte, MaxQueued/4)
	var err error
	for range 5 {
		if err = c.Publish(big); err != nil {
			break
		}
	}

	if !errors.Is(err, ErrOverflow) || !c.Closed() {
		t.Fatalf("expected an overflow: %v", err)
	}
}

// Concurrent publishers and a polling client receive every message exactly once and in order.
func TestChannel_Concurrent(t *testing.T) {
	c := New()
	const publishers, each = 4, 500
	var wg sync.WaitGroup
	var publishMutex sync.Mutex // the scope publishes in order, thus the order of the numbers is the order of the calls
	var published []string
	for p := range publishers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range each {
				publishMutex.Lock()
				msg := fmt.Sprintf("%d-%d", p, i)
				published = append(published, msg)
				if err := c.Publish([]byte(msg)); err != nil {
					t.Error(err)
				}
				publishMutex.Unlock()
			}
		}()
	}

	var received []string
	var ack uint64
	for len(received) < publishers*each {
		first, msgs, err := c.Poll(context.Background(), ack, time.Second)
		if err != nil {
			t.Fatal(err)
		}

		for i, m := range msgs {
			if seq := first + uint64(i); seq > ack {
				received = append(received, string(m))
				ack = seq
			}
		}
	}

	wg.Wait()
	if fmt.Sprint(received) != fmt.Sprint(published) {
		t.Fatal("the messages differ")
	}
}

func TestFrames(t *testing.T) {
	for _, msgs := range [][][]byte{nil, {{}}, {[]byte("a"), bytes.Repeat([]byte("b"), 300), {}}} {
		var buf bytes.Buffer
		if err := WriteFrames(&buf, 42, msgs); err != nil {
			t.Fatal(err)
		}

		first, got, err := ReadFrames(bytes.NewReader(buf.Bytes()), 1000)
		if err != nil || first != 42 || len(got) != len(msgs) {
			t.Fatalf("got %d %d messages: %v", first, len(got), err)
		}

		for i := range msgs {
			if !bytes.Equal(got[i], msgs[i]) {
				t.Fatalf("message %d differs", i)
			}
		}

		if len(msgs) > 1 {
			if _, _, err := ReadFrames(bytes.NewReader(buf.Bytes()[:buf.Len()-5]), 1000); err == nil {
				t.Fatal("expected an error for a truncated body")
			}

			if _, _, err := ReadFrames(bytes.NewReader(buf.Bytes()), 100); err == nil {
				t.Fatal("expected an error for a too large message")
			}
		}
	}

	if _, _, err := ReadFrames(bytes.NewReader(nil), 100); err == nil {
		t.Fatal("expected an error for an empty body")
	}
}
