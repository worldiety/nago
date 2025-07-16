// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package inspector

import (
	"context"
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

type Store struct {
	Name       string
	Store      blob.Store
	Stereotype backup.Stereotype
	Error      error
}
type FindAll func(subject auth.Subject) ([]Store, error)

type PageResult struct {
	Entries  []PageEntry
	PageNo   int
	PageSize int
	Pages    int
	Count    int
}

type PageEntry struct {
	Key      string
	Data     []byte
	MimeType string
	Error    error
}

type FilterOptions struct {
	PageNo          int
	PageSize        int // If 0, defaults to 20
	OnlyKeys        bool
	MaxDataSize     int // If 0, defaults to 64KiB
	DetectMimeTypes bool
	Context         context.Context
}
type Filter func(subject auth.Subject, store blob.Store, opts FilterOptions) (PageResult, error)

type UseCases struct {
	FindAll FindAll
	Filter  Filter
}

func NewUseCases(p blob.Stores) UseCases {

	return UseCases{
		FindAll: NewFindAll(p),
		Filter:  NewFilter(),
	}
}
