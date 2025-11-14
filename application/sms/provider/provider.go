// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/auth"
)

type ID string

type Provider interface {
	Send(subject auth.Subject, sms message.SendRequested) (message.ID, error)
	Name() string
	Identity() ID
}
