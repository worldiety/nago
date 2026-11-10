// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/secret"
	_ "go.wdy.de/nago/application/secret/gostrictnet"
	uisecret "go.wdy.de/nago/application/secret/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

const (
	keyValuesType = "go.wdy.de/nago/application/secret.KeyValues"
	storedToken   = "very-secret-token"
	testLogo      = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1" id="test-logo"></svg>`
)

func storedKeyValues() secret.KeyValues {
	return secret.KeyValues{Name: "Billing", Sections: []secret.KeyValueSection{{Label: "Zugang", SupportingText: "Vom Anbieter", Fields: []secret.KeyValueField{
		{Key: secret.KeyBaseURL, Label: "Adresse", Kind: secret.FieldURL, Value: "https://billing.example.com"},
		{Key: secret.KeyHeaderBearerAuth, Label: "Token", Kind: secret.FieldString, Protection: secret.WriteOnly, Value: storedToken},
		{Key: "tenant", Label: "Mandant", Kind: secret.FieldString, Value: "acme"},
		{Key: "pin", Label: "PIN", Kind: secret.FieldString, Protection: secret.Concealed, Value: "1234"},
		{Key: "note", Label: "Notiz", Kind: secret.FieldMultiLine, Value: "a\nb"},
		{Key: "retries", Label: "Versuche", Kind: secret.FieldInt, Value: "3"},
		{Key: "debug", Label: "Debug", Kind: secret.FieldBool, Value: "false"},
	}}}}
}

func openKeyValues(t *testing.T) (*nagotest.App, application.SecretManagement, *nagotest.Window, secret.ID) {
	t.Helper()
	app, secrets := newApp(t)
	id := std.Must(secrets.UseCases.CreateSecret(user.SU(), storedKeyValues()))
	wnd := app.Open(t, user.SU(), secrets.Pages.EditSecret, nagotest.Values(core.Values{"id": string(id)}))
	return app, secrets, wnd, id
}

func load(secrets application.SecretManagement, id secret.ID) secret.KeyValues {
	return std.Must(secrets.UseCases.FindMySecretByID(user.SU(), id)).Unwrap().Credentials.(secret.KeyValues)
}

func fieldByKey(t *testing.T, kv secret.KeyValues, key string) secret.KeyValueField {
	t.Helper()
	for field := range kv.Fields() {
		if field.Key == key {
			return field
		}
	}

	t.Fatalf("no field %q in %+v", key, kv)
	return secret.KeyValueField{}
}

func tree(wnd *nagotest.Window) string {
	return string(std.Must(json.Marshal(wnd.Tree())))
}

func saveEnabled(wnd *nagotest.Window) bool {
	n := wnd.Find(nagotest.Text("Speichern")).Node()
	for _, c := range append(slices.Clone(n.Parents), n.Component) {
		disabled := reflect.Indirect(reflect.ValueOf(c)).FieldByName("Disabled")
		if disabled.IsValid() && disabled.Bool() {
			return false
		}
	}

	return true
}

func imagesWithSVG(wnd *nagotest.Window) nagotest.Selection {
	return wnd.FindAll(nagotest.Where("svg logo", func(n nagotest.Node) bool {
		img, ok := n.Component.(*proto.Img)
		return ok && strings.Contains(string(img.SVG), "test-logo")
	}))
}

// menuEntry selects the entry of the menu, whose anchor has the label. All menus are part of the tree, even if closed.
func menuEntry(wnd *nagotest.Window, anchor, entry string) nagotest.Selection {
	return menuOf(wnd, anchor).Find(nagotest.Text(entry))
}

func menuOf(wnd *nagotest.Window, anchor string) nagotest.Selection {
	return wnd.Find(nagotest.And(nagotest.Type[*proto.Menu](), nagotest.Containing(nagotest.Label(anchor))))
}

// addField selects a preset from the menu of the first section.
func addField(wnd *nagotest.Window, preset string) {
	wnd.Click(wnd.FindAll(nagotest.Text(preset)).At(0))
}

