// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"errors"
	"fmt"
	"testing"

	"go.wdy.de/nago/application/permission"
)

func TestPermissionDeniedErrorRequiredPermission(t *testing.T) {
	tests := []struct {
		name   string
		err    PermissionDeniedError
		wantID permission.ID
		wantOk bool
	}{
		{
			name:   "bare sentinel names no permission",
			err:    PermissionDeniedErr,
			wantOk: false,
		},
		{
			// application/user/subject.go builds these from a permission.ID
			name:   "permission id",
			err:    PermissionDeniedError("nago.user.find_by_id"),
			wantID: "nago.user.find_by_id",
			wantOk: true,
		},
		{
			// Documented limitation: the discrimination is syntactic, so a single lower case
			// word is indistinguishable from a permission id. permission.RequiredOf resolves
			// the ambiguity through the registry, see TestRequiredOfUnregisteredPermission.
			name:   "single lower case word is treated as a candidate id",
			err:    PermissionDeniedError("findstaging"),
			wantID: "findstaging",
			wantOk: true,
		},
		{
			// application/flow/uc_find_form_share.go:42
			name:   "free form reason is not a permission",
			err:    PermissionDeniedError("workspace or share forbids access"),
			wantOk: false,
		},
		{
			name:   "free form reason with capitals",
			err:    PermissionDeniedError("Workspace forbids access"),
			wantOk: false,
		},
		{
			name:   "empty",
			err:    PermissionDeniedError(""),
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := tt.err.RequiredPermission()
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}

			if id != tt.wantID {
				t.Errorf("id = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestPermissionDeniedErrorSatisfiesInterfaces(t *testing.T) {
	var err error = PermissionDeniedError("nago.user.find_by_id")

	if !permission.IsDenied(err) {
		t.Error("must be recognized as a denial")
	}

	var carrier permission.RequiredPermission
	if !errors.As(err, &carrier) {
		t.Fatal("must satisfy permission.RequiredPermission")
	}
}

func TestDenialThroughWrapping(t *testing.T) {
	// this is the dominant pattern across application/drive/*
	wrapped := fmt.Errorf("cannot read file: %w", PermissionDeniedErr)

	if !permission.IsDenied(wrapped) {
		t.Error("wrapped denial must still be recognized")
	}

	if _, ok := permission.RequiredOf(wrapped); ok {
		t.Error("the bare sentinel must not name a permission")
	}
}

func TestInvalidSubjectErrorIsDenial(t *testing.T) {
	// regression: alert.makeMessageFromError used a type assertion on PermissionDeniedError
	// and therefore rendered "?" for this first party error.
	var err error = InvalidSubjectErr

	if !permission.IsDenied(err) {
		t.Error("InvalidSubjectError must be recognized as a denial")
	}

	if _, ok := permission.RequiredOf(err); ok {
		t.Error("InvalidSubjectError names no permission")
	}
}

func TestRequiredOfUnregisteredPermission(t *testing.T) {
	// an id which is syntactically valid but not registered must not leak into prose
	err := PermissionDeniedError("some.module.not_installed")

	if _, ok := permission.RequiredOf(err); ok {
		t.Error("unregistered permission must report false so callers use a generic message")
	}
}
