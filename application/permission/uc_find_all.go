// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package permission

import (
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
	"slices"
)

func NewFindAll() FindAll {
	return func(subject Auditable) iter.Seq2[Permission, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[Permission](err)
		}

		return xslices.ValuesWithError(slices.Collect(All()), nil)
	}
}
