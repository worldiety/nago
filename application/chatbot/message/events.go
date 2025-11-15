// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package message

import "go.wdy.de/nago/application/chatbot/user"

type SendRequested struct {
	// ProviderHint allows to narrow the wanted provider, e.g. for specific originator signatures or credit accounts.
	// The hint is matched against the [provider.provider.Name] or [provider.provider.ID]. Note that providers
	// are created by the according secrets shared with [group.System].
	// If empty or no match was found, the first found provider is used.
	ProviderHint string `json:"providerHint"`

	RecipientByID   user.ID    `json:"recipientByID,omitempty"`
	RecipientByMail user.Email `json:"RecipientByMail,omitempty"`

	// Text is the actual message.
	Text string `json:"body"`
}
