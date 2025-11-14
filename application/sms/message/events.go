// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package message

type SendRequested struct {
	// ProviderHint allows to narrow the wanted provider, e.g. for specific originator signatures or credit accounts.
	// The hint is matched against the [provider.provider.Name] or [provider.provider.ID]. Note that providers
	// are created by the according secrets shared with [group.System].
	// If empty or no match was found, the first found provider is used.
	ProviderHint string

	// Recipient for this message.
	Recipient MSISDN `json:"recipient,omitempty"`

	// Originator of the message
	Originator Originator `json:"originator,omitempty"`

	// Body is the actual message. Usually limited to 160 bytes, less for unicode. Depending on the provider,
	// a longer body may be split and send using 2-10 distinct SMS. Note, that each one causes additional costs.
	Body string `json:"body,omitempty"`
}
