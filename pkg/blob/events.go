// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package blob

type Written struct {
	Store string // name of the store
	Key   string // key of the affected entry
}

type Deleted struct {
	Store string
	Key   string
}
