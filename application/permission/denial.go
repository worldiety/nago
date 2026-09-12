// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package permission

import (
	"errors"

	"github.com/worldiety/i18n"
)

// Denial is implemented by errors which express that an operation was refused for
// authorization reasons. Implementing it is enough to be classified as a permission problem
// by the presentation layer.
//
// Not every denial can name a permission: a tenant boundary or an ownership check refuses
// access without a permission ever being involved. Such errors implement only Denial and are
// reported with a generic message.
type Denial interface {
	error
	PermissionDenied() bool
}

// RequiredPermission is the optional companion of [Denial] for errors which can name the
// permission that would have been necessary. Implement it to let the presentation layer tell
// the user which permission a rights holder has to grant.
//
// Returning false means that no nameable permission exists, which is not an error condition.
type RequiredPermission interface {
	RequiredPermission() (ID, bool)
}

// IsDenied reports whether anywhere in the error tree an error expresses a permission denial.
func IsDenied(err error) bool {
	var denial Denial
	return errors.As(err, &denial) && denial.PermissionDenied()
}

// RequiredOf extracts the permission which would have been required to pass the denial found
// in the error tree, if any error in it can name one.
//
// Note that this deliberately uses errors.As rather than a type assertion, so that every
// implementation is honoured and not just one concrete type.
func RequiredOf(err error) (Permission, bool) {
	var carrier RequiredPermission
	if !errors.As(err, &carrier) {
		return Permission{}, false
	}

	id, ok := carrier.RequiredPermission()
	if !ok || id == "" {
		return Permission{}, false
	}

	if p, ok := Find(id); ok {
		return p, true
	}

	// The permission is not registered, for example because the declaring module is not
	// installed in this application. The raw id is not fit for prose, so report it as unknown
	// rather than leaking something like "nago.user.find_by_id" into a sentence.
	return Permission{}, false
}

// LocalizedName resolves the human readable name of the permission for the given bundle.
// Name may either be a plain text fallback or an i18n key.
//
// Returns the empty string if the permission has no name at all, so that callers can fall
// back to a generic phrasing instead of showing a technical identifier.
func (p Permission) LocalizedName(b i18n.Bundler) string {
	if p.Name == "" || b == nil {
		return p.Name
	}

	bundle := b.Bundle()
	if bundle == nil {
		return p.Name
	}

	return bundle.Resolve(p.Name)
}
