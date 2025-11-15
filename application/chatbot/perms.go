// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chatbot

import "go.wdy.de/nago/application/permission"

var (
	PermReloadProvider = permission.DeclareReloadAll[ReloadProvider]("nago.chatbot.provider.reload", "Chatbot Provider")
	PermSend           = permission.DeclareSend[Send]("nago.chatbot.provider.send", "Chatbot Message")
)
