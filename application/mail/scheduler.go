// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
)

type ScheduleOptions struct {
	SendInterval         time.Duration // default is 30 seconds, fallback polling interval for retries, if no wakeup signal is received
	KeepMailAfterSuccess time.Duration // default is 24 hours, if negative unlimited
	KeepMailAfterError   time.Duration // default is 1 year, if negative unlimited
	// WaitBetweenSends is an anti-spam pause after each successful delivery, if more mails are pending. It never
	// delays the first mail of a run. Default is 1 second, if negative no pause. Use the per-server rate limits of
	// [secret.SMTP] to comply with provider limits.
	WaitBetweenSends time.Duration

	// MaxAttempts is the amount of attempts, before a mail is marked as [StatusFailed]. Default is 12, if negative
	// unlimited. Permanent SMTP rejections (5xx) of recipients fail immediately.
	MaxAttempts int

	// MaxBackoff limits the exponential backoff between attempts of the same mail. Default is 1 hour.
	MaxBackoff time.Duration

	// CleanupInterval defines how often expired mails are removed. Default is 10 minutes.
	CleanupInterval time.Duration

	// Wakeup is optional and triggers an immediate scheduler run, e.g. after a mail has been queued.
	// See [NewWakeup].
	Wakeup <-chan struct{}

	// SpamGuard configures the heuristic to detect runaway send loops.
	SpamGuard SpamGuardOptions

	// Stats is optional and receives hourly aggregated statistics.
	Stats StatsRepository
	// Health is optional and receives the latest state per smtp server.
	Health HealthRepository
}

// NewWakeup creates a non-blocking notify function and the according channel for [ScheduleOptions.Wakeup].
// Multiple notifications are coalesced.
func NewWakeup() (notify func(), wakeup <-chan struct{}) {
	ch := make(chan struct{}, 1)
	return func() {
		select {
		case ch <- struct{}{}:
		default:
		}
	}, ch
}

type sendFunc func(credentials secret.SMTP, m Mail) error

type scheduler struct {
	opts        ScheduleOptions
	mails       Repository
	sysUser     user.SysUser
	secrets     secret.FindGroupSecrets
	recorder    *statsRecorder
	limiter     *rateLimiter
	guard       *spamGuard
	send        sendFunc
	sleep       func(time.Duration)
	now         func() time.Time
	lastCleanup time.Time
}

// StartScheduler starts a new scheduler instance to process the [Outgoing] mails.
func StartScheduler(ctx context.Context, opts ScheduleOptions, mails Repository, sysUser user.SysUser, secrets secret.FindGroupSecrets) {
	s := newScheduler(opts, mails, sysUser, secrets)
	go s.run(ctx)
}

func newScheduler(opts ScheduleOptions, mails Repository, sysUser user.SysUser, secrets secret.FindGroupSecrets) *scheduler {
	if opts.SendInterval <= 0 {
		opts.SendInterval = time.Second * 30
	}

	if opts.KeepMailAfterSuccess == 0 {
		opts.KeepMailAfterSuccess = time.Hour * 24
	}

	if opts.KeepMailAfterError == 0 {
		opts.KeepMailAfterError = time.Hour * 24 * 30 * 12
	}

	if opts.WaitBetweenSends == 0 {
		opts.WaitBetweenSends = time.Second
	}

	if opts.MaxAttempts == 0 {
		opts.MaxAttempts = 12
	}

	if opts.MaxBackoff <= 0 {
		opts.MaxBackoff = time.Hour
	}

	if opts.CleanupInterval <= 0 {
		opts.CleanupInterval = 10 * time.Minute
	}

	s := &scheduler{
		opts:     opts,
		mails:    mails,
		sysUser:  sysUser,
		secrets:  secrets,
		recorder: &statsRecorder{stats: opts.Stats, health: opts.Health},
		limiter:  newRateLimiter(),
		guard:    newSpamGuard(opts.SpamGuard),
		send:     send,
		sleep:    time.Sleep,
		now:      time.Now,
	}

	if err := s.limiter.seed(opts.Stats, s.now()); err != nil {
		slog.Error("mail scheduler cannot seed rate limiter", "err", err)
	}

	return s
}

func (s *scheduler) run(ctx context.Context) {
	slog.Info("mail scheduler started")
	setSchedulerStatus(func(st *SchedulerStatus) { st.Running = true })
	defer func() {
		setSchedulerStatus(func(st *SchedulerStatus) { st.Running = false })
		slog.Info("mail scheduler stopped")
	}()

	ticker := time.NewTicker(s.opts.SendInterval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		if s.runOnce(ctx) {
			continue // more work is pending, e.g. mails queued while we were busy
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.opts.Wakeup: // a nil channel blocks forever, which is fine
		}
	}
}

// dueEntry is a lightweight snapshot to order the mails without keeping the entire content in memory.
type dueEntry struct {
	id        ID
	fresh     bool // never attempted
	queuedAt  time.Time
	nextAt    time.Time
	smtpHint  string
	attempted int
}

