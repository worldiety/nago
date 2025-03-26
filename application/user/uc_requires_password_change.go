// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import "fmt"

func NewRequiresPasswordChange(sysUser SysUser, findByID FindByID) RequiresPasswordChange {
	return func(uid ID) (bool, error) {
		optUser, err := findByID(sysUser(), uid)
		if err != nil {
			return false, err
		}

		if optUser.IsNone() {
			// not sure if we should return a nice type
			return false, fmt.Errorf("user not found: %v", uid)
		}

		usr := optUser.Unwrap()
		
		return usr.RequirePasswordChange, nil
	}
}
