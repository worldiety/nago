// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package channel

type ID string

type Channel struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}
