package mail

import (
	"context"
	"log/slog"
	"slices"
	"time"
)

type ScheduleOptions struct {
	SendInterval         time.Duration // default is 1 minute, within this interval new mails are checked and old mails are removed from queue
	KeepMailAfterSuccess time.Duration // default is 24 hours, if negative unlimited
	KeepMailAfterError   time.Duration // default is 1 year, if negative unlimited
	WaitBetweenSends     time.Duration // default is 2 Seconds
}

// StartScheduler starts a new scheduler instance to process the [Outgoing] mails.
func StartScheduler(ctx context.Context, opts ScheduleOptions, servers SmtpRepository, mails Repository) {
	if opts.SendInterval == 0 {
		opts.SendInterval = time.Minute
	}

	if opts.KeepMailAfterSuccess == 0 {
		opts.KeepMailAfterSuccess = time.Hour * 24
	}

	if opts.KeepMailAfterError == 0 {
		opts.KeepMailAfterError = time.Hour * 24 * 30 * 12
	}

	if opts.WaitBetweenSends == 0 {
		opts.WaitBetweenSends = time.Second * 2
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

			availableServers := make(map[SmtpID]Smtp)
			var primaryServer Smtp
			for smtp, err := range servers.All() {
				if err != nil {
					slog.Error("mail scheduler failed to iterate on smtp server repository", "err", err)
					continue
				}

				if smtp.Category == Disabled {
					continue
				}

				availableServers[smtp.ID] = smtp
				if primaryServer.Host == "" {
					primaryServer = smtp // in case of broken smtp data, just pick the first one
				}

				if smtp.Category == Primary {
					primaryServer = smtp
				}
			}

			if primaryServer.Host == "" {
				slog.Error("mail scheduler failed to find primary (or any at all) mail server")
				continue
			}

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

				srv, ok := availableServers[outgoing.Mail.SmtpHint]
				if !ok {
					srv = primaryServer
				}

				outgoing.Server = srv.ID
				outgoing.SendAt = time.Now()

				if err := srv.send(outgoing.Mail); err != nil {
					slog.Error("mail scheduler failed to send mail", "smtp", srv.ID, "id", outgoing.ID, "subject", outgoing.Mail.Subject, "err", err)

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
