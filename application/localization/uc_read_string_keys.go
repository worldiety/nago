// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/auth"
)

func NewReadStringKeys(res *i18n.Resources) ReadStringKeys {
	return func(subject auth.Subject) ([]i18n.Key, error) {
		if err := subject.Audit(PermReadDir); err != nil {
			return nil, err
		}

		keys := res.SortedKeys()

		var tmp []i18n.Key
		for _, key := range keys {
			if !key.StringKey() {
				continue
			}
			tmp = append(tmp, key)
		}

		return tmp, nil
	}
}
