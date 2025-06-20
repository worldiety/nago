// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package parser

import (
	"context"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/presentation/core"
	"io"
	"iter"
)

type ID string

type FromUpload struct {
	Enabled       bool
	MimeTypes     []string
	MaxUploadSize int64
}

// FromBuildIn importers accept the import of a nil reader, which means that they have their own data source
// implementation, e.g. by loading a secret or querying a URL.
type FromBuildIn struct {
	Enabled bool
}
type Configuration struct {
	Image       core.SVG
	Name        string
	Description string
	FromUpload  FromUpload
	FromBuildIn FromBuildIn
}

type Options struct {
}

type Parser interface {
	Identity() ID
	Configuration() Configuration
	Parse(ctx context.Context, reader io.Reader, opts Options) iter.Seq2[*jsonptr.Obj, error]
}
