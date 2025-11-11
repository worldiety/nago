// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package file

import (
	"io"
	"path/filepath"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type Type string

func (t Type) Ext() string {
	switch t {
	case PNG:
		return ".png"
	case JPEG:
		return ".jpg"
	case GIF:
		return ".gif"
	case PDF:
		return ".pdf"
	case DOCX:
		return ".docx"
	case PPTX:
		return ".pptx"
	case XLSX:
		return ".xlsx"
	default:
		return ""
	}
}

const (
	PNG    Type = "image/png"
	JPEG   Type = "image/jpeg"
	GIF    Type = "image/gif"
	PDF    Type = "application/pdf"
	DOCX   Type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	PPTX   Type = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	XLSX   Type = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	Binary Type = "application/octet-stream"
)

type CreateOptions struct {
	Name string
	Open func() (io.ReadCloser, error)
}

type File struct {
	ID        ID                     `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	MimeType  Type                   `json:"mimeType,omitempty"`
	CreatedBy user.ID                `json:"createdBy,omitempty"`
	CreatedAt xtime.UnixMilliseconds `json:"createdAt,omitempty"`
}

func (f File) Identity() ID {
	return f.ID
}

func (f File) FilenameWithExt() string {
	if filepath.Ext(f.Name) != "" {
		return f.Name
	}

	return f.Name + f.MimeType.Ext()
}

type Repository data.Repository[File, ID]
