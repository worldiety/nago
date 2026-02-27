// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ucrebac

import (
	"iter"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
)

type FindAllResources func(subject user.Subject) iter.Seq2[rebac.Resources, error]

type WithReBAC func(subject user.Subject, fn func(rdb *rebac.DB) error) error

type UseCases struct {
	FindAllResources FindAllResources
	WithReBAC        WithReBAC
}

func NewUseCases(rdb *rebac.DB) UseCases {
	return UseCases{
		FindAllResources: NewFindAllResources(rdb),
		WithReBAC:        NewWithReBAC(rdb),
	}
}
