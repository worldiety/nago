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
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"log/slog"
	"slices"
	"time"
)

type ScheduleOptions struct {
	SendInterval         time.Duration // default is 30 seconds, within this interval new mails are checked and old mails are removed from queue
	KeepMailAfterSuccess time.Duration // default is 24 hours, if negative unlimited
	KeepMailAfterError   time.Duration // default is 1 year, if negative unlimited
	WaitBetweenSends     time.Duration // default is 1 Seconds

	// MaxAttempts is the amount of attempts, before a mail is marked as [StatusFailed]. Default is 12, if negative
	// unlimited. Permanent SMTP rejections (5xx) of recipients fail immediately.
	MaxAttempts int

	// MaxBackoff limits the exponential backoff between attempts of the same mail. Default is 1 hour.
	MaxBackoff time.Duration

	// Stats is optional and receives hourly aggregated statistics.
	Stats StatsRepository
	// Health is optional and receives the latest state per smtp server.
	Health HealthRepository
}

// StartScheduler starts a new scheduler instance to process the [Outgoing] mails.
func StartScheduler(ctx context.Context, opts ScheduleOptions, mails Repository, sysUser user.SysUser, secrets secret.FindGroupSecrets) {
	if opts.SendInterval == 0 {
		opts.SendInterval = time.Second * 30
	}

	if opts.KeepMailAfterSuccess == 0 {
		opts.KeepMailAfterSuccess = time.Hour * 24
	}

	if opts.KeepMailAfterError == 0 {
		opts.KeepMailAfterError = time.Hour * 24 * 30 * 12
	}

	if opts.WaitBetweenSends == 0 {
		opts.WaitBetweenSends = time.Second * 1
	}

	if opts.MaxAttempts == 0 {
		opts.MaxAttempts = 12
	}

	if opts.MaxBackoff == 0 {
		opts.MaxBackoff = time.Hour
	}

	recorder := &statsRecorder{stats: opts.Stats, health: opts.Health}
	setSchedulerStatus(func(s *SchedulerStatus) { s.Running = true })

	go func() {
		slog.Info("mail scheduler started")
		for {
			select {
			case <-ctx.Done():
				slog.Info("mail scheduler stopped")
				setSchedulerStatus(func(s *SchedulerStatus) { s.Running = false })
				return
			default:
				// continue below
			}

			time.Sleep(opts.SendInterval)
			setSchedulerStatus(func(s *SchedulerStatus) { s.LastRunAt = time.Now() })

			if n, err := mails.Count(); err != nil || n == 0 {
				if err != nil {
					slog.Error("mail scheduler cannot count mail queue", "err", err)
					continue
				}

				if n == 0 {
					// nothing to do, don't need to check for smtp
					continue
				}
			}

			// TODO implement per hour/day rate limit, usually 500 mails per day
			// TODO implement wrong auth detection, otherwise we may try every second with wrong credentials and our IP is likely blocked
			// TODO implement global settings

			now := time.Now()
			var toRemove []ID
			for outgoing, err := range mails.All() {
				if err != nil {
					slog.Error("mail scheduler failed to iterate on outgoing mail repository", "err", err)
					continue
				}

				keepSuccessUnlimited := opts.KeepMailAfterSuccess < 0
				if !keepSuccessUnlimited {
					if outgoing.Status == StatusSendSuccess && now.Sub(outgoing.QueuedAt) > opts.KeepMailAfterSuccess {
						toRemove = append(toRemove, outgoing.ID)
					}
				}

				keepErrorUnlimited := opts.KeepMailAfterError < 0
				if !keepErrorUnlimited {
					if (outgoing.Status == StatusError || outgoing.Status == StatusFailed) && now.Sub(outgoing.QueuedAt) > opts.KeepMailAfterError {
						toRemove = append(toRemove, outgoing.ID)
					}
				}

				// go next, if nothing to do
				if outgoing.Status == StatusSendSuccess || outgoing.Status == StatusFailed {
					continue
				}

				if !outgoing.NextAttemptAt.IsZero() && now.Before(outgoing.NextAttemptAt) {
					continue
				}

				optSmtp, err := pickMailServerCandidate(sysUser, secrets, outgoing.Mail.SmtpHint)
				if err != nil {
					slog.Error("mail scheduler failed to pick mail server", "err", err)
					setSchedulerStatus(func(s *SchedulerStatus) { s.LastSmtpErrMsg = err.Error() })
					break
				}

				if optSmtp.IsNone() {
					slog.Error("cannot process mail queue, no smtp credentials available in system group")
					setSchedulerStatus(func(s *SchedulerStatus) { s.NoSmtpServer = true })
					break
				}

				setSchedulerStatus(func(s *SchedulerStatus) { s.NoSmtpServer = false; s.LastSmtpErrMsg = "" })

				smtp := optSmtp.Unwrap()

				start := time.Now()
				sendErr := send(smtp, outgoing.Mail)
				attempt := Attempt{
					At:       start,
					Server:   smtp.Name,
					Success:  sendErr == nil,
					Duration: time.Since(start),
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
					if permanent || (opts.MaxAttempts > 0 && outgoing.AttemptCount >= opts.MaxAttempts) {
						outgoing.Status = StatusFailed
						outgoing.NextAttemptAt = time.Time{}
					} else {
						outgoing.NextAttemptAt = time.Now().Add(backoff(opts.SendInterval, opts.MaxBackoff, outgoing.AttemptCount))
					}
				} else {
					slog.Info("mail scheduler send mail success", "id", outgoing.ID)
					outgoing.Status = StatusSendSuccess
					outgoing.LastError = ""
					outgoing.SentAt = time.Now()
					outgoing.NextAttemptAt = time.Time{}
				}

				recorder.record(outgoing, attempt)

				// this is an anti-spam heuristic
				time.Sleep(opts.WaitBetweenSends)

				if err := mails.Save(outgoing); err != nil {
					slog.Error("failed to save outgoing mail state", "id", outgoing.ID, "subject", outgoing.Mail.Subject, "err", err)
					continue
				}
			}

			if len(toRemove) > 0 {
				if err := mails.DeleteAllByID(slices.Values(toRemove)); err != nil {
					slog.Error("mail scheduler failed to remove mails", "err", err)
					continue
				}
			}

		}
	}()
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

func pickMailServerCandidate(sysUser user.SysUser, secrets secret.FindGroupSecrets, idOrNameHint string) (std.Option[secret.SMTP], error) {
	var bestMatch secret.SMTP
	for scr, err := range secrets(sysUser(), group.System) {
		if err != nil {
			return std.None[secret.SMTP](), err
		}

		if smtp, ok := scr.Credentials.(secret.SMTP); ok {
			if bestMatch.IsZero() {
				bestMatch = smtp
			} else {
				if string(scr.ID) == idOrNameHint || smtp.Name == idOrNameHint {
					bestMatch = smtp
				}
			}
		}

	}

	if bestMatch.IsZero() {
		return std.None[secret.SMTP](), nil
	}

	return std.Some(bestMatch), nil
}
