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

	management = Management{
		UseCases: uc,
		Pages:    uidrive.Pages{},
	}

	cfg.AddContextValue(core.ContextValue("nago.drive", management.UseCases))

	slog.Info("installed drive management")

	return management, nil
}
