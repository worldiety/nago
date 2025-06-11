// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

const (
	InvalidSubjectErr                            InvalidSubjectError                            = "invalid subject"
	PermissionDeniedErr                          PermissionDeniedError                          = "permission denied"
	NewPasswordMustBeDifferentFromOldPasswordErr NewPasswordMustBeDifferentFromOldPasswordError = "new password must be different from old password"
	PasswordsDontMatchErr                        PasswordsDontMatchError                        = "passwords dont match"
	InvalidOldPasswordErr                        InvalidOldPasswordError                        = "invalid old password"
)

type InvalidSubjectError string

func (e InvalidSubjectError) Error() string {
	return string(e)
}

func (e InvalidSubjectError) PermissionDenied() bool {
	return true
}

func (e InvalidSubjectError) NotLoggedIn() bool {
	return true
}

type PermissionDeniedError string

func (e PermissionDeniedError) Error() string {
	return string(e)
}

func (e PermissionDeniedError) PermissionDenied() bool {
	return true
}

type NewPasswordMustBeDifferentFromOldPasswordError string

func (e NewPasswordMustBeDifferentFromOldPasswordError) Error() string {
	return string(e)
}

type PasswordsDontMatchError string

func (e PasswordsDontMatchError) Error() string {
	return string(e)
}

type InvalidOldPasswordError string

func (e InvalidOldPasswordError) Error() string {
	return string(e)
}
