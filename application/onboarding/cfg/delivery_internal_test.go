// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding

import (
	"context"
	"errors"
	"iter"
	"slices"
	"testing"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/mail/nms"
	"go.wdy.de/nago/application/mail/nms/nmstest"
	"go.wdy.de/nago/application/onboarding"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	datamem "go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
)

// otherCredentials are no SMTP server.
type otherCredentials struct{ Name string }

func (o otherCredentials) GetName() string   { return o.Name }
func (o otherCredentials) Credentials() bool { return true }
func (o otherCredentials) IsZero() bool      { return o == otherCredentials{} }

// groupSecrets returns the secrets only for the system group, like the real use case for a secret shared with it.
func groupSecrets(err error, credentials ...secret.Credentials) secret.FindGroupSecrets {
	return func(subject auth.Subject, gid group.ID) iter.Seq2[secret.Secret, error] {
		return func(yield func(secret.Secret, error) bool) {
			if gid != group.System {
				return
			}

			if err != nil {
				yield(secret.Secret{}, err)
				return
			}

			for _, c := range credentials {
				if !yield(secret.Secret{ID: data.RandIdent[secret.ID](), Groups: []group.ID{group.System}, Credentials: c}, nil) {
					return
				}
			}
		}
	}
}

const origin = "https://onboarding.example.com"

// mailService returns the Nago Mail Service in the given condition.
func mailService(t *testing.T, condition string) *nms.Service {
	switch condition {
	case "":
		return nil
	case "disabled":
		return nms.NewService(nms.Options{})
	}

	srv := nmstest.NewServer(t)
	srv.VerifyOrigin = func(origin, nonce string) bool { return false }
	opts := nms.Options{Endpoint: srv.URL, Origin: func() string { return origin }, Nonces: nms.NewNonces()}
	if condition == "enrolled" {
		opts.Token = srv.IssueRefresh("tenant", "onboarding.example.com")
	}

	svc := nms.NewService(opts)
	if condition == "failed" {
		// the scheduler sends the first mail, which tries the exchange
		if _, err := svc.Send(context.Background(), nms.Message{From: nms.Address{Email: "noreply@onboarding.example.com"}, To: []nms.Address{{Email: "a@example.com"}}}); err == nil {
			t.Fatal("the exchange must fail")
		}
	}

	return svc
}

// The delivery of the code checks the transports of the mail scheduler: an SMTP server, which is not empty, shared
// with the system group, or the Nago Mail Service.
func TestMailDelivery(t *testing.T) {
	smtp := secret.SMTP{Name: "Mailserver", Host: "smtp.example.com", Port: 587}
	for _, tc := range []struct {
		name        string
		secrets     secret.FindGroupSecrets
		service     string // "", "disabled", "available", "enrolled" or "failed"
		want        onboarding.DeliveryLevel
		wantFailure bool
	}{
		{name: "nothing", secrets: groupSecrets(nil), want: onboarding.DeliveryNone},
		{name: "smtp", secrets: groupSecrets(nil, smtp), want: onboarding.DeliveryAssured},
		{name: "empty smtp", secrets: groupSecrets(nil, secret.SMTP{}), want: onboarding.DeliveryNone},
		{name: "other secret", secrets: groupSecrets(nil, otherCredentials{Name: "x"}), want: onboarding.DeliveryNone},
		{name: "unreadable secrets", secrets: groupSecrets(errors.New("broken"), smtp), want: onboarding.DeliveryNone},
		{name: "disabled mail service", secrets: groupSecrets(nil), service: "disabled", want: onboarding.DeliveryNone},
		{name: "available mail service", secrets: groupSecrets(nil), service: "available", want: onboarding.DeliveryTryable},
		{name: "enrolled mail service", secrets: groupSecrets(nil), service: "enrolled", want: onboarding.DeliveryAssured},
		{name: "failed exchange", secrets: groupSecrets(nil), service: "failed", want: onboarding.DeliveryNone, wantFailure: true},
		{name: "smtp and failed exchange", secrets: groupSecrets(nil, smtp), service: "failed", want: onboarding.DeliveryAssured},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := mailService(t, tc.service)
			got := mailDelivery(func() auth.Subject { return user.SU() }, tc.secrets, svc)()
			if got.Level != tc.want {
				t.Fatalf("expected %v, got %+v", tc.want, got)
			}

			if !tc.wantFailure {
				if !got.ExchangeFailedAt.IsZero() || !got.NextExchangeAt.IsZero() {
					t.Fatalf("expected no failed exchange, got %+v", got)
				}

				return
			}

			st := svc.Status()
			if got.ExchangeFailedAt.IsZero() || !got.ExchangeFailedAt.Equal(st.ExchangeFailedAt) || !got.NextExchangeAt.Equal(st.NextExchangeAt) {
				t.Fatalf("expected the times of the service %v and %v, got %+v", st.ExchangeFailedAt, st.NextExchangeAt, got)
			}
		})
	}
}

