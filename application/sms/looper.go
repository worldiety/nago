// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import (
	"context"
	"log/slog"
	"time"

	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xtime"
)

func loop(ctx context.Context, repo message.Repository, sendFn Send) {
	for ctx.Err() == nil {

		for sms, err := range repo.All() {
			if err != nil {
				slog.Error("failed to load sms from repository", "err", err.Error())
				continue
			}

			if sms.Status == message.StatusSent {
				// we retain sms history 7 days for successful messages, others are kept infinite
				retentionDuration := time.Hour * 24 * 7
				retain := time.UnixMilli(int64(sms.SendAt)).Add(retentionDuration).After(time.Now())
				if !retain {
					slog.Info("purge sent sms from history after retention period", "duration", retentionDuration)
					if err := repo.DeleteByID(sms.ID); err != nil {
						slog.Error("failed to delete sms from repository", "err", err.Error())
					}
				}

				continue
			}

			id, err := sendFn(user.SU(), message.SendRequested{
				ProviderHint: sms.ProviderHint,
				Recipient:    sms.Recipient,
				Originator:   sms.Originator,
				Body:         sms.Body,
			}, SendOptions{NoQueue: true})
			if err != nil {
				sms.LastError = err.Error()
				sms.Status = message.StatusFailed
				if err := repo.Save(sms); err != nil {
					slog.Error("failed to save sms to repository", "err", err.Error())
				}

				slog.Error("failed to send sms", "err", err.Error(), "id", sms.ID)
				continue
			}

			sms.ProviderMessage = id
			sms.Status = message.StatusSent
			sms.SendAt = xtime.Now()
			sms.LastError = ""

			if err := repo.Save(sms); err != nil {
				slog.Error("failed to save sms to repository", "err", err.Error())
			}
		}
	}
	time.Sleep(time.Minute)
}
