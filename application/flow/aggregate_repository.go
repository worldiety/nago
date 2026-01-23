// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type Repository struct {
	// ID is also human-readable and usually looks like my.namespace.myentity and follows the nago conventions.
	ID          RepositoryID
	StructType  TypeID
	Description string
}

func NewRepository(id RepositoryID, structType TypeID) *Repository {
	return &Repository{
		ID:         id,
		StructType: structType,
	}
}

func (r *Repository) Identity() RepositoryID {
	return r.ID
}

func (r *Repository) String() string {
	return string(r.ID)
}

func (r *Repository) Clone() *Repository {
	c := *r
	return &c
}
