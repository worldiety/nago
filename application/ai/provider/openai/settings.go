// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

var _ = enum.Variant[secret.Credentials, Settings](
	enum.Rename[Settings]("nago.ai.openai.settings"),
)

var (
	StrMyOpenAiSettingsTitle       = i18n.MustString("nago.ai.openai.settings_title", i18n.Values{language.English: "My OpenAI Token", language.German: "Mein OpenAI Token"})
	StrMyOpenAiSettingsName        = i18n.MustString("nago.ai.openai.settings_name", i18n.Values{language.English: "OpenAI", language.German: "OpenAI"})
	StrMyOpenAiSettingsDescription = i18n.MustString("nago.ai.openai.settings_desc", i18n.Values{language.English: "API Token to connect to OpenAI", language.German: "Token zur API Anbindung von OpenAI"})
)

type Settings struct {
	Name  string `value:"nago.ai.openai.settings_title"`
	Token string `style:"secret"`
	_     string `credentialName:"nago.ai.openai.settings_name" credentialDescription:"nago.ai.openai.settings_desc" credentialLogo:"https://openai.com/favicon.svg"`
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
