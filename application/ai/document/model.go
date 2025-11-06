// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"errors"
	"io"

	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type ProcessingStatus string

const (
	ProcessingCompleted ProcessingStatus = "Completed"
	ProcessingRunning   ProcessingStatus = "Running"
)

var (
	UnsupportedFormatError = errors.New("unsupported document format")
)

type Document struct {
	ID               ID
	CreatedAt        xtime.UnixMilliseconds
	CreatedBy        user.ID
	Hash             string
	Library          library.ID
	MimeType         string
	Name             string
	ProcessingStatus ProcessingStatus
	Size             int64
	Summary          string
}

func (d Document) Identity() ID {
	return d.ID
}

type CreateOptions struct {
	Filename string
	Reader   io.Reader
}

type Repository data.Repository[Document, ID]
