// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"encoding/json"
	"errors"
	"net/textproto"
	"testing"
	"time"
)

// legacyOutgoing is the persisted format before attempts and statistics have been introduced.
type legacyOutgoing struct {
	ID         ID
	Mail       Mail
	Subject    string
	Receiver   string
	Status     Status
	LastError  string
	ServerName string
	QueuedAt   time.Time
	SendAt     time.Time
}

func TestOutgoingJSONCompatibility(t *testing.T) {
	now := time.Date(2026, 9, 23, 14, 0, 0, 0, time.UTC)
	legacy := legacyOutgoing{ID: "a", Subject: "s", Receiver: "r", Status: StatusError, LastError: "boom", ServerName: "o365", QueuedAt: now, SendAt: now}
	buf, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}

	// old -> new
	var cur Outgoing
	if err := json.Unmarshal(buf, &cur); err != nil {
		t.Fatal(err)
	}

	if cur.ID != "a" || cur.Status != StatusError || cur.LastError != "boom" || !cur.SendAt.Equal(now) {
		t.Fatalf("unexpected %+v", cur)
	}

	if cur.Attempted() != 1 {
		t.Fatalf("expected legacy attempt count 1, got %d", cur.Attempted())
	}

	// new -> old
	cur.addAttempt(Attempt{At: now.Add(time.Minute), Server: "mg", Phase: PhaseAuth, Code: 535, Message: "auth"})
	buf, err = json.Marshal(cur)
	if err != nil {
		t.Fatal(err)
	}

	var back legacyOutgoing
	if err := json.Unmarshal(buf, &back); err != nil {
		t.Fatal(err)
	}

	if back.ID != "a" || back.ServerName != "mg" || !back.SendAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("unexpected %+v", back)
	}

	// unchanged legacy mails keep their exact json shape
	buf1, _ := json.Marshal(legacy)
	buf2, _ := json.Marshal(Outgoing{ID: "a", Subject: "s", Receiver: "r", Status: StatusError, LastError: "boom", ServerName: "o365", QueuedAt: now, SendAt: now})
	if string(buf1) != string(buf2) {
		t.Fatalf("json shape changed:\n%s\n%s", buf1, buf2)
	}
}

func TestAddAttempt(t *testing.T) {
	var o Outgoing
	for i := 0; i < MaxAttemptHistory+5; i++ {
		o.addAttempt(Attempt{At: time.Unix(int64(i), 0), Server: "x"})
	}

	if o.AttemptCount != MaxAttemptHistory+5 {
		t.Fatalf("got %d", o.AttemptCount)
	}

	if len(o.Attempts) != MaxAttemptHistory {
		t.Fatalf("got %d", len(o.Attempts))
	}

	if a, _ := o.LastAttempt(); a.At.Unix() != int64(MaxAttemptHistory+4) {
		t.Fatalf("unexpected last %v", a)
	}
}

func TestBackoff(t *testing.T) {
	base := 30 * time.Second
	cases := map[int]time.Duration{0: base, 1: base, 2: time.Minute, 3: 2 * time.Minute, 20: time.Hour}
	for n, want := range cases {
		if got := backoff(base, time.Hour, n); got != want {
			t.Errorf("attempt %d: want %v got %v", n, want, got)
		}
	}
}

func TestSendErrCode(t *testing.T) {
	err := sendErr(PhaseAuth, &textproto.Error{Code: 535, Msg: "auth failed"})
	var se *SendError
	if !errors.As(err, &se) || se.Code != 535 || se.Phase != PhaseAuth {
		t.Fatalf("unexpected %#v", err)
	}

	if sendErr(PhaseQuit, nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestStatsRecorderAndFilter(t *testing.T) {
	now := time.Now()
	f := OutgoingFilter{Stuck: time.Hour}
	if !f.matches(Outgoing{Status: StatusError, QueuedAt: now.Add(-2 * time.Hour)}, now) {
		t.Fatal("expected stuck")
	}

	if f.matches(Outgoing{Status: StatusFailed, QueuedAt: now.Add(-2 * time.Hour)}, now) {
		t.Fatal("failed is not stuck")
	}

	tot := Totals{Sent: 3, Failed: 1, LatencySum: 30 * time.Second}
	if tot.ErrorRate() != 0.25 || tot.AvgLatency() != 10*time.Second {
		t.Fatalf("unexpected %v %v", tot.ErrorRate(), tot.AvgLatency())
	}
}
