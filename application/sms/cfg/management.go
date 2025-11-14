// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgsms

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/sms"
	"go.wdy.de/nago/application/sms/message"
	uisms "go.wdy.de/nago/application/sms/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

var (
	StrTestSMSTest              = i18n.MustString("nago.sms.admin.test", i18n.Values{language.English: "Send SMS", language.German: "SMS versenden"})
	StrTestSMSTestDesc          = i18n.MustString("nago.sms.admin.test_desc", i18n.Values{language.English: "Test and send your SMS provider.", language.German: "SMS Provider testen und versenden."})
	StrMaintenanceAdminCardDesc = i18n.MustString("nago.sms.admin.maintenance_desc", i18n.Values{language.English: "Apply some maintenance tasks to the SMS subsystem.", language.German: "Wartungsarbeiten am SMS Subsystem durchführen."})
)

type Management struct {
	UseCases sms.UseCases
	Pages    uisms.Pages
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

	repoSMS, err := application.JSONRepository[message.SMS](cfg, "nago.sms.message")
	if err != nil {
		return Management{}, err
	}

	ucSMS := sms.NewUseCases(cfg.Context(), cfg.EventBus(), secrets.UseCases.FindGroupSecrets, repoSMS)

	management = Management{
		UseCases: ucSMS,
		Pages: uisms.Pages{
			Queue: "admin/sms/queue",
			Send:  "admin/sms/send",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.Queue, func(wnd core.Window) core.View {
		return uisms.PageQueue(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.Pages.Send, func(wnd core.Window) core.View {
		return uisms.PageSend(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {

		grp := admin.Group{
			Title: "SMS",
		}

		grp.Entries = append(grp.Entries, admin.Card{
			Title:      rstring.LabelMaintenance.Get(subject),
			Text:       StrMaintenanceAdminCardDesc.Get(subject),
			Target:     management.Pages.Queue,
			Permission: sms.PermFindAllMessageIDs,
		})

		grp.Entries = append(grp.Entries, admin.Card{
			Title:      StrTestSMSTest.Get(subject),
			Text:       StrTestSMSTestDesc.Get(subject),
			Target:     management.Pages.Send,
			Permission: sms.PermSend,
		})

		return grp
	})

	//	cfg.AddContextValue(core.ContextValue("nago.ai", management))

	slog.Info("installed SMS module")
	return management, nil
}