func TestKeyValuesEditorShowsFieldsAndHidesWriteOnly(t *testing.T) {
	_, _, wnd, _ := openKeyValues(t)

	content := tree(wnd)
	if strings.Contains(content, storedToken) {
		t.Fatal("the stored token must not be sent to the browser")
	}

	for _, want := range []string{
		"https://billing.example.com",
		"Verbindung · Adresse",
		"Verbindung · Bearer-Token",
		"Programme lesen ·",
		"Zugang",
		"Vom Anbieter",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q", want)
		}
	}

	// the stored value is a state of the row and no empty input
	wnd.Find(nagotest.Text("Gespeichert")).Exactly(1)
	wnd.Find(nagotest.Text("Ersetzen")).Exactly(1)
	wnd.FindAll(nagotest.And(nagotest.Type[*proto.PasswordField](), nagotest.Label("Token"))).None()

	wnd.Find(nagotest.And(nagotest.Type[*proto.PasswordField](), nagotest.Label("PIN"))).Exactly(1)
	wnd.Find(nagotest.And(nagotest.Type[*proto.TextField](), nagotest.Label("Adresse"))).Exactly(1)
	if lines := wnd.Find(nagotest.Label("Notiz")).Node().Component.(*proto.TextField).Lines; lines == 0 {
		t.Error("expected a multi-line input")
	}
	wnd.Find(nagotest.Type[*proto.Toggle]()).Exactly(1)
	wnd.FindAll(nagotest.Label("Schlüssel")).None()
}

// A write-only value can only be replaced, and cancelling keeps what is stored.
func TestKeyValuesEditorReplacesWriteOnly(t *testing.T) {
	_, secrets, wnd, id := openKeyValues(t)

	wnd.Click(wnd.Find(nagotest.Text("Ersetzen")))
	wnd.Type(wnd.Find(nagotest.And(nagotest.Type[*proto.PasswordField](), nagotest.Label("Token"))), "typed")
	if !saveEnabled(wnd) {
		t.Fatal("expected the new value to be detected")
	}

	wnd.Click(wnd.FindAll(nagotest.Text("Abbrechen")).At(0)) // the page has a button of the same name
	wnd.Find(nagotest.Text("Gespeichert")).Exactly(1)
	if saveEnabled(wnd) {
		t.Fatal("expected to be clean after cancelling")
	}

	wnd.Click(wnd.Find(nagotest.Text("Ersetzen")))
	wnd.Type(wnd.Find(nagotest.And(nagotest.Type[*proto.PasswordField](), nagotest.Label("Token"))), "typed")
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))
	if value := fieldByKey(t, load(secrets, id), secret.KeyHeaderBearerAuth).Value; value != "typed" {
		t.Fatalf("expected the new token, got %q", value)
	}
}

// A password field is single-line and would drop the line breaks of a pasted key.
func TestKeyValuesEditorProtectedMultiLineKeepsLineBreaks(t *testing.T) {
	const pem = "-----BEGIN KEY-----\nabc\n-----END KEY-----"

	app, secrets := newApp(t)
	id := std.Must(secrets.UseCases.CreateSecret(user.SU(), secret.KeyValues{Name: "Key", Sections: []secret.KeyValueSection{{Fields: []secret.KeyValueField{
		{Key: "pem", Label: "Schlüsseldatei", Kind: secret.FieldMultiLine, Protection: secret.Concealed, Value: pem},
	}}}}))
	wnd := app.Open(t, user.SU(), secrets.Pages.EditSecret, nagotest.Values(core.Values{"id": string(id)}))

	if strings.Contains(tree(wnd), "BEGIN KEY") {
		t.Fatal("expected the value to stay on the server, until it is edited")
	}

	wnd.Click(wnd.Find(nagotest.Text("Aufdecken")))
	input := wnd.Find(nagotest.And(nagotest.Type[*proto.TextField](), nagotest.Label("Schlüsseldatei")))
	if input.Node().Component.(*proto.TextField).Lines == 0 {
		t.Fatal("expected a multi-line input")
	}

	wnd.Type(input, pem+"\nmore")
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))
	if value := fieldByKey(t, load(secrets, id), "pem").Value; value != pem+"\nmore" {
		t.Fatalf("expected the line breaks to be kept, got %q", value)
	}
}

func TestKeyValuesEditorNumbersUseTheIntegerKeyboard(t *testing.T) {
	_, _, wnd, _ := openKeyValues(t)

	input := wnd.Find(nagotest.Label("Versuche")).Node().Component.(*proto.TextField)
	if input.KeyboardOptions.KeyboardType != proto.KeyboardInteger || !bool(input.ShowZero) {
		t.Fatalf("expected an integer input, which shows a stored 0, got %+v", input.KeyboardOptions)
	}
}

