// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"cmp"
	"slices"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

// StatisticsOptions defines the time range of the requested statistics.
type StatisticsOptions struct {
	From time.Time
	To   time.Time
	// StuckAfter defines when an unsent mail is considered as stuck. Default is 1 hour.
	StuckAfter time.Duration
}

// Totals summarizes the attempts of a time range.
type Totals struct {
	Sent       int
	Failed     int // failed attempts
	Retries    int
	LatencySum time.Duration
}

func (t *Totals) add(b StatsBucket) {
	t.Sent += b.Sent
	t.Failed += b.Failed
	t.Retries += b.Retries
	t.LatencySum += b.LatencySum
}

// ErrorRate returns the ratio of failed attempts to all attempts in [0..1].
func (t Totals) ErrorRate() float64 {
	n := t.Sent + t.Failed
	if n == 0 {
		return 0
	}

	return float64(t.Failed) / float64(n)
}

// AvgLatency returns the average queued-to-sent duration.
func (t Totals) AvgLatency() time.Duration {
	if t.Sent == 0 {
		return 0
	}

	return t.LatencySum / time.Duration(t.Sent)
}

// QueueCounts is a snapshot of the current queue.
type QueueCounts struct {
	Queued  int
	Error   int // temporary errors, will be retried
	Failed  int // gave up
	Success int // still kept in queue
	Stuck   int // unsent and older than [StatisticsOptions.StuckAfter]
}

func (q QueueCounts) Total() int {
	return q.Queued + q.Error + q.Failed + q.Success
}

// ServerInfo describes a configured smtp server without exposing any credentials.
type ServerInfo struct {
	SecretID secret.ID
	Name     string
	Host     string
	Port     int
	Health   ServerHealth
	Totals   Totals // within the requested time range
}

// ErrorCount counts equal error messages of mails in the queue.
type ErrorCount struct {
	Message string
	Phase   Phase
	Code    int
	Count   int
}

type StatisticsResult struct {
	Options        StatisticsOptions
	Totals         Totals
	Previous       Totals        // the same duration immediately before From
	Buckets        []StatsBucket // hourly, all servers, sorted by time
	Queue          QueueCounts
	Servers        []ServerInfo
	TopErrors      []ErrorCount
	RecentFailures []Outgoing // latest 5 unsuccessful mails
	Scheduler      SchedulerStatus
	// HasStats is false, if no statistics have been recorded yet, e.g. directly after the update.
	HasStats bool
}

// Statistics aggregates statistics, problems and server health of the mail system.
type Statistics func(subject auth.Subject, opts StatisticsOptions) (StatisticsResult, error)

func NewStatistics(repo Repository, stats StatsRepository, health HealthRepository, sysUser user.SysUser, secrets secret.FindGroupSecrets) Statistics {
	return func(subject auth.Subject, opts StatisticsOptions) (StatisticsResult, error) {
		if err := subject.Audit(PermStatistics); err != nil {
			return StatisticsResult{}, err
		}

		now := time.Now()
		if opts.To.IsZero() {
			opts.To = now
		}

		if opts.From.IsZero() {
			opts.From = opts.To.Add(-7 * 24 * time.Hour)
		}

		if opts.StuckAfter == 0 {
			opts.StuckAfter = time.Hour
		}

		res := StatisticsResult{Options: opts, Scheduler: currentSchedulerStatus()}
		prevFrom := opts.From.Add(-opts.To.Sub(opts.From))
		perServer := map[string]Totals{}

		if stats != nil {
			for b, err := range stats.All() {
				if err != nil {
					return res, err
				}

				res.HasStats = true
				switch {
				case !b.Hour.Before(opts.From.Truncate(time.Hour)) && b.Hour.Before(opts.To):
					res.Totals.add(b)
					res.Buckets = append(res.Buckets, b)
					t := perServer[b.Server]
					t.add(b)
					perServer[b.Server] = t
				case !b.Hour.Before(prevFrom) && b.Hour.Before(opts.From):
					res.Previous.add(b)
				}
			}
		}

		slices.SortFunc(res.Buckets, func(a, b StatsBucket) int { return a.Hour.Compare(b.Hour) })

		errCounts := map[string]*ErrorCount{}
		var failures []Outgoing
		for o, err := range repo.All() {
			if err != nil {
				return res, err
			}

			switch o.Status {
			case StatusSendSuccess:
				res.Queue.Success++
			case StatusError:
				res.Queue.Error++
			case StatusFailed:
				res.Queue.Failed++
			default:
				res.Queue.Queued++
			}

			if o.Status != StatusSendSuccess && o.Status != StatusFailed && now.Sub(o.QueuedAt) > opts.StuckAfter {
				res.Queue.Stuck++
			}

			if o.Status == StatusError || o.Status == StatusFailed {
				failures = append(failures, o)
				ec := ErrorCount{Message: o.LastError}
				if a, ok := o.LastAttempt(); ok && !a.Success {
					ec = ErrorCount{Message: a.Message, Phase: a.Phase, Code: a.Code}
				}

				key := string(ec.Phase) + "|" + ec.Message
				if c, ok := errCounts[key]; ok {
					c.Count++
				} else {
					ec.Count = 1
					errCounts[key] = &ec
				}
			}
		}

		slices.SortFunc(failures, func(a, b Outgoing) int { return cmp.Compare(b.SendAt.UnixNano(), a.SendAt.UnixNano()) })
		if len(failures) > 5 {
			failures = failures[:5]
		}

		for i := range failures {
			failures[i].Mail.Parts = nil // not required and may be large
		}

		res.RecentFailures = failures

		for _, c := range errCounts {
			res.TopErrors = append(res.TopErrors, *c)
		}

		slices.SortFunc(res.TopErrors, func(a, b ErrorCount) int { return cmp.Compare(b.Count, a.Count) })
		if len(res.TopErrors) > 5 {
			res.TopErrors = res.TopErrors[:5]
		}

		if secrets != nil {
			for scr, err := range secrets(sysUser(), group.System) {
				if err != nil {
					return res, err
				}

				smtp, ok := scr.Credentials.(secret.SMTP)
				if !ok {
					continue
				}

				info := ServerInfo{SecretID: scr.ID, Name: smtp.Name, Host: smtp.Host, Port: smtp.Port, Totals: perServer[smtp.Name]}
				if health != nil {
					optH, err := health.FindByID(smtp.Name)
					if err != nil {
						return res, err
					}

					info.Health = optH.UnwrapOr(ServerHealth{ID: smtp.Name})
				}

				res.Servers = append(res.Servers, info)
			}
		}

		return res, nil
	}
}
