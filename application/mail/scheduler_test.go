// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"context"
	"fmt"
	"iter"
	"net/mail"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
)

type sentLog struct {
	mutex sync.Mutex
	subj  []string
	at    []time.Time
}

func (l *sentLog) add(m Mail) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.subj = append(l.subj, m.Subject)
	l.at = append(l.at, time.Now())
}

func (l *sentLog) subjects() []string {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return append([]string(nil), l.subj...)
}

func testSecrets(servers ...secret.SMTP) secret.FindGroupSecrets {
	return func(subject auth.Subject, gid group.ID) iter.Seq2[secret.Secret, error] {
		return func(yield func(secret.Secret, error) bool) {
			for i, s := range servers {
				if !yield(secret.Secret{ID: secret.ID(fmt.Sprint(i)), Credentials: s}, nil) {
					return
				}
			}
		}
	}
}

func newTestScheduler(t *testing.T, opts ScheduleOptions, servers ...secret.SMTP) (*scheduler, Repository, *sentLog) {
	t.Helper()
	repo := json.NewSloppyJSONRepository[Outgoing, ID](mem.NewBlobStore("outgoing"))
	if len(servers) == 0 {
		servers = []secret.SMTP{{Name: "a", Host: "localhost", Port: 25}}
	}

	s := newScheduler(opts, repo, func() user.Subject { return user.SU() }, testSecrets(servers...))
	log := &sentLog{}
	s.send = func(credentials secret.SMTP, m Mail) error {
		log.add(m)
		return nil
	}
	s.sleep = func(time.Duration) {}
	return s, repo, log
}

func queue(t *testing.T, repo Repository, subject string, to string, queuedAt time.Time) Outgoing {
	t.Helper()
	o := Outgoing{
		ID:       data.RandIdent[ID](),
		Mail:     Mail{To: []mail.Address{{Address: to}}, Subject: subject, Parts: []Part{NewTextPart("body " + subject)}},
		Subject:  subject,
		Receiver: to,
		Status:   StatusQueued,
		QueuedAt: queuedAt,
	}
	if err := repo.Save(o); err != nil {
		t.Fatal(err)
	}
	return o
}

func TestSchedulerWakeupSendsImmediately(t *testing.T) {
	notify, wakeup := NewWakeup()
	s, repo, log := newTestScheduler(t, ScheduleOptions{SendInterval: time.Hour, Wakeup: wakeup})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.run(ctx)

	time.Sleep(50 * time.Millisecond) // the initial run finds nothing and waits for an hour
	start := time.Now()
	queue(t, repo, "hello", "a@example.com", start)
	notify()

	for time.Since(start) < 2*time.Second {
		if len(log.subjects()) == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("mail has not been sent after wakeup")
}

func TestSchedulerFreshMailsFirstAndBrokenServer(t *testing.T) {
	s, repo, log := newTestScheduler(t, ScheduleOptions{})
	now := time.Now()

	// two old retries, one new mail
	for i := 0; i < 2; i++ {
		o := queue(t, repo, fmt.Sprintf("retry-%d", i), "r@example.com", now.Add(-time.Hour))
		o.Status = StatusError
		o.addAttempt(Attempt{At: now.Add(-time.Minute), Server: "a", Phase: PhaseAuth})
		if err := repo.Save(o); err != nil {
			t.Fatal(err)
		}
	}
	queue(t, repo, "fresh", "f@example.com", now)

	s.send = func(credentials secret.SMTP, m Mail) error {
		log.add(m)
		return &SendError{Phase: PhaseAuth, Code: 535, Err: fmt.Errorf("auth failed")}
	}

	s.runOnce(context.Background())

	got := log.subjects()
	if len(got) != 1 || got[0] != "fresh" {
		t.Fatalf("expected only the fresh mail to be attempted on a broken server, got %v", got)
	}
}

func TestSchedulerRateLimit(t *testing.T) {
	s, repo, log := newTestScheduler(t, ScheduleOptions{}, secret.SMTP{Name: "a", Host: "h", RateLimitPerHour: 2})
	now := time.Now()
	for i := 0; i < 5; i++ {
		queue(t, repo, fmt.Sprintf("m%d", i), fmt.Sprintf("u%d@example.com", i), now.Add(time.Duration(i)*time.Second))
	}

	s.runOnce(context.Background())
	if n := len(log.subjects()); n != 2 {
		t.Fatalf("expected 2 sent mails, got %d", n)
	}

	if st := currentSchedulerStatus(); len(st.RateLimited) != 1 || st.RateLimited[0] != "a" {
		t.Fatalf("expected rate limited status, got %v", st.RateLimited)
	}

	queued := 0
	for o, err := range repo.All() {
		if err != nil {
			t.Fatal(err)
		}
		if o.Status == StatusQueued {
			queued++
		}
	}

	if queued != 3 {
		t.Fatalf("expected 3 postponed mails, got %d", queued)
	}
}

func TestSchedulerSpamGuard(t *testing.T) {
	s, repo, log := newTestScheduler(t, ScheduleOptions{})
	now := time.Now()
	for i := 0; i < 8; i++ {
		queue(t, repo, "runaway", "victim@example.com", now.Add(time.Duration(i)*time.Millisecond))
	}

	s.runOnce(context.Background())
	if n := len(log.subjects()); n != 5 {
		t.Fatalf("expected 5 sent mails, got %d", n)
	}

	var suppressed []ID
	for o, err := range repo.All() {
		if err != nil {
			t.Fatal(err)
		}
		if o.Status == StatusSuppressed {
			suppressed = append(suppressed, o.ID)
		}
	}

	if len(suppressed) != 3 {
		t.Fatalf("expected 3 suppressed mails, got %d", len(suppressed))
	}

	// an explicit retry bypasses the guard
	var mutex sync.Mutex
	if err := NewRetryOutgoing(&mutex, repo)(user.SU(), suppressed[0]); err != nil {
		t.Fatal(err)
	}

	s.runOnce(context.Background())
	if n := len(log.subjects()); n != 6 {
		t.Fatalf("expected the retried mail to be sent, got %d", n)
	}
}

func TestRateLimiterSeed(t *testing.T) {
	stats := json.NewSloppyJSONRepository[StatsBucket, StatsBucketID](mem.NewBlobStore("stats"))
	now := time.Now()
	hour := now.UTC().Truncate(time.Hour)
	if err := stats.Save(StatsBucket{ID: statsBucketID(hour, "a"), Hour: hour, Server: "a", Sent: 3}); err != nil {
		t.Fatal(err)
	}

	r := newRateLimiter()
	if err := r.seed(stats, now); err != nil {
		t.Fatal(err)
	}

	if r.allow("a", 3, 0, now) {
		t.Fatal("expected hourly limit from seeded statistics")
	}

	if !r.allow("a", 0, 4, now) {
		t.Fatal("expected daily limit not yet reached")
	}
}
