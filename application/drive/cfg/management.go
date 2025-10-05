// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgdrive

import (
	"log/slog"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/drive"
	uidrive "go.wdy.de/nago/application/drive/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
)

type Management struct {
	UseCases drive.UseCases
	Pages    uidrive.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	fileRepo, err := application.JSONRepository[drive.File, drive.FID](cfg, "nago.drive.file")
	if err != nil {
		return Management{}, err
	}

	globalRootsRepo, err := application.JSONRepository[drive.NamedRoot, string](cfg, "nago.drive.root.global")
	if err != nil {
		return Management{}, err
	}

	userRootsRepo, err := application.JSONRepository[drive.UserRoots, user.ID](cfg, "nago.drive.root.user")
	if err != nil {
		return Management{}, err
	}

	fileBlobs, err := cfg.FileStore("nago.drive.blob")
	if err != nil {
		return Management{}, err
	}

	uc := drive.NewUseCases(fileRepo, globalRootsRepo, userRootsRepo, fileBlobs)

	management = Management{
		UseCases: uc,
		Pages:    uidrive.Pages{},
	}

	cfg.AddContextValue(core.ContextValue("nago.drive", management.UseCases))

	slog.Info("installed drive management")

	return management, nil
}
