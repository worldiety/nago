// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package image

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

// LoadSrcSet just returns the according source set, if any.
type LoadSrcSet func(user permission.Auditable, id ID) (std.Option[SrcSet], error)

func NewLoadSrcSet(repository Repository) LoadSrcSet {
	return func(user permission.Auditable, id ID) (std.Option[SrcSet], error) {
		// TODO solve permission questions
		optSet, err := repository.FindByID(id)
		if err != nil {
			return std.None[SrcSet](), fmt.Errorf("failed to load SrcSet from repo: %w", err)
		}

		return optSet, nil
	}
}
