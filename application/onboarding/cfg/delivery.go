// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding

import (
	"log/slog"
	netmail "net/mail"
	"slices"
	"strings"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/mail/nms"
	"go.wdy.de/nago/application/onboarding"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// mailDelivery checks the transports of the mail scheduler like it does: an SMTP secret shared with the system group
// or the Nago Mail Service. The service may be nil.
func mailDelivery(sysUser func() auth.Subject, findGroupSecrets secret.FindGroupSecrets, svc *nms.Service) onboarding.CheckDelivery {
	return func() onboarding.Delivery {
		if hasSMTP(sysUser(), findGroupSecrets) {
			return onboarding.Delivery{Level: onboarding.DeliveryAssured}
		}

		if svc == nil {
			return onboarding.Delivery{Level: onboarding.DeliveryNone}
		}

		st := svc.Status()
		if st.Enrolled {
			return onboarding.Delivery{Level: onboarding.DeliveryAssured}
		}

		delivery := onboarding.Delivery{Level: onboarding.DeliveryNone, ExchangeFailedAt: st.ExchangeFailedAt, NextExchangeAt: st.NextExchangeAt}
		if svc.Available() {
			delivery.Level = onboarding.DeliveryTryable
		}

		return delivery
	}
}

// hasSMTP tells, if an SMTP secret, which is not empty, is shared with the system group. A secret, which cannot be
// read, does not count.
func hasSMTP(subject auth.Subject, findGroupSecrets secret.FindGroupSecrets) bool {
	for scr, err := range findGroupSecrets(subject, group.System) {
		if err != nil {
			slog.Error("onboarding: cannot read the secrets of the system group", "err", err)
			return false
		}

		if smtp, ok := scr.Credentials.(secret.SMTP); ok && !smtp.IsZero() {
			return true
		}
	}

	return false
}

// mailCodeStatus reads the newest outgoing mail to the address since the code has been sent. It returns only its
// status.
func mailCodeStatus(sysUser func() auth.Subject, findIDs mail.FindOutgoingIDs, findByID mail.FindOutgoingByID) onboarding.ReadCodeStatus {
	return func(to user.Email, sentAt time.Time) onboarding.CodeStatus {
		if findIDs == nil || findByID == nil || to == "" {
			return onboarding.CodeStatusUnknown
		}

		subject := sysUser()
		// the newest mail first
		for id, err := range findIDs(subject, mail.OutgoingFilter{Since: sentAt}) {
			if err != nil {
				slog.Error("onboarding: cannot read the outgoing mails", "err", err)
				return onboarding.CodeStatusUnknown
			}

			optMail, err := findByID(subject, id)
			if err != nil {
				slog.Error("onboarding: cannot read an outgoing mail", "err", err)
				return onboarding.CodeStatusUnknown
			}

			if optMail.IsNone() || !addressedTo(optMail.Unwrap().Receiver, string(to)) {
				continue
			}

			switch optMail.Unwrap().Status {
			case mail.StatusQueued:
				return onboarding.CodeStatusQueued
			case mail.StatusSendSuccess:
				return onboarding.CodeStatusAccepted
			case mail.StatusError:
				return onboarding.CodeStatusRetrying
			case mail.StatusFailed, mail.StatusSuppressed:
				return onboarding.CodeStatusFailed
			default:
				return onboarding.CodeStatusUnknown
			}
		}

		return onboarding.CodeStatusUnknown
	}
}

// addressedTo tells, if the receivers of an outgoing mail contain the address, regardless of the case.
func addressedTo(receiver, to string) bool {
	list, err := netmail.ParseAddressList(receiver)
	if err != nil {
		return strings.EqualFold(strings.TrimSpace(receiver), to)
	}

	for _, addr := range list {
		if strings.EqualFold(addr.Address, to) {
			return true
		}
	}

	return false
}

// exemptRoutes are the internal routes of the legal texts of the theme and the options and the routes of
// [Options.Exempt]. An external address with "://" is no route.
func exemptRoutes(themeSettings theme.Settings, opts Options) []core.NavigationPath {
	var res []core.NavigationPath
	for _, v := range []string{
		themeSettings.Impress,
		themeSettings.PrivacyPolicy,
		themeSettings.GeneralTermsAndConditions,
		themeSettings.TermsOfUse,
		opts.Impress,
		opts.PrivacyPolicy,
	} {
		v = strings.TrimSpace(v)
		if v == "" || strings.Contains(v, "://") {
			continue
		}

		if route := core.NavigationPath(strings.TrimPrefix(v, "/")); !slices.Contains(res, route) {
			res = append(res, route)
		}
	}

	for _, route := range opts.Exempt {
		if !slices.Contains(res, route) {
			res = append(res, route)
		}
	}

	return res
}
