package mail

import (
	"context"
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

	go func() {
		slog.Info("mail scheduler started")
		for {
			select {
			case <-ctx.Done():
				slog.Info("mail scheduler stopped")
				return
			default:
				// continue below
			}

			time.Sleep(opts.SendInterval)

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
					if outgoing.Status == StatusError && now.Sub(outgoing.QueuedAt) > opts.KeepMailAfterError {
						toRemove = append(toRemove, outgoing.ID)
					}
				}

				// go next, if nothing to do
				if outgoing.Status == StatusSendSuccess {
					continue
				}

				optSmtp, err := pickMailServerCandidate(sysUser, secrets, outgoing.Mail.SmtpHint)
				if err != nil {
					slog.Error("mail scheduler failed to pick mail server", "err", err)
					break
				}

				if optSmtp.IsNone() {
					slog.Error("cannot process mail queue, no smtp credentials available in system group")
					break
				}

				smtp := optSmtp.Unwrap()

				outgoing.ServerName = smtp.Name
				outgoing.SendAt = time.Now()

				if err := send(smtp, outgoing.Mail); err != nil {
					slog.Error("mail scheduler failed to send mail", "smtp", smtp.Name, "id", outgoing.ID, "subject", outgoing.Mail.Subject, "err", err)

					outgoing.Status = StatusError
					outgoing.LastError = err.Error()
				} else {
					slog.Info("mail scheduler send mail success", "id", outgoing.ID)
					outgoing.Status = StatusSendSuccess
					outgoing.LastError = ""
				}

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
