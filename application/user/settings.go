// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"log/slog"
	"regexp"
)

var _ = enum.Variant[settings.GlobalSettings, Settings](
	enum.Rename[Settings]("nago.user.settings"),
)

type FieldConstraint string

func (f FieldConstraint) Hidden() bool {
	return f == ""
}

func (f FieldConstraint) Optional() bool {
	return f == "^.*$"
}

func (f FieldConstraint) Required() bool {
	return f == "^.+$"
}

func (f FieldConstraint) Match(str string) bool {
	if f == "" || f.Optional() {
		return true
	}

	if f.Required() && str == "" {
		return false
	}

	regex, err := regexp.Compile(string(f))
	if err != nil {
		slog.Error("failed to compile regex for field constraint", "regex", f, "err", err)
		return false
	}

	return regex.MatchString(str)
}

type ConsentOption struct {
	ID       consent.ID  `json:"id,omitempty"`
	Register ConsentText `json:"register,omitempty"`
	Profile  ConsentText `json:"profile,omitempty"`
	Required bool        `json:"required,omitempty"`
}

func (c ConsentOption) Label() string {
	if c.Profile.Label != "" {
		return c.Profile.Label
	}

	if c.Register.Label != "" {
		return c.Register.Label
	}

	return ""
}

func (c ConsentOption) WithIdentity(id consent.ID) ConsentOption {
	c.ID = id
	return c
}

func (c ConsentOption) Identity() consent.ID {
	return c.ID
}

func (c ConsentOption) String() string {
	return c.Label()
}

type ConsentText struct {
	Label          string `json:"name,omitempty"`
	SupportingText string `json:"text,omitempty"`
}

