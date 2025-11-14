// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import (
	"fmt"
	"log/slog"

	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/application/sms/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/eventstore"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

func NewSend(repo message.Repository, providers *concurrent.RWMap[provider.ID, provider.Provider]) Send {
	return func(subject auth.Subject, sms message.SendRequested, opts SendOptions) (message.ID, error) {
		if err := subject.Audit(PermSend); err != nil {
			return "", err
		}

		var prov provider.Provider
		for id, p := range providers.All() {
			if prov == nil {
				prov = p
				continue
			}

			if provider.ID(sms.ProviderHint) == id || prov.Name() == sms.ProviderHint {
				prov = p
				break
			}
		}

		msg := message.SMS{
			ID:           message.ID(eventstore.NewID()),
			ProviderHint: sms.ProviderHint,
			Recipient:    sms.Recipient,
			Originator:   sms.Originator,
			Body:         sms.Body,
			CreatedAt:    xtime.Now(),
		}

		if prov == nil {
			if opts.NoQueue {
				return "", fmt.Errorf("no provider available")
			}

			msg := msg
			msg.LastError = "no provider available"
			msg.Status = message.StatusQueued

			if err := repo.Save(msg); err != nil {
				return "", err
			}

			return msg.ID, nil
		}

		// try to send immediately
		provSendId, err := prov.Send(subject, sms)
		if err != nil {
			if opts.NoQueue {
				return "", err
			}

			msg.LastError = err.Error()
			msg.Status = message.StatusQueued

			if err := repo.Save(msg); err != nil {
				return "", err
			}

			return msg.ID, nil
		}

		slog.Info("sms sent successfully", "id", msg.ID, "provider", prov.Identity())

		msg.LastError = ""
		msg.Status = message.StatusSent
		msg.SendAt = xtime.Now()
		msg.Provider = string(subject.ID())
		msg.ProviderMessage = provSendId
		
		if err := repo.Save(msg); err != nil {
			return "", err
		}

		return provSendId, nil
	}
}
