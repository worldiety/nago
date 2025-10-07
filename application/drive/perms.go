// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/permission"
	"golang.org/x/text/language"
)

var (
	// TODO this permission is not understandable
	PermOpenFile = permission.Declare[OpenRoot](
		"nago.drive.open_file",
		i18n.MustString(
			"nago.permissions.drive.open_file",
			i18n.Values{
				language.English: "Open a drive file",
				language.German:  "Eine Drive Datei öffnen",
			},
		).String(),
		i18n.MustString(
			"nago.permissions.drive.open_file_desc",
			i18n.Values{
				language.English: "Holders of this authorisation can open a drive file. This may also include create, read, update or delete operations depending on the actual file permissions. When globally assigned, this grants essentially root rights for all drives in the system.",
				language.German:  "Träger dieser Berechtigung können einen Drive-Datei öffnen. Das kann auch das Erstellen, Lesen, Aktualisieren oder Löschen von Dateien basierend auf den tatsächlichen Berechtigungen beinhalten. Sofern global gesetzt, erhält ein Nutzer Root-Rechte für alle Dateien im System.",
			},
		).String(),
	)

	PermMkDir = permission.Declare[OpenRoot](
		"nago.drive.mkdir",
		i18n.MustString(
			"nago.permissions.drive.mkdir",
			i18n.Values{
				language.English: "Create a folder",
				language.German:  "Einen Ordner erstellen",
			},
		).String(),
		i18n.MustString(
			"nago.permissions.drive.mkdir_desc",
			i18n.Values{
				language.English: "Holders of this authorisation can create a folder in the assigned parent folder.",
				language.German:  "Träger dieser Berechtigung können einen Ordner im zugewiesenen Elternordner erstellen.",
			},
		).String(),
	)

	PermDelete = permission.Declare[OpenRoot](
		"nago.drive.delete",
		i18n.MustString(
			"nago.permissions.drive.delete",
			i18n.Values{
				language.English: "Delete a file or folder",
				language.German:  "Eine Datei oder Ordner löschen",
			},
		).String(),
		i18n.MustString(
			"nago.permissions.drive.delete_desc",
			i18n.Values{
				language.English: "Holders of this authorisation can delete an assigned folder or file.",
				language.German:  "Träger dieser Berechtigung können einen zugewiesenen Ordner oder eine Datei löschen.",
			},
		).String(),
	)

	PermPut = permission.Declare[OpenRoot](
		"nago.drive.put",
		i18n.MustString(
			"nago.permissions.drive.put",
			i18n.Values{
				language.English: "Create or update a file",
				language.German:  "Eine Datei erstellen oder aktualisieren",
			},
		).String(),
		i18n.MustString(
			"nago.permissions.drive.put_desc",
			i18n.Values{
				language.English: "Holders of this authorisation can create or update a file.",
				language.German:  "Träger dieser Berechtigung können Datei erstellen oder aktualisieren.",
			},
		).String(),
	)
)
