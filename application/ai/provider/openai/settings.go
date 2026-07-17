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

// NOTE: Unlike the other providers, openai currently ships only its credentials Settings and has no working
// Provider implementation (no provider.go / NewProvider). Once a NewProvider(id provider.ID, cfg Settings)
// provider.Provider exists, add the registration next to the enum.Variant above so the provider becomes
// available when this package is side-imported:
//
//	var _ = registerProvider()
//	func registerProvider() any { provider.Register[Settings](NewProvider); return nil }

var (
	StrMyOpenAiSettingsTitle       = i18n.MustString("nago.ai.openai.settings_title", i18n.Values{language.English: "My OpenAI Token", language.German: "Mein OpenAI Token"})
	StrMyOpenAiSettingsName        = i18n.MustString("nago.ai.openai.settings_name", i18n.Values{language.English: "OpenAI", language.German: "OpenAI"})
	StrMyOpenAiSettingsDescription = i18n.MustString("nago.ai.openai.settings_desc", i18n.Values{language.English: "API Token to connect to OpenAI", language.German: "Token zur API Anbindung von OpenAI"})
)

type Settings struct {
	Name  string   `value:"nago.ai.openai.settings_title"`
	Token string   `style:"secret"`
	_     struct{} `credentialName:"nago.ai.openai.settings_name" credentialDescription:"nago.ai.openai.settings_desc" credentialLogo:"https://openai.com/favicon.svg"`
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