type Settings struct {
	_ any `title:"Nutzerverwaltung" description:"Allgemeine Vorgaben bezüglich der Nutzerverwaltung vornehmen."`

	SelfPasswordReset bool `json:"selfPasswordReset" label:"Passwort vergessen Funktion" supportingText:"Nutzer können im Self-Service ihre Kennwörter zurücksetzen. Hierfür ist ein Mail-Server erforderlich."`

	SelfRegistration bool `json:"selfRegistration" label:"Freie Registrierung" supportingText:"Wenn erlaubt, dann kann sich jeder anonyme Besucher ein eigenes Konto erstellen. Ansonsten müssen die Nutzerkonten manuell durch einen Administrator erstellt werden."`
	_____            any  `label:"---"`
	__               any  `label:"Die folgenden Einschränkungen gelten für die freie Registrierung und Nutzerkreisverwaltung."`

	AllowedDomains []string   `section:"Rechte" json:"allowedDomains" lines:"5" label:"Erlaubte Domains" supportingText:"Jede Zeile stellt einen erlaubten Domänen Suffix dar, also z.B. @worldiety.de. Wenn diese Liste leer ist, darf sich jeder registrieren."`
	DefaultRoles   []role.ID  `section:"Rechte" json:"defaultRoles" source:"nago.roles" label:"Standardrolle" supportingText:"Diese Rollen werden pauschal einem neuen Nutzer hinzugefügt."`
	DefaultGroups  []group.ID `section:"Rechte" json:"defaultGroups" source:"nago.groups" label:"Standardgruppen" supportingText:"Diese Gruppen werden pauschal einem neuen Nutzer hinzugefügt."`

	___ any `section:"Rechtliches" label:"Die folgenden Einschränkungen gelten für die freie Registrierung."`
	// deprecated: use Consents
	RequireTermsAndConditions bool `section:"Rechtliches" visible:"false" json:"requireTermsAndConditions" label:"AGB Zustimmung erforderlich" supportingText:"Wenn erforderlich, muss der Nutzer bei der Registrierung der AGB explizit zustimmen."`
	// deprecated: use Consents
	RequireTermsOfUse bool `section:"Rechtliches" visible:"false" json:"requireTermsOfUse" label:"Zustimmung zu den Nutzungsbedingungen erforderlich" supportingText:"Wenn erforderlich, muss der Nutzer bei der Registrierung den Nutzungsbedingungen explizit zustimmen."`
	// deprecated: use Consents
	RequireDataProtectionConditions bool `section:"Rechtliches" visible:"false" json:"requireDataProtectionConditions" label:"Datenschutz Zustimmung erforderlich" supportingText:"Wenn erforderlich, muss der Nutzer bei der Registrierung den Datenschutzbestimmungen explizit zustimmen."`
	// deprecated: use Consents
	CanAcceptNewsletter bool `section:"Rechtliches" visible:"false" json:"canAcceptNewsletter" label:"Newsletter anbieten" supportingText:"Wenn eingeschaltet, wird die Möglichkeit angeboten, dass der Nutzer dem Empfang von Newslettern zustimmen kann."`
	// deprecated: use Consents
	CanReceiveSMS bool `section:"Rechtliches" visible:"false" json:"canAcceptSMS" label:"SMS-Versand anbieten" supportingText:"Wenn eingeschaltet, wird die Möglichkeit angeboten, dass der Nutzer dem Empfang von SMS zustimmen kann."`
	// deprecated: use Consents
	RequireMinAge int             `section:"Rechtliches" visible:"false" json:"requireMinAge" label:"Mindestalter bestätigen" supportingText:"Je nach Angebot und Markt, gibt es ein Mindestalter, um als Nutzer geschäftsfähig zu sein. Vollgeschäftsfähig gilt man in Deutschland grundsätzlich ab 18 Jahre. Es ist jedoch üblich, gemäß Taschengeldparagraphen auch Minderjährige und damit beschränkt geschäftsfähige Personen zu erlauben."`
	Consents      []ConsentOption `section:"Rechtliches" json:"adoptionOptions"`

	____              any             `section:"Kontakt" label:"Die folgenden Kontaktinformationen müssen bei der freien Registrierung abgefragt werden. Ein leeres Feld bedeutet, dass das bezeichnete Feld ausgeblendet wird. Ansonsten drückt ein regulärer Ausdruck die Validierung aus. ^.*$ steht für optional und ^.+$ für erforderlich. Um einen Wert aus einer festen Menge zu verwenden, kannst du einen Ausdruck wie ^(OptionA|OptionB)$ verwenden."`
	Salutation        FieldConstraint `section:"Kontakt" label:"Anrede"`
	Title             FieldConstraint `section:"Kontakt" label:"Titel"`
	Position          FieldConstraint `section:"Kontakt" label:"Position"`
	CompanyName       FieldConstraint `section:"Kontakt" label:"Unternehmen"`
	City              FieldConstraint `section:"Kontakt" label:"Stadt"`
	PostalCode        FieldConstraint `section:"Kontakt" label:"Postleitzahl"`
	State             FieldConstraint `section:"Kontakt" label:"Bundesland"`
	Country           FieldConstraint `section:"Kontakt" label:"Land"`
	ProfessionalGroup FieldConstraint `section:"Kontakt" label:"Berufsgruppe"`
	MobilePhone       FieldConstraint `section:"Kontakt" label:"Mobile"`

	______     any        `section:"Anonyme Nutzer" label:"Standardrollen und Gruppen von anonymen Nutzern. Diese Rollen werden nicht auf gültige, ungültige oder angemeldete Nutzer vererbt, sodass eine entsprechende Unterscheidung möglich ist."`
	AnonRoles  []role.ID  `section:"Anonyme Nutzer" json:"anonRoles" source:"nago.roles" label:"Standardrolle" supportingText:"Diese Rollen hat jeder anonyme Nutzer."`
	AnonGroups []group.ID `section:"Anonyme Nutzer" json:"anonGroups" source:"nago.groups" label:"Standardgruppen" supportingText:"Diese Gruppen hat jeder anonyme Nutzer."`
}

func (s Settings) GlobalSettings() bool { return true }