func TestKeyValuesEditorMoves(t *testing.T) {
	_, secrets, wnd, id := openKeyValues(t)

	menuOf(wnd, "Aktionen: Adresse").FindAll(nagotest.Text("Nach oben")).None()
	menuOf(wnd, "Aktionen: Debug").FindAll(nagotest.Text("Nach unten")).None()

	wnd.Click(menuEntry(wnd, "Aktionen: Mandant", "Nach unten"))
	wnd.Click(menuEntry(wnd, "Aktionen: Versuche", "Nach oben"))
	if !saveEnabled(wnd) {
		t.Fatal("expected the order to be detected")
	}

	wnd.Click(wnd.Find(nagotest.Text("Speichern")))

	var keys []string
	for field := range load(secrets, id).Fields() {
		keys = append(keys, field.Key)
	}

	if want := []string{secret.KeyBaseURL, secret.KeyHeaderBearerAuth, "pin", "tenant", "retries", "note", "debug"}; !slices.Equal(keys, want) {
		t.Fatalf("unexpected order %v", keys)
	}
}

func TestKeyValuesEditorKeepsTokenWithoutNewInput(t *testing.T) {
	_, secrets, wnd, id := openKeyValues(t)
	before := load(secrets, id)

	wnd.Type(wnd.Find(nagotest.Label("Mandant")), "other")
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))

	got := load(secrets, id)
	if fieldByKey(t, got, "tenant").Value != "other" || fieldByKey(t, got, secret.KeyHeaderBearerAuth).Value != storedToken {
		t.Fatalf("expected the new value and the stored token, got %+v", got)
	}

	if fieldByKey(t, got, "tenant").ID != fieldByKey(t, before, "tenant").ID {
		t.Fatal("expected a stable identifier")
	}
}

func TestKeyValuesEditorDropsTokenWhenAddressChanges(t *testing.T) {
	_, secrets, wnd, id := openKeyValues(t)

	wnd.Type(wnd.Find(nagotest.Label("Adresse")), "https://evil.example.com")
	wnd.Find(nagotest.Text("Neu eingeben")).Exactly(1)
	wnd.FindAll(nagotest.Text("Ersetzen")).None()
	if !strings.Contains(tree(wnd), "Neue Adresse: gespeicherte Werte neu eingeben.") {
		t.Fatal("expected a hint at the address")
	}

	wnd.Click(wnd.Find(nagotest.Text("Speichern")))

	got := load(secrets, id)
	if fieldByKey(t, got, secret.KeyBaseURL).Value != "https://evil.example.com" || fieldByKey(t, got, secret.KeyHeaderBearerAuth).Value != "" {
		t.Fatalf("expected the new address without the token, got %+v", got)
	}
}

func TestKeyValuesEditorDirtyCheck(t *testing.T) {
	_, _, wnd, _ := openKeyValues(t)
	if saveEnabled(wnd) {
		t.Fatal("nothing has changed")
	}

	wnd.Type(wnd.Find(nagotest.Label("Mandant")), "other")
	if !saveEnabled(wnd) {
		t.Fatal("expected the change to be detected")
	}

	wnd.Type(wnd.Find(nagotest.Label("Mandant")), "acme")
	if saveEnabled(wnd) {
		t.Fatal("expected the original value to be clean again")
	}

	// adding and removing again leaves no trace, also not as an empty slice
	wnd.Click(wnd.Find(nagotest.Text("Abschnitt hinzufügen")))
	if !saveEnabled(wnd) {
		t.Fatal("expected the new section to be detected")
	}

	wnd.Click(wnd.Find(nagotest.Text("Schließen")))
	wnd.Click(menuEntry(wnd, "Aktionen: Abschnitt", "Entfernen"))
	if saveEnabled(wnd) {
		t.Fatal("expected to be clean after removing the section")
	}

	addField(wnd, "Text")
	wnd.Click(wnd.Find(nagotest.Text("Schließen")))
	wnd.Click(menuEntry(wnd, "Aktionen: Neues Feld", "Entfernen"))
	if saveEnabled(wnd) {
		t.Fatal("expected to be clean after removing the field")
	}
}

