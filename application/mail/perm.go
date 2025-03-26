// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import "go.wdy.de/nago/application/permission"

var (
	PermSendMail             = permission.Declare[SendMail]("nago.mail.send", "Mail Senden", "Träger dieser Berechtigung können Mails versenden.")
	PermInitDefaultTemplates = permission.Declare[SendMail]("nago.mail.init_default_templates", "Standard Templates setzen", "Träger dieser Berechtigung können die Standard Mail templates aktivieren.")

	PermOutgoingFindAll    permission.ID
	PermOutgoingFindByID   permission.ID
	PermOutgoingDeleteByID permission.ID
)
