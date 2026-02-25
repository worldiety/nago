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
	StrMistralAISettingsTitle       = i18n.MustString("nago.ai.mistralai.settings_title", i18n.Values{language.English: "My Mistral AI Token", language.German: "Mein Mistral AI Token"})
	StrMistralAISettingsName        = i18n.MustString("nago.ai.mistralai.settings_name", i18n.Values{language.English: "Mistral AI", language.German: "mistral AI"})
	StrMistralAISettingsDescription = i18n.MustString("nago.ai.mistralai.settings_desc", i18n.Values{language.English: "API Token to connect to Mistral AI", language.German: "Token zur API Anbindung von Mistral AI"})
	StrMistralAISettingsRPS         = i18n.MustString("nago.ai.mistralai.settings_rps", i18n.Values{language.English: "Requests per Second", language.German: "Anfragen pro Sekunde"})
	StrMistralAISettingsRPSDesc     = i18n.MustString("nago.ai.mistralai.settings_rps_desc", i18n.Values{language.English: "Limit the rate of requests against the API", language.German: "Anfragebegrenzung pro Sekunde an die API."})
)

type Settings struct {
	Name        string   `value:"nago.ai.mistralai.settings_title"`
	Description string   `label:"nago.common.label.description" lines:"3"`
	Token       string   `style:"secret"`
	RPS         int      `label:"nago.ai.mistralai.settings_rps" supportingText:"nago.ai.mistralai.settings_rps_desc" json:"rps"`
	Debug       bool     `json:"debug"`
	_           struct{} `credentialName:"nago.ai.mistralai.settings_name" credentialDescription:"nago.ai.mistralai.settings_desc" credentialLogo:"https://mistral.ai/favicon.ico?favicon.95e802d0.ico"`
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
