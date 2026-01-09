// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
)

func NewDeleteByPrimary[Evt any](perms Permissions, delete Delete[Evt], opts Options[Evt]) DeleteByPrimary[Evt] {
	return func(subject auth.Subject, index IdxID, primary string) error {
		if err := subject.Audit(perms.Delete); err != nil {
			return err
		}

		var idxer Indexer[Evt]
		for _, i := range opts.Indexer {
			if i.Info().ID == index {
				idxer = i
				break
			}
		}

		if idxer == nil {
			// semantically all removed
			return nil
		}

		idents, err := xslices.Collect2(idxer.GroupByPrimary(primary))
		if err != nil {
			return err
		}

		for _, key := range idents {
			id, err := key.Parse()
			if err != nil {
				return err
			}

			if err := delete(subject, id); err != nil {
				return err
			}
		}

		return nil
	}
}
