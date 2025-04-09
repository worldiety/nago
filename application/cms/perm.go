// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import "go.wdy.de/nago/application/permission"

var (
	PermCreate          = permission.Declare[Create]("nago.cms.doc.create", "CMS Dokument erstellen", "Träger dieser Berechtigung können ein CMS Dokument erstellen.")
	PermDelete          = permission.Declare[Delete]("nago.cms.doc.delete", "CMS Dokument löschen", "Träger dieser Berechtigung können ein CMS Dokument löschen.")
	PermUpdateSlug      = permission.Declare[UpdateSlug]("nago.cms.doc.slug.update", "CMS Dokument-Slug aktualisieren", "Träger dieser Berechtigung können den Slug eines Dokumentes aktualisieren.")
	PermUpdatePublished = permission.Declare[UpdatePublished]("nago.cms.doc.published.update", "CMS Dokument-Published aktualisieren", "Träger dieser Berechtigung können die Veröffentlichung eines Dokumentes aktualisieren.")
	PermUpdateTitle     = permission.Declare[UpdateTitle]("nago.cms.doc.title.update", "CMS Dokument-Titel aktualisieren", "Träger dieser Berechtigung können den Titel eines Dokumentes aktualisieren.")
	PermFindAll         = permission.Declare[FindAll]("nago.cms.doc.find_all", "CMS Dokumente auflisten", "Träger dieser Berechtigung können CMS Dokumente auflisten.")
	PermAppendElement   = permission.Declare[AppendElement]("nago.cms.elem.append", "CMS Dokument Element anhängen", "Träger dieser Berechtigung können ein Element an ein CMS Dokument anhängen.")
	PermUpdateElement   = permission.Declare[UpdateElement]("nago.cms.elem.update", "CMS Dokument Element aktualisieren", "Träger dieser Berechtigung können ein Element aktualisieren.")
	PermReplaceElement  = permission.Declare[ReplaceElement]("nago.cms.elem.replace", "CMS Dokument Element ersetzen", "Träger dieser Berechtigung können ein Element ersetzen.")
	PermDeleteElement   = permission.Declare[DeleteElement]("nago.cms.elem.delete", "CMS Dokument Element entfernen", "Träger dieser Berechtigung können ein Element entfernen.")
)
