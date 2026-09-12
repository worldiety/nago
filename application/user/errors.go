// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/permission"
)

const (
	InvalidSubjectErr                            InvalidSubjectError                            = "invalid subject"
	PermissionDeniedErr                          PermissionDeniedError                          = "permission denied"
	NewPasswordMustBeDifferentFromOldPasswordErr NewPasswordMustBeDifferentFromOldPasswordError = "new password must be different from old password"
	PasswordsDontMatchErr                        PasswordsDontMatchError                        = "passwords dont match"
	InvalidOldPasswordErr                        InvalidOldPasswordError                        = "invalid old password"
	InvalidEMailErr                              InvalidEMailError                              = "invalid email"
	EMailAlreadyInUseErr                         EMailAlreadyInUseError                         = "email already in use"
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

// PermissionDeniedError signals that an operation was refused for authorization reasons.
//
// Historically this type carries two different kinds of payload: a [permission.ID] such as
// "nago.user.find_by_id", and free form prose such as "workspace forbids access". Use
// [PermissionDeniedError.RequiredPermission] to distinguish them instead of interpreting the
// string yourself.
type PermissionDeniedError string

func (e PermissionDeniedError) Error() string {
	return string(e)
}

func (e PermissionDeniedError) PermissionDenied() bool {
	return true
}

// RequiredPermission reports the permission which would have been required, if this error
// carries one rather than a free form reason.
//
// The discrimination is syntactic and therefore approximate: [permission.ID.Valid] accepts any
// lower case dotted identifier, so prose such as "workspace forbids access" is rejected, but a
// single lower case word such as "forbidden" would be accepted. Callers must treat the result
// as a candidate and resolve it through the registry; [permission.RequiredOf] does exactly
// that and reports false for anything unregistered.
//
// The ambiguity is inherent to this type carrying two disjoint meanings in one string. Use
// [permission.RequiredOf] rather than this method directly.
//
// This implements the optional permission.RequiredPermission interface.
func (e PermissionDeniedError) RequiredPermission() (permission.ID, bool) {
	if e == PermissionDeniedErr {
		return "", false
	}

	id := permission.ID(e)
	if !id.Valid() {
		return "", false
	}

	return id, true
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

type InvalidEMailError string

func (e InvalidEMailError) Error() string {
	return string(e)
}

type EMailAlreadyInUseError string

func (e EMailAlreadyInUseError) Error() string {
	return string(e)
}
