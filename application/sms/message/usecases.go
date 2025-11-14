// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package message

import (
	"strconv"
	"strings"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type Repository data.Repository[SMS, ID]

type SMS struct {
	ID ID `json:"id"`

	Status          Status `json:"status"`
	Provider        string `json:"sendBy"`             // provider.ID would cause a cycle
	ProviderMessage ID     `json:"sendByID,omitempty"` // provider may returns his own ID, whatever that is

	LastError string `json:"lastError,omitempty"`

	// ProviderHint allows to narrow the wanted provider, e.g. for specific originator signatures or credit accounts.
	// The hint is matched against the [provider.provider.Name] or [provider.provider.ID] of all secrets available to
	// [group.System].
	// If no match was found, the first found provider secret shared with [group.System] is used.
	ProviderHint string

	// Recipients for this message.
	Recipient MSISDN `json:"recipients,omitempty"`

	// Originator of the message
	Originator Originator `json:"originator,omitempty"`

	// Body is the actual message. Usually limited to 160 bytes, less for unicode. Depending on the provider,
	// a longer body may be split and send using 2-10 distinct SMS. Note, that each one causes additional costs.
	Body string `json:"body,omitempty"`

	CreatedAt xtime.UnixMilliseconds `json:"createdAt,omitempty"`
	SendAt    xtime.UnixMilliseconds `json:"sendAt,omitempty"`
}

// MSISDN is a special number format like 49179555111XXX.
type MSISDN int64

func NewMSISDN(str string) (MSISDN, error) {
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "+") {
		str = str[1:]
	}

	str = strings.ReplaceAll(str, " ", "")
	i, err := strconv.ParseInt(str, 10, 64)
	return MSISDN(i), err
}

func (s MSISDN) String() string {
	return strconv.FormatInt(int64(s), 10)
}

// Originator is limited to 11 alphanumeric chars or max 14 digits. Not all providers support the inclusion of
// a custom sender.
type Originator string

type ID string

type Status string

const (
	StatusQueued Status = "queued"
	StatusSent   Status = "sent"
	StatusFailed Status = "failed"
)

func (s SMS) Identity() ID {
	return s.ID
}
