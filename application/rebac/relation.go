// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import "slices"

// Relation defines the relation between two entities and should be one of the constants below, however,
// it can be entirely arbitrary.
type Relation string

const (
	Owner   Relation = "owner"
	Writer  Relation = "writer"
	Deleter Relation = "deleter"
	// Member is has-member semantics.
	Member Relation = "member"
	Viewer Relation = "viewer"
	Parent Relation = "parent"
)

var AllRelations = slices.Values([]Relation{Owner, Writer, Deleter, Member, Viewer, Parent})
