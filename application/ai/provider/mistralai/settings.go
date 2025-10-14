// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

var _ = enum.Variant[secret.Credentials, Settings](
	enum.Rename[Settings]("nago.ai.mistralai.settings"),
)

var (
	StrMymistralaiSettingsTitle       = i18n.MustString("nago.ai.mistralai.settings_title", i18n.Values{language.English: "My Mistral AI Token", language.German: "Mein Mistral AI Token"})
	StrMymistralaiSettingsName        = i18n.MustString("nago.ai.mistralai.settings_name", i18n.Values{language.English: "Mistral AI", language.German: "mistral AI"})
	StrMymistralaiSettingsDescription = i18n.MustString("nago.ai.mistralai.settings_desc", i18n.Values{language.English: "API Token to connect to Mistral AI", language.German: "Token zur API Anbindung von Mistral AI"})
)

type Settings struct {
	Name  string `value:"nago.ai.mistralai.settings_title"`
	Token string `style:"secret"`
	_     string `credentialName:"nago.ai.mistralai.settings_name" credentialDescription:"nago.ai.mistralai.settings_desc" credentialLogo:"https://mistral.ai/favicon.ico?favicon.95e802d0.ico"`
}

func (Settings) Credentials() bool {
	return true
}

func (s Settings) GetName() string {
	return s.Name
}

func (s Settings) IsZero() bool {
	return s == Settings{}
}
