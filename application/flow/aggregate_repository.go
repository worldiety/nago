// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type Repository struct {
	parent     *Workspace
	id         RepositoryID
	structType *StructType
}

func (r *Repository) Identity() RepositoryID {
	return r.id
}

func (r *Repository) Type() *StructType {
	return r.structType
}

func (r *Repository) String() string {
	return string(r.id)
}
