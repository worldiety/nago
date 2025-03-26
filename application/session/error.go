// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

const (
	NotLoggedInErr NotLoggedInError = "not logged in"
)

type NotLoggedInError string

func (e NotLoggedInError) Error() string {
	return string(e)
}

func (e NotLoggedInError) PermissionDenied() bool {
	return true
}

func (e NotLoggedInError) NotLoggedIn() bool {
	return true
}
