// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

type ID string

type Library struct {
	ID          ID
	Name        string
	Description string
}

type CreateOptions struct {
	Name        string
	Description string
}
