// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package aam

import (
	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/auth"
)

type Create func(subject auth.Subject, env environment.ID, app app.ID) (*App, error)

type UseCases struct {
	Create Create
}

func NewUseCases(replay environment.Replay) UseCases {
	return UseCases{
		Create: NewCreate(replay),
	}
}