// A new secret needs no section before its first field, and the dialog of a new field is open for its key.
func TestKeyValuesEditorCreates(t *testing.T) {
	app, secrets := newApp(t)
	wnd := app.Open(t, user.SU(), secrets.Pages.CreateSecret, nagotest.Values(core.Values{"type": keyValuesType}))

	wnd.Type(wnd.Find(nagotest.Label("Name")), "Neu")
	wnd.Find(nagotest.Text("Noch keine Felder")).Exactly(1)
	addField(wnd, "Text")
	wnd.Type(wnd.Find(nagotest.Label("Schlüssel")), secret.KeyBaseURL)
	if kind := string(wnd.Find(nagotest.Label("Art")).Node().Component.(*proto.Select).Value); kind != string(secret.FieldURL) {
		t.Fatalf("expected the address to become a url, got %q", kind)
	}

	wnd.Click(wnd.Find(nagotest.Text("Schließen")))
	wnd.Type(wnd.Find(nagotest.Label(secret.KeyBaseURL)), "https://new.example.com")
	addField(wnd, "Text")
	wnd.Type(wnd.Find(nagotest.Label("Schlüssel")), "token")
	wnd.Click(wnd.Find(nagotest.Text("Schließen")))
	wnd.Click(menuEntry(wnd, "Schutz von token: Sichtbar", "Nur schreibbar"))
	wnd.Type(wnd.Find(nagotest.And(nagotest.Type[*proto.PasswordField](), nagotest.Label("token"))), "typed-token")
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))

	if count(t, secrets) != 1 {
		t.Fatal("expected the secret to be stored")
	}

	var kv secret.KeyValues
	for scr, err := range secrets.UseCases.FindMySecrets(user.SU()) {
		if err != nil {
			t.Fatal(err)
		}

		kv = scr.Credentials.(secret.KeyValues)
	}

	if kv.Name != "Neu" || kv.Sections[0].ID == "" {
		t.Fatalf("unexpected secret %+v", kv)
	}

	address := fieldByKey(t, kv, secret.KeyBaseURL)
	if address.Kind != secret.FieldURL || address.Value != "https://new.example.com" || address.ID == "" {
		t.Fatalf("unexpected address %+v", address)
	}

	if token := fieldByKey(t, kv, "token"); token.Protection != secret.WriteOnly || token.Value != "typed-token" {
		t.Fatalf("unexpected token %+v", token)
	}
}

// The server decides, so an invalid field stays on the page and nothing is stored.
func TestKeyValuesEditorShowsValidationErrors(t *testing.T) {
	_, secrets, wnd, id := openKeyValues(t)

	wnd.Click(menuEntry(wnd, "Aktionen: Mandant", "Bearbeiten"))
	wnd.Type(wnd.Find(nagotest.Label("Schlüssel")), "HEADER_HOST")
	wnd.Click(wnd.Find(nagotest.Text("Schließen")))
	wnd.Click(wnd.Find(nagotest.Text("Speichern")))

	if fieldByKey(t, load(secrets, id), "tenant").Value != "acme" {
		t.Fatal("expected nothing to be stored")
	}

	wnd.Find(nagotest.Text("Speichern")).Exactly(1)
}

func TestKeyValuesEditorWriteOnlyOnlyForText(t *testing.T) {
	_, _, wnd, _ := openKeyValues(t)

	menuOf(wnd, "Schutz von Versuche: Sichtbar").FindAll(nagotest.Text("Nur schreibbar")).None()
	menuEntry(wnd, "Schutz von Mandant: Sichtbar", "Nur schreibbar").Exactly(1)

	// a kind without write-only demotes the protection, and a value which does not fit is gone
	wnd.Click(menuEntry(wnd, "Schutz von Mandant: Sichtbar", "Nur schreibbar"))
	wnd.Click(menuEntry(wnd, "Aktionen: Mandant", "Bearbeiten"))
	wnd.Type(wnd.Find(nagotest.Label("Art")), string(secret.FieldInt))
	wnd.Click(wnd.Find(nagotest.Text("Schließen")))
	wnd.Find(nagotest.Label("Schutz von Mandant: Verdeckt")).Exactly(1)

	if value := string(wnd.Find(nagotest.Label("Mandant")).Node().Component.(*proto.PasswordField).Value); value != "" {
		t.Fatalf("expected the value to be gone, got %q", value)
	}
}

