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

// Management is a nago system(Drive Management).
// It provides a file storage and file management subsystem that can be used from both
// frontend UI components and backend use-cases.
//
// The system implements an owner/group/permission model. Files contain Owner, Group and FileMode fields,
// and permission checks are implemented in File.CanRead, File.CanWrite, File.CanDelete, and File.CanRename.
// Shares and resource-level permissions are also considered by the permission checks.
//
// Use the provided use-cases (drive.OpenRoot, drive.Put, drive.MkDir, drive.Delete, drive.Stat, drive.Zip, drive.Get, drive.Rename)
// to integrate Drive into your application logic or to expose it through custom APIs.
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

	uc := drive.NewUseCases(cfg.EventBus(), fileRepo, globalRootsRepo, userRootsRepo, fileBlobs)

	management = Management{
		UseCases: uc,
		Pages:    uidrive.Pages{},
	}

	cfg.AddContextValue(core.ContextValue("nago.drive", management.UseCases))

	slog.Info("installed drive management")

	return management, nil
}
