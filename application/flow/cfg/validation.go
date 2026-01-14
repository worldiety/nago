// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import "github.com/worldiety/i18n"

type VID string
type Validation[T any] struct {
	ID          VID
	Label       i18n.StrHnd
	Description i18n.StrHnd
	Validate    func(T) error
}
