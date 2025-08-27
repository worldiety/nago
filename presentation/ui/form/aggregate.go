// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import "go.wdy.de/nago/pkg/data"

// Aggregate defines a generic aggregate interface.
// It extends data.Aggregate by requiring a method to return
// a new instance with the specified identity.
type Aggregate[A any, ID comparable] interface {
	data.Aggregate[ID]
	WithIdentity(ID) A
}