type smtpCandidate struct {
	id   secret.ID
	smtp secret.SMTP
}

// runOnce processes all due mails. It returns true, if the run was interrupted because new mails arrived.
func (s *scheduler) runOnce(ctx context.Context) (more bool) {
	now := s.now()
	setSchedulerStatus(func(st *SchedulerStatus) { st.LastRunAt = now })

	if now.Sub(s.lastCleanup) >= s.opts.CleanupInterval {
		s.cleanup(now)
		s.lastCleanup = now
	}

	due, err := s.collectDue(now)
	if err != nil {
		slog.Error("mail scheduler failed to iterate on outgoing mail repository", "err", err)
		return false
	}

	if len(due) == 0 {
		setSchedulerStatus(func(st *SchedulerStatus) { st.RateLimited = nil })
		return false
	}

	candidates, err := s.loadCandidates()
	if err != nil {
		slog.Error("mail scheduler failed to pick mail server", "err", err)
		setSchedulerStatus(func(st *SchedulerStatus) { st.LastSmtpErrMsg = err.Error() })
		return false
	}

	if len(candidates) == 0 {
		slog.Error("cannot process mail queue, no smtp credentials available in system group")
		setSchedulerStatus(func(st *SchedulerStatus) { st.NoSmtpServer = true })
		return false
	}

	setSchedulerStatus(func(st *SchedulerStatus) { st.NoSmtpServer = false; st.LastSmtpErrMsg = "" })

	broken := map[string]bool{}      // servers which failed with a connection or configuration problem in this run
	rateLimited := map[string]bool{} // servers which reached their rate limit in this run
	defer func() {
		var names []string
		for name := range rateLimited {
			names = append(names, name)
		}
		slices.Sort(names)
		setSchedulerStatus(func(st *SchedulerStatus) { st.RateLimited = names })
	}()

	sentInRun := false
	for _, entry := range due {
		if ctx.Err() != nil {
			return false
		}

		// prefer new mails: if something has been queued in the meantime, restart the run before the next retry
		if !entry.fresh && s.wakeupPending() {
			return true
		}

		smtp := pickCandidate(candidates, entry.smtpHint)

		if !entry.fresh && broken[smtp.Name] {
			continue // don't hammer a broken server with retries, but still try new mails once
		}

		if rateLimited[smtp.Name] || !s.limiter.allow(smtp.Name, smtp.RateLimitPerHour, smtp.RateLimitPerDay, s.now()) {
			if !rateLimited[smtp.Name] {
				slog.Warn("mail scheduler: smtp rate limit reached, postponing mails", "smtp", smtp.Name)
			}
			rateLimited[smtp.Name] = true
			continue
		}

		// reload the current state, it may have been deleted or changed in the meantime
		optOut, err := s.mails.FindByID(entry.id)
		if err != nil {
			slog.Error("mail scheduler failed to load outgoing mail", "id", entry.id, "err", err)
			continue
		}

		if optOut.IsNone() {
			continue
		}

		outgoing := optOut.Unwrap()
		if outgoing.done() || (!outgoing.NextAttemptAt.IsZero() && s.now().Before(outgoing.NextAttemptAt)) {
			continue
		}

		if sentInRun && s.opts.WaitBetweenSends > 0 {
			s.sleep(s.opts.WaitBetweenSends)
		}

		if reason := s.guard.check(outgoing, s.now()); reason != "" {
			slog.Error("mail scheduler suppressed mail", "id", outgoing.ID, "subject", outgoing.Mail.Subject, "receiver", outgoing.Receiver, "reason", reason)
			outgoing.Status = StatusSuppressed
			outgoing.LastError = reason
			outgoing.NextAttemptAt = time.Time{}
			if err := s.mails.Save(outgoing); err != nil {
				slog.Error("failed to save outgoing mail state", "id", outgoing.ID, "err", err)
			}
			continue
		}

		attempt := s.deliver(smtp, &outgoing)
		s.limiter.record(smtp.Name, attempt.At)
		s.recorder.record(outgoing, attempt)

		if attempt.Success {
			s.guard.record(outgoing, attempt.At)
			sentInRun = true
		} else if attempt.Phase == PhaseConfig || attempt.Phase == PhaseDial || attempt.Phase == PhaseTLS || attempt.Phase == PhaseAuth {
			broken[smtp.Name] = true
		}

		if err := s.mails.Save(outgoing); err != nil {
			slog.Error("failed to save outgoing mail state", "id", outgoing.ID, "subject", outgoing.Mail.Subject, "err", err)
			continue
		}
	}

	return false
}

func (s *scheduler) wakeupPending() bool {
	select {
	case <-s.opts.Wakeup:
		return true
	default:
		return false
	}
}

