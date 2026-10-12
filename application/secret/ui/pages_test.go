// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret_test

import (
	"encoding/json"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
)

const smtpType = "go.wdy.de/nago/application/secret.SMTP"

func newApp(t *testing.T) (*nagotest.App, application.SecretManagement) {
	t.Helper()
	var secrets application.SecretManagement
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.secretui")
		secrets = std.Must(cfg.SecretManagement())
	})

	return app, secrets
}

func count(t *testing.T, secrets application.SecretManagement) int {
	t.Helper()
	n := 0
	for _, err := range secrets.UseCases.FindMySecrets(user.SU()) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	return n
}

// Choosing a type and leaving the form again must not leave an empty secret behind.
func TestCreateStoresNothingUntilSaved(t *testing.T) {
	app, secrets := newApp(t)
	wnd := app.Open(t, user.SU(), secrets.Pages.CreateSecret, nagotest.Values(core.Values{"type": smtpType}))
	if count(t, secrets) != 0 {
		t.Fatal("opening the form must not store a secret")
	}

	wnd.Click(wnd.Find(nagotest.Text("Abbrechen")))
	if count(t, secrets) != 0 {
		t.Fatal("cancelling must not store a secret")
	}

	wnd = app.Open(t, user.SU(), secrets.Pages.CreateSecret, nagotest.Values(core.Values{"type": smtpType}))
	wnd.Type(wnd.Find(nagotest.Label("Host")), "smtp.example.com")
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))
	if count(t, secrets) != 1 {
		t.Fatal("saving must store the secret")
	}
}

// The stored password is never sent to the browser, and saving the form without a new password keeps it.
func TestEditDoesNotRevealPasswords(t *testing.T) {
	app, secrets := newApp(t)
	id := std.Must(secrets.UseCases.CreateSecret(user.SU(), secret.SMTP{Name: "Mail", Host: "smtp.example.com", Password: "very-secret-pw"}))

	wnd := app.Open(t, user.SU(), secrets.Pages.EditSecret, nagotest.Values(core.Values{"id": string(id)}))
	tree := string(std.Must(json.Marshal(wnd.Tree())))
	if strings.Contains(tree, "very-secret-pw") {
		t.Fatal("the stored password must not be sent to the browser")
	}

	if !strings.Contains(tree, "smtp.example.com") {
		t.Fatal("expected the other fields to be shown")
	}

	if !strings.Contains(tree, "Bleibt das Feld leer, wird er nicht geändert.") {
		t.Fatal("expected a hint below the password field, that an empty field keeps the stored value")
	}

	wnd.Type(wnd.Find(nagotest.Label("Host")), "mail.example.com")
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))

	stored := std.Must(secrets.UseCases.FindMySecretByID(user.SU(), id)).Unwrap().Credentials.(secret.SMTP)
	if stored.Host != "mail.example.com" || stored.Password != "very-secret-pw" {
		t.Fatalf("expected the new host and the stored password, got %+v", stored)
	}
}
