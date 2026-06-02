// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

var _ = enum.Variant[secret.Credentials, Settings](
	enum.Rename[Settings]("nago.ai.anthropic.settings"),
)

var (
	StrAnthropicSettingsTitle       = i18n.MustString("nago.ai.anthropic.settings_title", i18n.Values{language.English: "My Anthropic Token", language.German: "Mein Anthropic Token"})
	StrAnthropicSettingsName        = i18n.MustString("nago.ai.anthropic.settings_name", i18n.Values{language.English: "Anthropic", language.German: "Anthropic"})
	StrAnthropicSettingsDescription = i18n.MustString("nago.ai.anthropic.settings_desc", i18n.Values{language.English: "API Token to connect to Anthropic (Claude)", language.German: "Token zur API Anbindung von Anthropic (Claude)"})
	StrAnthropicSettingsRPS         = i18n.MustString("nago.ai.anthropic.settings_rps", i18n.Values{language.English: "Requests per Second", language.German: "Anfragen pro Sekunde"})
	StrAnthropicSettingsRPSDesc     = i18n.MustString("nago.ai.anthropic.settings_rps_desc", i18n.Values{language.English: "Limit the rate of requests against the API", language.German: "Anfragebegrenzung pro Sekunde an die API."})
	StrAnthropicSettingsVersion     = i18n.MustString("nago.ai.anthropic.settings_version", i18n.Values{language.English: "API Version", language.German: "API Version"})
	StrAnthropicSettingsVersionDesc = i18n.MustString("nago.ai.anthropic.settings_version_desc", i18n.Values{language.English: "Value of the anthropic-version header. Leave empty for the built-in default.", language.German: "Wert des anthropic-version Headers. Leer lassen für den eingebauten Standardwert."})
	StrAnthropicSettingsMaxTokens   = i18n.MustString("nago.ai.anthropic.settings_max_tokens", i18n.Values{language.English: "Default max. output tokens", language.German: "Standard max. Ausgabe-Tokens"})
)

type Settings struct {
	Name        string   `value:"nago.ai.anthropic.settings_title"`
	Description string   `label:"nago.common.label.description" lines:"3"`
	Token       string   `style:"secret"`
	Version     string   `label:"nago.ai.anthropic.settings_version" supportingText:"nago.ai.anthropic.settings_version_desc" json:"version"`
	MaxTokens   int      `label:"nago.ai.anthropic.settings_max_tokens" json:"maxTokens"`
	RPS         int      `label:"nago.ai.anthropic.settings_rps" supportingText:"nago.ai.anthropic.settings_rps_desc" json:"rps"`
	Debug       bool     `json:"debug"`
	_           struct{} `credentialName:"nago.ai.anthropic.settings_name" credentialDescription:"nago.ai.anthropic.settings_desc" credentialLogo:"https://www.anthropic.com/favicon.ico"`
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

