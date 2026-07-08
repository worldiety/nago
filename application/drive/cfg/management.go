// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgdrive

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/drive"
	drivehttp "go.wdy.de/nago/application/drive/http"
	uidrive "go.wdy.de/nago/application/drive/ui"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

var (
	strResFiles     = i18n.MustString("nago.drive.resources.name", i18n.Values{language.German: "Drive Dateien", language.English: "Drive files"})
	strResFilesDesc = i18n.MustString("nago.drive.resources.desc", i18n.Values{language.German: "Dateien und Ordner mit ihren Zugriffsrechten (ACL).", language.English: "Files and folders with their access rights (ACL)."})
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

	fileRepo, err := application.JSONRepository[drive.File, drive.FID](cfg, string(drive.FileNamespace))
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

	rdb, err := cfg.RDB()
	if err != nil {
		return Management{}, err
	}

	// Register the ReBAC static rules that allow granting per-file resource permissions to users and groups
	// (required before rebac.DB.Put may write those triples, see drive.GrantFileAccess). The global
	// user/group -> global rules for these permissions are already registered by the user management for
	// every declared permission.
	for _, ns := range []rebac.Namespace{user.Namespace, group.Namespace} {
		for _, pid := range drive.ACLPermissions {
			rdb.RegisterStaticRule(rebac.StaticRule{
				Source:   ns,
				Relation: rebac.Relation(pid),
				Target:   drive.FileNamespace,
			})
		}
	}

	// Make drive files browsable and manageable in the general ReBAC editor. The mapper renders a
	// human-readable path/name for each file instance.
	rdb.RegisterResources(rebac.NewRepositoryResources(strResFiles, strResFilesDesc, fileRepo).Map(func(f drive.File) rebac.InstanceInfo {
		name := f.Name()
		if path, err := f.AbsolutePath(); err == nil && path != "" {
			name = path
		}

		return rebac.InstanceInfo{
			Namespace: drive.FileNamespace,
			ID:        rebac.Instance(f.Identity()),
			Name:      name,
		}
	}))

	uc := drive.NewUseCases(cfg.EventBus(), fileRepo, globalRootsRepo, userRootsRepo, fileBlobs, rdb)

	// Authenticated endpoint that streams a file's binary content, used by the UI preview (ui.Image,
	// video.Video) and downloads. Authorization is enforced by uc.Get (CanRead) for the resolved subject.
	if err := cfg.HandleFuncSubject(drivehttp.Endpoint, drivehttp.NewHandler(uc.Get)); err != nil {
		return Management{}, err
	}

	management = Management{
		UseCases: uc,
		Pages:    uidrive.Pages{},
	}

	// Register the Management under its own type so a repeated Enable() short-circuits at the top (the
	// idempotency check reads core.FromContext[Management]). This must be set, otherwise a second Enable
	// (e.g. once directly and once transitively via cfgai.Enable) would run again and attempt to Mount the
	// drive download handler a second time, panicking chi with "attempting to Mount() a handler on an
	// existing path".
	cfg.AddContextValue(core.ContextValue("nago.drive.management", management))

	// The bare UseCases is additionally exposed for consumers that resolve it by type (e.g. the drive UI).
	cfg.AddContextValue(core.ContextValue("nago.drive", management.UseCases))

	slog.Info("installed drive management")

	return management, nil
}