// The status of the code is the one of the newest mail to the address since the code has been sent. Mails to other
// addresses and older mails do not count.
func TestMailCodeStatus(t *testing.T) {
	sentAt := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	type outgoing struct {
		receiver string
		status   mail.Status
		queued   time.Duration // after the code has been sent
	}

	for _, tc := range []struct {
		name  string
		mails []outgoing
		want  onboarding.CodeStatus
	}{
		{name: "no mail", want: onboarding.CodeStatusUnknown},
		{name: "queued", mails: []outgoing{{"<torben@example.com>", mail.StatusQueued, 0}}, want: onboarding.CodeStatusQueued},
		{name: "sent", mails: []outgoing{{"<torben@example.com>", mail.StatusSendSuccess, 0}}, want: onboarding.CodeStatusAccepted},
		{name: "error", mails: []outgoing{{"<torben@example.com>", mail.StatusError, 0}}, want: onboarding.CodeStatusRetrying},
		{name: "failed", mails: []outgoing{{"<torben@example.com>", mail.StatusFailed, 0}}, want: onboarding.CodeStatusFailed},
		{name: "suppressed", mails: []outgoing{{"<torben@example.com>", mail.StatusSuppressed, 0}}, want: onboarding.CodeStatusFailed},
		{name: "other status", mails: []outgoing{{"<torben@example.com>", mail.StatusUndefined, 0}}, want: onboarding.CodeStatusUnknown},
		{name: "other address", mails: []outgoing{{"<anna@example.com>", mail.StatusSendSuccess, 0}}, want: onboarding.CodeStatusUnknown},
		{name: "name and case", mails: []outgoing{{`"Torben" <TORBEN@Example.com>, <anna@example.com>`, mail.StatusSendSuccess, 0}}, want: onboarding.CodeStatusAccepted},
		{name: "before the code", mails: []outgoing{{"<torben@example.com>", mail.StatusSendSuccess, -time.Second}}, want: onboarding.CodeStatusUnknown},
		{name: "newest to the address", mails: []outgoing{
			{"<torben@example.com>", mail.StatusSendSuccess, time.Second},
			{"<torben@example.com>", mail.StatusError, 2 * time.Second},
			{"<anna@example.com>", mail.StatusFailed, 3 * time.Second},
		}, want: onboarding.CodeStatusRetrying},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &datamem.Repository[mail.Outgoing, mail.ID]{}
			for i, m := range tc.mails {
				id := mail.ID(string(rune('a' + i)))
				if err := repo.Save(mail.Outgoing{ID: id, Receiver: m.receiver, Status: m.status, QueuedAt: sentAt.Add(m.queued)}); err != nil {
					t.Fatal(err)
				}
			}

			read := mailCodeStatus(func() auth.Subject { return user.SU() }, mail.NewFindOutgoingIDs(repo), mail.NewFindOutgoingByID(repo))
			if got := read("torben@example.com", sentAt); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// A mail, which cannot be read, has an unknown status, while the same mail would be shown otherwise.
func TestMailCodeStatusUnreadable(t *testing.T) {
	sentAt := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	repo := &datamem.Repository[mail.Outgoing, mail.ID]{}
	if err := repo.Save(mail.Outgoing{ID: "a", Receiver: "<torben@example.com>", Status: mail.StatusSendSuccess, QueuedAt: sentAt}); err != nil {
		t.Fatal(err)
	}

	sys := func() auth.Subject { return user.SU() }
	findIDs, findByID := mail.NewFindOutgoingIDs(repo), mail.NewFindOutgoingByID(repo)
	broken := errors.New("broken")
	for _, tc := range []struct {
		name     string
		findIDs  mail.FindOutgoingIDs
		findByID mail.FindOutgoingByID
		want     onboarding.CodeStatus
	}{
		{name: "readable", findIDs: findIDs, findByID: findByID, want: onboarding.CodeStatusAccepted},
		{name: "list unreadable", findIDs: func(auth.Subject, mail.OutgoingFilter) iter.Seq2[mail.ID, error] {
			return func(yield func(mail.ID, error) bool) { yield("", broken) }
		}, findByID: findByID, want: onboarding.CodeStatusUnknown},
		{name: "mail unreadable", findIDs: findIDs, findByID: func(auth.Subject, mail.ID) (std.Option[mail.Outgoing], error) {
			return std.None[mail.Outgoing](), broken
		}, want: onboarding.CodeStatusUnknown},
		{name: "no use cases", want: onboarding.CodeStatusUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mailCodeStatus(sys, tc.findIDs, tc.findByID)("torben@example.com", sentAt); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestAddressedTo(t *testing.T) {
	for _, tc := range []struct {
		receiver string
		want     bool
	}{
		{"<torben@example.com>", true},
		{"Torben <TORBEN@EXAMPLE.COM>", true},
		{"<anna@example.com>, <torben@example.com>", true},
		{"<anna@example.com>", false},
		{"<torben@example.com.evil>", false},
		{"Torben..S@example.com", false}, // no address list and not the same
		{"A..B@Example.com", true},       // no address list, but the same
		{"", false},
	} {
		to := "torben@example.com"
		if tc.receiver == "A..B@Example.com" {
			to = "a..b@example.com"
		}

		if got := addressedTo(tc.receiver, to); got != tc.want {
			t.Fatalf("addressedTo(%q, %q) = %v, want %v", tc.receiver, to, got, tc.want)
		}
	}
}

// The internal routes of the legal texts are exempt, an external address is no route.
func TestExemptRoutes(t *testing.T) {
	got := exemptRoutes(theme.Settings{
		Impress:                   " /impressum ",
		PrivacyPolicy:             "https://example.org/privacy",
		GeneralTermsAndConditions: "page/agb",
		TermsOfUse:                "http://localhost:3000/terms",
	}, Options{
		Impress:       "https://example.org/imprint",
		PrivacyPolicy: "datenschutz",
		Exempt:        []core.NavigationPath{"extra", "impressum"},
	})

	want := []core.NavigationPath{"impressum", "page/agb", "datenschutz", "extra"}
	if !slices.Equal(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}

	if got := exemptRoutes(theme.Settings{}, Options{}); len(got) != 0 {
		t.Fatalf("without legal texts, nothing is exempt: %v", got)
	}
}
