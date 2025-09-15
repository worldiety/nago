// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfglocalization

import (
	"fmt"
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/localization"
	uilocalization "go.wdy.de/nago/application/localization/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type Management struct {
	UseCases localization.UseCases
	Pages    uilocalization.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	repo, err := application.JSONRepository[localization.StringData, i18n.Key](cfg, "nago.localization.stringdata")
	if err != nil {
		return Management{}, err
	}

	uc, err := localization.NewUseCases(repo, i18n.Default)
	if err != nil {
		return Management{}, fmt.Errorf("cannot create localization usecases: %w", err)
	}

	management.UseCases = uc
	management.Pages = uilocalization.Pages{
		PageDirectory: "admin/localization/directory",
		PageMessage:   "admin/localization/message",
		PageLanguage:  "admin/localization/language",
	}

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var grp admin.Group
		grp.Title = uilocalization.StrTranslations.Get(subject)
		dir, err := uc.ReadDir(subject, ".")
		if err != nil {
			slog.Error("failed to localization directory", "subject", subject, "err", err.Error())
			return grp
		}

		for _, info := range dir.Directories {
			grp.Entries = append(grp.Entries, admin.Card{
				Title:        info.Name,
				Text:         uilocalization.StrTranslationSecText.Get(subject, i18n.Int("totalAmount", info.TotalKeys), i18n.Int("missingAmount", info.TotalMissingKeys)),
				Target:       management.Pages.PageDirectory,
				TargetParams: core.Values{"path": string(info.Path)},
			})
		}

		strKeys, err := uc.ReadStringKeys(subject)
		if err != nil {
			slog.Error("failed to localization string keys", "subject", subject, "err", err.Error())
			return grp
		}

		if len(strKeys) > 0 {
			grp.Entries = append(grp.Entries, admin.Card{
				Title:        uilocalization.StrStringKeysTitle.Get(subject),
				Text:         uilocalization.StrStringKeysDesc.Get(subject),
				Target:       management.Pages.PageDirectory,
				TargetParams: core.Values{"stringkeys": "true"},
			})
		}

		grp.Entries = append(grp.Entries, admin.Card{
			Title:      uilocalization.StrLanguagesTitle.Get(subject),
			Text:       uilocalization.StrLanguagesDesc.Get(subject),
			Target:     management.Pages.PageLanguage,
			Permission: localization.PermAddLanguage,
		})

		return grp
	})

	cfg.RootViewWithDecoration(management.Pages.PageDirectory, func(wnd core.Window) core.View {
		return uilocalization.PageDir(wnd, uc)
	})

	cfg.RootViewWithDecoration(management.Pages.PageMessage, func(wnd core.Window) core.View {
		return uilocalization.PageMessage(wnd, uc)
	})

	cfg.RootViewWithDecoration(management.Pages.PageLanguage, func(wnd core.Window) core.View {
		return uilocalization.PageLanguage(wnd, uc)
	})

	cfg.AddContextValue(core.ContextValue("nago.localization", management))

	slog.Info("installed localization management")

	return management, nil
}
