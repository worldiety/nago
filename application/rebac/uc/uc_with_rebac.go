// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ucrebac

import (
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
)

func NewWithReBAC(rdb *rebac.DB) WithReBAC {
	return func(subject user.Subject, fn func(rdb *rebac.DB) error) error {
		if err := subject.Audit(PermWithReBAC); err != nil {
			return err
		}

		return fn(rdb)
	}
}
