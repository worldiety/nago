// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgchatbot

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/chatbot"
	uichatbot "go.wdy.de/nago/application/chatbot/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

var (
	StrTestSMSTest              = i18n.MustString("nago.chatbot.admin.test", i18n.Values{language.English: "Send message", language.German: "Nachricht versenden"})
	StrTestSMSTestDesc          = i18n.MustString("nago.chatbot.admin.test_desc", i18n.Values{language.English: "Test and send your Chatbot provider.", language.German: "Chatbot Provider testen und versenden."})
	StrMaintenanceAdminCardDesc = i18n.MustString("nago.chatbot.admin.maintenance_desc", i18n.Values{language.English: "Apply some maintenance tasks to the Chatbot subsystem.", language.German: "Wartungsarbeiten am Chatbot Subsystem durchführen."})
)

type Management struct {
	UseCases chatbot.UseCases
	Pages    uichatbot.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	secrets, err := cfg.SecretManagement()
	if err != nil {
		return Management{}, err
	}

	ucSMS := chatbot.NewUseCases(cfg.Context(), cfg.EventBus(), secrets.UseCases.FindGroupSecrets)

	management = Management{
		UseCases: ucSMS,
		Pages: uichatbot.Pages{
			Send: "admin/chatbot/send",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.Send, func(wnd core.Window) core.View {
		return uichatbot.PageSend(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {

		grp := admin.Group{
			Title: "Chatbot",
		}

		grp.Entries = append(grp.Entries, admin.Card{
			Title:      StrTestSMSTest.Get(subject),
			Text:       StrTestSMSTestDesc.Get(subject),
			Target:     management.Pages.Send,
			Permission: chatbot.PermSend,
		})

		return grp
	})

	//	cfg.AddContextValue(core.ContextValue("nago.ai", management))

	slog.Info("installed chatbot module")
	return management, nil
}
