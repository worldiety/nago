// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import "go.wdy.de/nago/application/permission"

var (
	PermReloadProvider    = permission.DeclareReloadAll[ReloadProvider]("nago.sms.provider.reload", "SMS Provider")
	PermSend              = permission.DeclareSend[Send]("nago.sms.provider.send", "SMS")
	PermFindAllMessageIDs = permission.DeclareFindAllIdentifiers[FindAllMessageIDs]("nago.sms.provider.find_all_idents", "SMS")
	PermFindByID          = permission.DeclareFindByID[FindMessageByID]("nago.sms.provider.find_by_id", "SMS")
	PermDeleteMessageByID = permission.DeclareDeleteByID[DeleteMessageByID]("nago.sms.provider.delete_by_id", "SMS")
)