// Leaving write-only without a new value loses the stored one, which the row has to tell before saving.
func TestKeyValuesEditorWarnsAboutLostValue(t *testing.T) {
	_, _, wnd, _ := openKeyValues(t)

	wnd.FindAll(nagotest.Text("Gespeicherter Wert geht verloren")).None()
	wnd.Click(menuEntry(wnd, "Schutz von Token: Nur schreibbar", "Verdeckt"))
	wnd.Find(nagotest.Text("Gespeicherter Wert geht verloren")).Exactly(1)
}

// Only what can still be added is offered, because a key is unique and a login sets a single header.
func TestKeyValuesEditorOffersPresets(t *testing.T) {
	app, secrets := newApp(t)
	wnd := app.Open(t, user.SU(), secrets.Pages.CreateSecret, nagotest.Values(core.Values{"type": keyValuesType}))

	addField(wnd, "Basic-Auth")
	for _, key := range []string{"Benutzername", "Passwort"} {
		wnd.FindAll(nagotest.Label(key)).Exactly(1)
	}

	wnd.FindAll(nagotest.Text("Basic-Auth")).None()
	wnd.FindAll(nagotest.Text("Bearer-Token")).None()
	wnd.FindAll(nagotest.Text("Adresse")).Exactly(1)

	addField(wnd, "Adresse")
	wnd.FindAll(nagotest.Text("Adresse")).None()
	if strings.Contains(tree(wnd), "Feld bearbeiten") {
		t.Fatal("a preset with its key needs no dialog")
	}

	addField(wnd, "Query-Parameter")
	if key := string(wnd.Find(nagotest.Label("Schlüssel")).Node().Component.(*proto.TextField).Value); key != secret.KeyPrefixQuery {
		t.Fatalf("expected the key to be prepared, got %q", key)
	}
}

func TestKeyValuesEditorLogo(t *testing.T) {
	_, secrets, wnd, id := openKeyValues(t)

	wnd.Click(wnd.Find(nagotest.Text("Logo hochladen")))
	imports := wnd.Imports()
	if len(imports) != 1 || int64(imports[0].MaxBytes) <= secret.MaxLogoSVGBytes {
		t.Fatalf("unexpected imports %+v", imports)
	}

	wnd.Upload(string(imports[0].ID), nagotest.File("logo.svg", []byte("<html></html>")))
	imagesWithSVG(wnd).None()

	wnd.Click(wnd.Find(nagotest.Text("Logo hochladen")))
	wnd.Upload(string(wnd.Imports()[1].ID), nagotest.File("logo.svg", []byte(`<?xml version="1.0"?>`+testLogo)))
	imagesWithSVG(wnd).Exactly(1)
	if !saveEnabled(wnd) {
		t.Fatal("expected the logo to be detected")
	}

	wnd.Click(wnd.Find(nagotest.Text("Speichern")))
	if got := load(secrets, id).LogoSVG; got != testLogo {
		t.Fatalf("unexpected logo %q", got)
	}
}

func TestKeyValuesLogoReplacesTypeLogo(t *testing.T) {
	app, secrets := newApp(t)
	kv := storedKeyValues()
	kv.LogoSVG = testLogo
	id := std.Must(secrets.UseCases.CreateSecret(user.SU(), kv))

	vault := app.Open(t, user.SU(), secrets.Pages.Vault)
	imagesWithSVG(vault).Exactly(1)

	edit := app.Open(t, user.SU(), secrets.Pages.EditSecret, nagotest.Values(core.Values{"id": string(id)}))
	imagesWithSVG(edit).Exactly(1) // the editor brings it along, there is no card of the type

	edit.Click(edit.Find(nagotest.Text("Logo entfernen")))
	imagesWithSVG(edit).None()
	edit.Click(edit.Find(nagotest.Text("Speichern")))
	if got := load(secrets, id).LogoSVG; got != "" {
		t.Fatalf("expected the logo to be removed, got %q", got)
	}
}

