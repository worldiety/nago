// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"io"

	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type ProcessingStatus string
type Document struct {
	ID               ID
	CreatedAt        xtime.UnixMilliseconds
	Hash             string
	Library          library.ID
	MimeType         string
	Name             string
	ProcessingStatus ProcessingStatus
	Size             int64
	Summary          string
}

type CreateOptions struct {
	Filename string
	Reader   io.Reader
}
