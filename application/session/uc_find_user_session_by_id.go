// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

func NewFindUserSessionByID(repository Repository) FindUserSessionByID {
	return func(id ID) UserSession {
		return &sessionImpl{id: id, repo: repository}
	}
}
