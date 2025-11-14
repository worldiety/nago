// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package spryng

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

var (
	StrSmsSpryngSettingsTitle       = i18n.MustString("nago.sms.spryng.settings_title", i18n.Values{language.English: "My Spryng Token", language.German: "Mein Spryng Token"})
	StrSmsSpryngSettingsName        = i18n.MustString("nago.sms.spryng.settings_name", i18n.Values{language.English: "Spryng SMS", language.German: "Spryng SMS"})
	StrSmsSpryngSettingsDescription = i18n.MustString("nago.sms.spryng.settings_desc", i18n.Values{language.English: "API Token to connect to Spryng SMS", language.German: "Token zur API Anbindung von Spryng SMS"})
)

type Settings struct {
	Name  string `value:"nago.sms.spryng.settings_title" json:"name"`
	Token string `json:"token"`
	RPS   int    `json:"rps"` // 20 by default
	_     string `credentialName:"nago.sms.spryng.settings_name" credentialDescription:"nago.sms.spryng.settings_desc" credentialLogo:"https://www.spryng.de/favicon.png"`
}

var _ = enum.Variant[secret.Credentials, Settings](enum.Rename[Settings]("nago.sms.spryng.settings"))

func (s Settings) GetName() string {
	return s.Name
}

func (s Settings) Credentials() bool {
	return true
}

func (s Settings) IsZero() bool {
	return Settings{} == s
}
