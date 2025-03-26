// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package permission

import "iter"

type FindAll func(subject Auditable) iter.Seq2[Permission, error]

type UseCases struct {
	FindAll FindAll
}

func NewUseCases() UseCases {
	return UseCases{
		FindAll: NewFindAll(),
	}
}
