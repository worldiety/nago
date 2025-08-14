// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgsignature

import (
	_ "embed"
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/signature"
	"go.wdy.de/nago/application/signature/document"
	uisignature "go.wdy.de/nago/application/signature/ui"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	cfgworkflow "go.wdy.de/nago/application/workflow/cfg"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

//go:embed caveat.ttf
var fntCaveat application.StaticBytes

type Management struct {
	UseCases signature.UseCases
	Pages    uisignature.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	stores, err := cfg.Stores()
	if err != nil {
		return Management{}, fmt.Errorf("failed to get stores: %w", err)
	}

	images, err := cfg.ImageManagement()
	if err != nil {
		return Management{}, fmt.Errorf("failed to get image management: %w", err)
	}

	settingsRepo, err := application.JSONRepository[signature.UserSettings](cfg, "nago.signature.settings")
	if err != nil {
		return Management{}, err
	}

	signatureRepo, err := application.JSONRepository[signature.Signature](cfg, "nago.signature")
	if err != nil {
		return Management{}, err
	}

	uc, err := signature.NewUseCases(stores, signatureRepo, settingsRepo, images.UseCases.OpenReader)
	if err != nil {
		return Management{}, fmt.Errorf("failed to create signature usecases: %w", err)
	}

	workflows, err := cfgworkflow.Enable(cfg)
	if err != nil {
		return Management{}, err
	}

	if _, err := workflows.UseCases.Declare(user.SU(), document.NewWorkflow()); err != nil {
		return Management{}, fmt.Errorf("failed to declare workflow: %w", err)
	}

	management = Management{
		UseCases: uc,
		Pages: uisignature.Pages{
			MyImageSignature: "admin/my-signature/settings",
		},
	}

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		return admin.Group{
			Title: "Elektronische Signaturen",
			Entries: []admin.Card{
				{
					Title:  "Meine Unterschrift",
					Text:   "Um symbolisiert eine Unterschrift zu zeigen, kann hier ein digitales Bild der Unterschrift eingestellt werden.",
					Target: management.Pages.MyImageSignature,
				},
			},
		}
	})

	cfg.RootViewWithDecoration(management.Pages.MyImageSignature, func(wnd core.Window) core.View {
		return uisignature.PageMySignature(wnd, management.UseCases, images.UseCases.CreateSrcSet)
	})

	cfg.AddContextValue(core.ContextValue("nago.signatures", management))
	cfg.AddContextValue(core.ContextValue("nago.signatures.usecases", management.UseCases))

	uriCaveat := cfg.Resource(fntCaveat)

	cfgTheme := settings.ReadGlobal[theme.Settings](option.Must(cfg.SettingsManagement()).UseCases.LoadGlobal)
	if !cfgTheme.Fonts.Contains("Caveat") {
		cfgTheme.Fonts.Faces = append(cfgTheme.Fonts.Faces,
			core.FontFace{
				Family: "Caveat",
				Source: uriCaveat,
			},
		)

		settings.WriteGlobal(option.Must(cfg.SettingsManagement()).UseCases.StoreGlobal, cfgTheme)
	}

	slog.Info("installed signature management")

	return management, nil
}
