// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

type SMTP struct {
	Name          string `value:"Mein SMTP Server"`
	Host          string
	Port          int `value:"587"`
	Username      string
	Password      string   `style:"secret" credentialWriteOnly:"true"`
	SenderAddress string   `value:"" label:"Absenderadresse" supportingText:"Muss eine gültige E-Mail-Adresse sein. Wenn leer oder ungültig, wird der Username verwendet, sofern dieser eine gültige E-Mail-Adresse ist."`
	// RateLimitPerHour limits the amount of send attempts per hour. 0 means unlimited.
	RateLimitPerHour int `label:"Max. Mails pro Stunde" supportingText:"Optionale Ratenbegrenzung des Anbieters. 0 bedeutet unbegrenzt."`
	// RateLimitPerDay limits the amount of send attempts within 24 hours. 0 means unlimited.
	RateLimitPerDay int `label:"Max. Mails pro Tag" supportingText:"Optionale Ratenbegrenzung des Anbieters, z.B. 500. 0 bedeutet unbegrenzt."`
	_             struct{} `credentialName:"SMTP Postausgangsserver" credentialDescription:"Ein Postausgangsserver wird benötigt, um E-Mails zu verschicken." credentialLogo:"https://www.thunderbird.net/media/img/thunderbird/favicon-196.png"`
}

func (SMTP) Credentials() bool {
	return true
}

func (s SMTP) GetName() string {
	return s.Name
}

func (s SMTP) IsZero() bool {
	return s == SMTP{}
}

// HTTP is a connection of a program to an HTTP API, see package gostrictnet. A program refers to it by its name and
// never sees its address or token. Auth is none, bearer for an Authorization Bearer token, basic for a user name and
// a token as password or header for the token in the header HeaderName.
type HTTP struct {
	Name       string   `value:"Meine API" label:"Name" supportingText:"Unter diesem Namen verwendet ein Programm die Verbindung."`
	BaseURL    string   `label:"Basis-URL" supportingText:"Adresse der API, z.B. https://api.example.com/v1"`
	Auth       string   `label:"Anmeldung" values:"[\"none\", \"bearer\", \"basic\", \"header\"]"`
	Username   string   `label:"Benutzername" supportingText:"Nur für die Anmeldung basic."`
	HeaderName string   `label:"Header" supportingText:"Nur für die Anmeldung header, z.B. X-API-Key."`
	Token      string   `label:"Token oder Passwort" style:"secret" credentialWriteOnly:"true"`
	_          struct{} `credentialName:"HTTP-Verbindung" credentialDescription:"Eine Verbindung, über die Programme eine HTTP-API aufrufen, ohne die Zugangsdaten zu sehen."`
}

func (HTTP) Credentials() bool {
	return true
}

func (h HTTP) GetName() string {
	return h.Name
}

func (h HTTP) IsZero() bool {
	return h == HTTP{}
}
