// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser_test

import (
	"fmt"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/proto"
)

// The message after an action in the tab "Sonstiges" of an account names the account by its name and mail
// address and shows nothing else of it. Especially it never shows a pending verification code, because whoever
// knows the code confirms the mail address through the confirmation page and sets the password there, if the
// account has none yet. Each row proves its premise: the code is pending in the account, when the button is
// clicked, and the message is shown, so that a missing code is not a missing message.
func TestEtcActionMessageNamesTheAccountOnly(t *testing.T) {
	var cfg *application.Configurator
	app := nagotest.New(t, func(c *application.Configurator) {
		cfg = c
		c.SetApplicationID("de.worldiety.useretcmessagetest")
		if err := c.StandardSystems(); err != nil {
			t.Fatal(err)
		}

		// the scaffold shows the banner messages, like in an application
		c.SetDecorator(c.NewScaffold().Decorator())
	})

	users := std.Must(cfg.UserManagement()).UseCases
	const phone = "+49 441 1234567"

	// an administrator with exactly the permissions of the dialog and its actions, logged in like in a browser
	admin, err := users.Create(user.SU(), user.ShortRegistrationUser{Firstname: "Ada", Lastname: "Admin", Email: "ada.admin@example.com", Password: "Sup3r-Geheim!2026", PasswordRepeated: "Sup3r-Geheim!2026", Verified: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := users.UpdateOtherPermissions(user.SU(), admin.ID, []permission.ID{user.PermFindAll, user.PermFindByID, user.PermUpdateOtherContact}); err != nil {
		t.Fatal(err)
	}

	subject := std.Must(users.SubjectFromUser(user.SU(), admin.ID)).Unwrap()

	tests := []struct {
		name   string
		mail   user.Email
		before func(t *testing.T, id user.ID)
		button string
		prefix string // the message is prefix + "Erika Mustermann (<mail>)" + suffix
		suffix string
	}{
		{
			name:   "deactivate",
			mail:   "erika.deactivate@example.com",
			button: "Nutzer deaktivieren",
			prefix: "Der Nutzer ",
			suffix: " wurde deaktiviert.",
		},
		{
			name: "activate",
			mail: "erika.activate@example.com",
			before: func(t *testing.T, id user.ID) {
				if err := users.UpdateAccountStatus(user.SU(), id, user.Disabled{}); err != nil {
					t.Fatal(err)
				}
			},
			button: "Nutzer aktivieren",
			prefix: "Der Nutzer ",
			suffix: " wurde aktiviert.",
		},
		{
			name:   "notify",
			mail:   "erika.notify@example.com",
			button: "Nutzer benachrichtigen",
			prefix: "Ereignis für ",
			suffix: " erstellt.",
		},
		{
			name:   "confirm mail",
			mail:   "erika.confirm@example.com",
			button: "E-Mail bestätigen",
			prefix: "Der Nutzer ",
			suffix: " wurde bestätigt.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usr, err := users.Create(user.SU(), user.ShortRegistrationUser{Firstname: "Erika", Lastname: "Mustermann", Email: tt.mail})
			if err != nil {
				t.Fatal(err)
			}

			contact := usr.Contact
			contact.Phone = phone
			if err := users.UpdateOtherContact(user.SU(), usr.ID, contact); err != nil {
				t.Fatal(err)
			}

			code, err := users.ResetVerificationCode(usr.ID, user.DefaultVerificationLifeTime)
			if err != nil {
				t.Fatal(err)
			}

			if tt.before != nil {
				tt.before(t, usr.ID)
			}

			// a small window shows the accounts as cards, whose click opens the account like a row of the table
			w := app.Open(t, subject, "admin/accounts", nagotest.Size(400, 900))
			w.Click(w.Find(nagotest.Text(string(tt.mail))))
			w.Click(w.Find(nagotest.Label("Sonstiges")))

			// premise: the code is pending in the account, which the dialog shows
			pending := std.Must(users.FindByID(user.SU(), usr.ID)).Unwrap().VerificationCode.Value
			if pending == "" || pending != code {
				t.Fatalf("premise: verification code %q is not pending, got %q", code, pending)
			}

			w.Click(w.Find(nagotest.Label(tt.button)))

			want := tt.prefix + "Erika Mustermann (" + string(tt.mail) + ")" + tt.suffix
			msg := w.Find(nagotest.Where("message", func(n nagotest.Node) bool {
				tv, ok := n.Component.(*proto.TextView)
				return ok && strings.HasPrefix(string(tv.Value), tt.prefix) && strings.HasSuffix(string(tv.Value), tt.suffix)
			})).Text()
			if msg != want {
				t.Errorf("message:\n got %q\nwant %q", msg, want)
			}

			for _, secret := range []string{code, string(usr.ID), phone} {
				if strings.Contains(msg, secret) {
					t.Errorf("message %q shows %q", msg, secret)
				}
			}
		})
	}
}

// Formatting the model, e.g. by the String method of a state holding it, names the account and shows none of its
// other fields, whatever the verb.
func TestUserModelFormatsAsNameAndMail(t *testing.T) {
	model := uiuser.UserModel{
		ID:               "e282bca23adc53441682fabe6652b62d",
		Email:            "erika@example.com",
		Contact:          user.Contact{Firstname: "Erika", Lastname: "Mustermann", Phone: "+49 441 1234567"},
		VerificationCode: user.Code{Value: "b1d3b028e297a9719c0cff2cea5cbdda"},
	}

	for _, verb := range []string{"%v", "%s", "%+v"} {
		if got, want := fmt.Sprintf(verb, model), "Erika Mustermann (erika@example.com)"; got != want {
			t.Errorf("%s: got %q, want %q", verb, got, want)
		}
	}
}