func TestKeyValuesEditorIsReusable(t *testing.T) {
	var state *core.State[secret.KeyValues]
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.secretui.editor")
		cfg.RootView("editor", func(wnd core.Window) core.View {
			// without any identifiers, because an application may build the model by itself
			state = core.AutoState[secret.KeyValues](wnd).Init(func() secret.KeyValues {
				return secret.KeyValues{Name: "App", Sections: []secret.KeyValueSection{{Fields: []secret.KeyValueField{
					{Key: "token", Label: "Token", Kind: secret.FieldString, Protection: secret.WriteOnly},
					{Key: "tenant", Label: "Mandant", Kind: secret.FieldString, Value: "acme"},
				}}}}
			})

			return uisecret.KeyValuesEditor(state).Disabled(wnd.Values()["disabled"] == "true")
		})
	})

	wnd := app.Open(t, user.SU(), "editor")
	wnd.Find(nagotest.And(nagotest.Type[*proto.PasswordField](), nagotest.Label("Token"))).Exactly(1)
	wnd.FindAll(nagotest.Text("Gespeichert")).None()

	wnd.Type(wnd.Find(nagotest.Label("Mandant")), "other")
	if value := fieldByKey(t, state.Get(), "tenant").Value; value != "other" {
		t.Fatalf("expected the state to be updated, got %q", value)
	}

	disabled := app.Open(t, user.SU(), "editor", nagotest.Values(core.Values{"disabled": "true"}))
	if !bool(disabled.Find(nagotest.Label("Mandant")).Node().Component.(*proto.TextField).Disabled) {
		t.Fatal("expected a disabled input")
	}

	disabled.FindAll(nagotest.Text("Feld hinzufügen")).None()
	disabled.FindAll(nagotest.Text("Abschnitt hinzufügen")).None()
	disabled.FindAll(nagotest.Label("Aktionen: Mandant")).None()
	disabled.FindAll(nagotest.Type[*proto.Menu]()).None()
	disabled.Find(nagotest.Label("Schutz von Mandant: Sichtbar")).Exactly(1)
}

func TestKeyValuesDetails(t *testing.T) {
	kv := storedKeyValues()
	kv.LogoSVG = testLogo
	kv.Sections[0].Fields = append(kv.Sections[0].Fields, secret.KeyValueField{Key: "unset", Label: "Offen", Kind: secret.FieldString, Protection: secret.WriteOnly})

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.secretui.details")
		cfg.RootView("details", func(wnd core.Window) core.View {
			return uisecret.KeyValuesDetails(kv)
		})
	})

	wnd := app.Open(t, user.SU(), "details")
	content := tree(wnd)
	if strings.Contains(content, storedToken) {
		t.Fatal("a write-only value must never be sent")
	}

	if strings.Contains(content, `"1234"`) {
		t.Fatal("a concealed value must not be sent, before it is revealed")
	}

	imagesWithSVG(wnd).Exactly(1)
	wnd.Find(nagotest.Text("Gespeichert")).Exactly(1)
	wnd.Find(nagotest.Text("Fehlt")).Exactly(1)
	wnd.Find(nagotest.Text("acme")).Exactly(1)
	wnd.Find(nagotest.Text("Nein")).Exactly(1)
	wnd.Find(nagotest.Text("••••••••")).Exactly(1)

	wnd.Click(wnd.Find(nagotest.Label("Wert anzeigen: PIN")))
	wnd.Find(nagotest.Text("1234")).Exactly(1)
	wnd.FindAll(nagotest.Text("••••••••")).None()

	wnd.Click(wnd.Find(nagotest.Label("Wert verbergen: PIN")))
	wnd.FindAll(nagotest.Text("1234")).None()

	// copying goes to the clipboard and not into the page, and a write-only value offers nothing to copy
	if strings.Contains(tree(wnd), `"1234"`) {
		t.Fatal("a concealed value must not be sent, before it is revealed")
	}

	wnd.Find(nagotest.Label("Wert kopieren: PIN")).Exactly(1)
	wnd.FindAll(nagotest.Label("Wert kopieren: Token")).None()
	wnd.FindAll(nagotest.Label("Wert anzeigen: Token")).None()
	wnd.Find(nagotest.And(nagotest.Text("https://billing.example.com"), nagotest.Where("link", func(n nagotest.Node) bool {
		return n.Component.(*proto.TextView).Link.Url != ""
	}))).Exactly(1)
}

func TestKeyValuesIsOffered(t *testing.T) {
	app, secrets := newApp(t)

	wnd := app.Open(t, user.SU(), secrets.Pages.CreateSecret)
	if !strings.Contains(tree(wnd), "Schlüssel und Werte") {
		t.Fatal("expected the key values to be offered")
	}
}
