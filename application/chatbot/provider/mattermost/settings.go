// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mattermost

import (
	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/secret"
	"golang.org/x/text/language"
)

var (
	StrMattermostSettingsTitle        = i18n.MustString("nago.chatbot.mattermost.settings_title", i18n.Values{language.English: "My Mattermost Token", language.German: "Mein Mattermost Token"})
	StrMattermostSettingsName         = i18n.MustString("nago.chatbot.mattermost.settings_name", i18n.Values{language.English: "Mattermost Chatbot", language.German: "Mattermost Chatbot"})
	StrSMattermostSettingsDescription = i18n.MustString("nago.chatbot.mattermost.settings_desc", i18n.Values{language.English: "API Token to connect to Mattermost", language.German: "Token zur API Anbindung von Mattermost"})
)

type Settings struct {
	Name  string `value:"nago.chatbot.mattermost.settings_title" json:"name"`
	URL   string `json:"url"`
	Token string `json:"token"`
	RPS   int    `json:"rps"`
	_     string `credentialName:"nago.chatbot.mattermost.settings_name" credentialDescription:"nago.chatbot.mattermost.settings_desc" credentialLogo:"https://mattermost.worldiety.net/static/images/favicon/favicon-default-32x32.png"`
}

var _ = enum.Variant[secret.Credentials, Settings](enum.Rename[Settings]("nago.chatbot.mattermost.settings"))

func (s Settings) GetName() string {
	return s.Name
}

func (s Settings) Credentials() bool {
	return true
}

func (s Settings) IsZero() bool {
	return Settings{} == s
}
