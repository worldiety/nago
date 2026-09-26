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
	"net/mail"
	"net/textproto"
	"testing"
	"time"

	"go.wdy.de/nago/application/secret"
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

func TestResolveSender(t *testing.T) {
	const apiKey = "3f1c9a7e0b2d4e6f8a1b3c5d7e9f0a2b"
	cases := []struct {
		name     string
		from     mail.Address
		sender   string
		username string
		want     mail.Address
		wantErr  bool
	}{
		{name: "explicit from wins", from: mail.Address{Address: "caller@example.com"}, sender: "settings@example.com", username: "login@example.com", want: mail.Address{Address: "caller@example.com"}},
		{name: "invalid from falls back to sender, keeps name", from: mail.Address{Name: "Team", Address: "not a mail"}, sender: "settings@example.com", username: apiKey, want: mail.Address{Name: "Team", Address: "settings@example.com"}},
		{name: "sender setting", sender: " settings@example.com ", username: apiKey, want: mail.Address{Address: "settings@example.com"}},
		{name: "empty sender uses login", username: "login@example.com", want: mail.Address{Address: "login@example.com"}},
		{name: "invalid sender uses login", sender: "Team <x@example.com>", username: "login@example.com", want: mail.Address{Address: "login@example.com"}},
		{name: "mailjet api key without sender", username: apiKey, wantErr: true},
		{name: "invalid everything", from: mail.Address{Address: "x"}, sender: "y", username: apiKey, wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolveSender(secret.SMTP{Name: "mj", SenderAddress: c.sender, Username: c.username}, c.from)
			if c.wantErr {
				var se *SendError
				if !errors.As(err, &se) || se.Phase != PhaseConfig {
					t.Fatalf("expected config SendError, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if got != c.want {
				t.Fatalf("want %+v got %+v", c.want, got)
			}
		})
	}
}

func TestSendWithoutValidSenderFailsBeforeDial(t *testing.T) {
	// host is unreachable: a dial attempt would yield PhaseDial instead of PhaseConfig
	err := send(secret.SMTP{Name: "mj", Host: "invalid.invalid", Port: 587, Username: "apikey"}, Mail{To: []mail.Address{{Address: "a@example.com"}}})
	var se *SendError
	if !errors.As(err, &se) || se.Phase != PhaseConfig {
		t.Fatalf("expected config SendError, got %v", err)
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
