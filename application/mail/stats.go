// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/data"
)

// StatsBucketID is formatted as <RFC3339 hour in UTC>|<server name>.
type StatsBucketID string

// StatsBucket aggregates the send attempts of a single hour and smtp server. Buckets are kept independently
// of the retention time of the [Outgoing] mails.
type StatsBucket struct {
	ID          StatsBucketID `json:"id"`
	Hour        time.Time     `json:"hour"` // truncated to the hour, UTC
	Server      string        `json:"server,omitempty"`
	Sent        int           `json:"sent,omitempty"`        // successful deliveries
	Failed      int           `json:"failed,omitempty"`      // failed attempts
	Retries     int           `json:"retries,omitempty"`     // attempts which were not the first attempt
	LatencySum  time.Duration `json:"latencySum,omitempty"`  // sum of queued-to-sent durations of successful deliveries
	DurationSum time.Duration `json:"durationSum,omitempty"` // sum of the smtp conversation durations
}

func (b StatsBucket) Identity() StatsBucketID {
	return b.ID
}

func (b StatsBucket) WithIdentity(id StatsBucketID) StatsBucket {
	b.ID = id
	return b
}

type StatsRepository data.Repository[StatsBucket, StatsBucketID]

// ServerHealth tracks the latest state of an smtp server as seen by the scheduler.
type ServerHealth struct {
	ID                  string    `json:"id"` // the server name
	LastSuccessAt       time.Time `json:"lastSuccessAt,omitzero"`
	LastErrorAt         time.Time `json:"lastErrorAt,omitzero"`
	LastError           string    `json:"lastError,omitempty"`
	LastErrorPhase      Phase     `json:"lastErrorPhase,omitempty"`
	LastErrorCode       int       `json:"lastErrorCode,omitempty"`
	ConsecutiveFailures int       `json:"consecutiveFailures,omitempty"`
}

func (h ServerHealth) Identity() string {
	return h.ID
}

func (h ServerHealth) WithIdentity(id string) ServerHealth {
	h.ID = id
	return h
}

type HealthRepository data.Repository[ServerHealth, string]

// SchedulerStatus is a volatile snapshot of the scheduler.
type SchedulerStatus struct {
	Running        bool
	LastRunAt      time.Time
	NoSmtpServer   bool // true, if the last run could not find any smtp credentials
	LastSmtpErrMsg string
}

var schedulerStatus atomic.Pointer[SchedulerStatus]

func setSchedulerStatus(fn func(s *SchedulerStatus)) {
	var s SchedulerStatus
	if cur := schedulerStatus.Load(); cur != nil {
		s = *cur
	}
	fn(&s)
	schedulerStatus.Store(&s)
}

func currentSchedulerStatus() SchedulerStatus {
	if s := schedulerStatus.Load(); s != nil {
		return *s
	}

	return SchedulerStatus{}
}

func statsBucketID(hour time.Time, server string) StatsBucketID {
	return StatsBucketID(hour.UTC().Format(time.RFC3339) + "|" + server)
}

// statsRecorder updates statistics and health. All repositories are optional.
type statsRecorder struct {
	mutex  sync.Mutex
	stats  StatsRepository
	health HealthRepository
}

func (r *statsRecorder) record(out Outgoing, a Attempt) {
	if r == nil {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.stats != nil {
		hour := a.At.UTC().Truncate(time.Hour)
		id := statsBucketID(hour, a.Server)
		optB, err := r.stats.FindByID(id)
		if err != nil {
			slog.Error("mail stats: cannot load bucket", "id", id, "err", err)
		} else {
			b := optB.UnwrapOr(StatsBucket{ID: id, Hour: hour, Server: a.Server})
			if a.Success {
				b.Sent++
				if !out.QueuedAt.IsZero() {
					b.LatencySum += a.At.Add(a.Duration).Sub(out.QueuedAt)
				}
			} else {
				b.Failed++
			}

			if out.AttemptCount > 1 {
				b.Retries++
			}

			b.DurationSum += a.Duration

			if err := r.stats.Save(b); err != nil {
				slog.Error("mail stats: cannot save bucket", "id", id, "err", err)
			}
		}
	}

	if r.health != nil && a.Server != "" {
		optH, err := r.health.FindByID(a.Server)
		if err != nil {
			slog.Error("mail stats: cannot load server health", "server", a.Server, "err", err)
			return
		}

		h := optH.UnwrapOr(ServerHealth{ID: a.Server})
		if a.Success {
			h.LastSuccessAt = a.At
			h.ConsecutiveFailures = 0
		} else {
			h.LastErrorAt = a.At
			h.LastError = a.Message
			h.LastErrorPhase = a.Phase
			h.LastErrorCode = a.Code
			h.ConsecutiveFailures++
		}

		if err := r.health.Save(h); err != nil {
			slog.Error("mail stats: cannot save server health", "server", a.Server, "err", err)
		}
	}
}
