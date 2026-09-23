// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"
)

// SpamGuardOptions configures the heuristic, which detects runaway send loops, e.g. caused by programming errors
// which enqueue the same mail over and over again. Suppressed mails get the [StatusSuppressed] state and are
// not sent, unless explicitly requested by [RetryOutgoing].
type SpamGuardOptions struct {
	// Disabled turns the guard off.
	Disabled bool
	// Window is the sliding time window of the heuristic. Default is 1 hour.
	Window time.Duration
	// MaxIdentical is the amount of identical mails (same recipients, subject and content) which may be sent
	// within the window. Default is 5.
	MaxIdentical int
	// MaxPerRecipient is the amount of mails to the same recipient address within the window. Default is 30.
	MaxPerRecipient int
}

func (o SpamGuardOptions) withDefaults() SpamGuardOptions {
	if o.Window <= 0 {
		o.Window = time.Hour
	}

	if o.MaxIdentical <= 0 {
		o.MaxIdentical = 5
	}

	if o.MaxPerRecipient <= 0 {
		o.MaxPerRecipient = 30
	}

	return o
}

// spamGuard is not thread safe and must only be used by the scheduler goroutine.
type spamGuard struct {
	opts        SpamGuardOptions
	identical   map[string][]time.Time
	byRecipient map[string][]time.Time
}

func newSpamGuard(opts SpamGuardOptions) *spamGuard {
	return &spamGuard{opts: opts.withDefaults(), identical: map[string][]time.Time{}, byRecipient: map[string][]time.Time{}}
}

func fingerprint(o Outgoing) string {
	h := sha256.New()
	for _, list := range [][]string{addrs(o.Mail.To), addrs(o.Mail.CC), addrs(o.Mail.BCC)} {
		h.Write([]byte(strings.Join(list, ",")))
		h.Write([]byte{0})
	}

	h.Write([]byte(o.Mail.Subject))
	h.Write([]byte{0})
	for _, p := range o.Mail.Parts {
		h.Write(p.Encoded)
		h.Write([]byte{0})
	}

	return hex.EncodeToString(h.Sum(nil))
}

func addrs(list []mail.Address) []string {
	res := make([]string, 0, len(list))
	for _, a := range list {
		res = append(res, strings.ToLower(strings.TrimSpace(a.Address)))
	}

	slices.Sort(res)
	return res
}

func recipientAddrs(o Outgoing) []string {
	res := append(append(addrs(o.Mail.To), addrs(o.Mail.CC)...), addrs(o.Mail.BCC)...)
	return slices.Compact(res)
}

func prune(list []time.Time, since time.Time) []time.Time {
	idx := 0
	for idx < len(list) && list[idx].Before(since) {
		idx++
	}

	return list[idx:]
}

// check returns a non-empty reason, if the mail shall be suppressed.
func (g *spamGuard) check(o Outgoing, now time.Time) string {
	if g.opts.Disabled || o.BypassSpamGuard {
		return ""
	}

	since := now.Add(-g.opts.Window)
	fp := fingerprint(o)
	g.identical[fp] = prune(g.identical[fp], since)
	if len(g.identical[fp]) >= g.opts.MaxIdentical {
		return fmt.Sprintf("spam guard: identical mail already sent %d times within %s, possible runaway send loop", len(g.identical[fp]), shortDuration(g.opts.Window))
	}

	for _, rcpt := range recipientAddrs(o) {
		g.byRecipient[rcpt] = prune(g.byRecipient[rcpt], since)
		if len(g.byRecipient[rcpt]) >= g.opts.MaxPerRecipient {
			return fmt.Sprintf("spam guard: %d mails sent to %s within %s, possible runaway send loop", len(g.byRecipient[rcpt]), rcpt, shortDuration(g.opts.Window))
		}
	}

	return ""
}

// record must be called after a successful delivery.
func (g *spamGuard) record(o Outgoing, now time.Time) {
	if g.opts.Disabled {
		return
	}

	fp := fingerprint(o)
	g.identical[fp] = append(g.identical[fp], now)
	for _, rcpt := range recipientAddrs(o) {
		g.byRecipient[rcpt] = append(g.byRecipient[rcpt], now)
	}

	// avoid unbounded growth of stale keys
	if len(g.identical) > 10_000 || len(g.byRecipient) > 10_000 {
		since := now.Add(-g.opts.Window)
		for _, m := range []map[string][]time.Time{g.identical, g.byRecipient} {
			for k, v := range m {
				if v = prune(v, since); len(v) == 0 {
					delete(m, k)
				} else {
					m[k] = v
				}
			}
		}
	}
}

// rateLimiter tracks the send attempts per smtp server within the last 24 hours. It is not thread safe and
// must only be used by the scheduler goroutine.
type rateLimiter struct {
	attempts map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{attempts: map[string][]time.Time{}}
}

// seed initializes the limiter from persisted statistics, so that a restart does not reset the limits.
func (r *rateLimiter) seed(stats StatsRepository, now time.Time) error {
	if stats == nil {
		return nil
	}

	since := now.Add(-24 * time.Hour)
	for b, err := range stats.All() {
		if err != nil {
			return err
		}

		if b.Hour.Add(time.Hour).Before(since) {
			continue
		}

		// we only know the hour, thus assume the latest possible point in time to be on the safe side
		at := b.Hour.Add(time.Hour - time.Second)
		if at.After(now) {
			at = now
		}
		for i := 0; i < b.Sent+b.Failed; i++ {
			r.attempts[b.Server] = append(r.attempts[b.Server], at)
		}
	}

	for k := range r.attempts {
		slices.SortFunc(r.attempts[k], func(a, b time.Time) int { return a.Compare(b) })
	}

	return nil
}

// allow returns true, if another attempt is allowed for the given server limits.
func (r *rateLimiter) allow(server string, perHour, perDay int, now time.Time) bool {
	if perHour <= 0 && perDay <= 0 {
		return true
	}

	list := prune(r.attempts[server], now.Add(-24*time.Hour))
	r.attempts[server] = list

	if perDay > 0 && len(list) >= perDay {
		return false
	}

	if perHour > 0 {
		since := now.Add(-time.Hour)
		n := 0
		for i := len(list) - 1; i >= 0 && !list[i].Before(since); i-- {
			n++
		}

		if n >= perHour {
			return false
		}
	}

	return true
}

func (r *rateLimiter) record(server string, now time.Time) {
	r.attempts[server] = append(r.attempts[server], now)
}

// shortDuration formats e.g. 1h0m0s as 1h.
func shortDuration(d time.Duration) string {
	str := d.String()
	if strings.HasSuffix(str, "m0s") {
		str = strings.TrimSuffix(str, "0s")
	}
	if strings.HasSuffix(str, "h0m") {
		str = strings.TrimSuffix(str, "0m")
	}
	return str
}