// deliver performs the send attempt and updates the outgoing state.
func (s *scheduler) deliver(smtp secret.SMTP, outgoing *Outgoing) Attempt {
	start := s.now()
	sendErr := s.send(smtp, outgoing.Mail)
	attempt := Attempt{
		At:       start,
		Server:   smtp.Name,
		Success:  sendErr == nil,
		Duration: s.now().Sub(start),
	}

	if sendErr != nil {
		attempt.Message = sendErr.Error()
		var se *SendError
		if errors.As(sendErr, &se) {
			attempt.Phase = se.Phase
			attempt.Code = se.Code
			attempt.Message = se.Err.Error()
		}
	}

	outgoing.addAttempt(attempt)

	if sendErr != nil {
		slog.Error("mail scheduler failed to send mail", "smtp", smtp.Name, "id", outgoing.ID, "subject", outgoing.Mail.Subject, "err", sendErr)

		outgoing.Status = StatusError
		outgoing.LastError = sendErr.Error()

		permanent := attempt.Phase == PhaseRcpt && attempt.Code >= 500
		if permanent || (s.opts.MaxAttempts > 0 && outgoing.AttemptCount >= s.opts.MaxAttempts) {
			outgoing.Status = StatusFailed
			outgoing.NextAttemptAt = time.Time{}
		} else {
			outgoing.NextAttemptAt = s.now().Add(backoff(s.opts.SendInterval, s.opts.MaxBackoff, outgoing.AttemptCount))
		}
	} else {
		slog.Info("mail scheduler send mail success", "id", outgoing.ID)
		outgoing.Status = StatusSendSuccess
		outgoing.LastError = ""
		outgoing.SentAt = s.now()
		outgoing.NextAttemptAt = time.Time{}
	}

	return attempt
}

// collectDue returns all mails which shall be attempted now, new mails first in queue order, then retries.
func (s *scheduler) collectDue(now time.Time) ([]dueEntry, error) {
	var due []dueEntry
	for o, err := range s.mails.All() {
		if err != nil {
			return nil, err
		}

		if o.done() {
			continue
		}

		if !o.NextAttemptAt.IsZero() && now.Before(o.NextAttemptAt) {
			continue
		}

		due = append(due, dueEntry{
			id:        o.ID,
			fresh:     o.Attempted() == 0,
			queuedAt:  o.QueuedAt,
			nextAt:    o.NextAttemptAt,
			smtpHint:  o.Mail.SmtpHint,
			attempted: o.Attempted(),
		})
	}

	slices.SortStableFunc(due, func(a, b dueEntry) int {
		if a.fresh != b.fresh {
			if a.fresh {
				return -1
			}
			return 1
		}

		if c := a.queuedAt.Compare(b.queuedAt); c != 0 {
			return c
		}

		return a.nextAt.Compare(b.nextAt)
	})

	return due, nil
}

func (s *scheduler) cleanup(now time.Time) {
	var toRemove []ID
	for o, err := range s.mails.All() {
		if err != nil {
			slog.Error("mail scheduler failed to iterate on outgoing mail repository", "err", err)
			return
		}

		if s.opts.KeepMailAfterSuccess >= 0 && o.Status == StatusSendSuccess && now.Sub(o.QueuedAt) > s.opts.KeepMailAfterSuccess {
			toRemove = append(toRemove, o.ID)
		}

		if s.opts.KeepMailAfterError >= 0 && (o.Status == StatusError || o.Status == StatusFailed || o.Status == StatusSuppressed) && now.Sub(o.QueuedAt) > s.opts.KeepMailAfterError {
			toRemove = append(toRemove, o.ID)
		}
	}

	if len(toRemove) > 0 {
		if err := s.mails.DeleteAllByID(slices.Values(toRemove)); err != nil {
			slog.Error("mail scheduler failed to remove mails", "err", err)
		}
	}
}

func (s *scheduler) loadCandidates() ([]smtpCandidate, error) {
	var res []smtpCandidate
	for scr, err := range s.secrets(s.sysUser(), group.System) {
		if err != nil {
			return nil, err
		}

		if smtp, ok := scr.Credentials.(secret.SMTP); ok && !smtp.IsZero() {
			res = append(res, smtpCandidate{id: scr.ID, smtp: smtp})
		}
	}

	return res, nil
}

// pickCandidate returns the server whose id or name matches the hint or the first one. Candidates must not be empty.
func pickCandidate(candidates []smtpCandidate, idOrNameHint string) secret.SMTP {
	if idOrNameHint != "" {
		for _, c := range candidates {
			if string(c.id) == idOrNameHint || c.smtp.Name == idOrNameHint {
				return c.smtp
			}
		}
	}

	return candidates[0].smtp
}

// backoff returns base * 2^(attempts-1) limited to maxBackoff.
func backoff(base, maxBackoff time.Duration, attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}

	d := base
	for i := 1; i < attempts; i++ {
		d *= 2
		if d >= maxBackoff {
			return maxBackoff
		}
	}

	return min(d, maxBackoff)
}
