// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type WorkspaceEvent interface {
	WorkspaceID() WorkspaceID
}

type TypeCreated struct {
	// Workspace
	Workspace WorkspaceID

	// Package for this type. Only for universe types are allowed to have empty package names.
	Package PackageID

	// Type at least uniquely identifies this type within the workspace, but probably it is also unique
	// in the whole system.
	Type TypeID

	// Name of the type. This follows the naming conventions of Go packages and types including their visiblity
	// rules. Note that build-in universe types have lowercase (unexported) names.
	Name Typename

	// Kind determines the basic memory layout to expect for this type.
	Kind Kind

	// BuildIn types must not be resolved in the context and need no further definition. For example, primitive
	// types like int, string, bool, etc. are build-in but also higher types like time.Time, json.RawMessage, user.ID
	// etc. may be valid.
	BuildIn bool
}

func (e TypeCreated) WorkspaceID() WorkspaceID {
	return e.Workspace
}
