// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import "go.wdy.de/nago/application/user"

type ID string

type Email = user.Email

type User struct {
	ID        ID         `json:"id"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	Nickname  string     `json:"nickname"`
	Email     user.Email `json:"email"`
}
