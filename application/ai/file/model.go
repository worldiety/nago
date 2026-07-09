// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package file

import (
	"errors"
	"io"
	"path/filepath"
	"strings"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

// ErrNotDownloadable indicates that a file cannot be downloaded through the provider. Some providers only
// allow downloading files they generated themselves (e.g. Anthropic permits downloading API/tool-generated
// files but rejects user-uploaded ones). Callers can check for this with errors.Is.
var ErrNotDownloadable = errors.New("file is not downloadable")

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
	case Text:
		return ".txt"
	case Markdown:
		return ".md"
	case CSV:
		return ".csv"
	case JSON:
		return ".json"
	case XML:
		return ".xml"
	default:
		return ""
	}
}

const (
	PNG      Type = "image/png"
	JPEG     Type = "image/jpeg"
	GIF      Type = "image/gif"
	PDF      Type = "application/pdf"
	DOCX     Type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	PPTX     Type = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	XLSX     Type = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	Text     Type = "text/plain"
	Markdown Type = "text/markdown"
	CSV      Type = "text/csv"
	JSON     Type = "application/json"
	XML      Type = "application/xml"
	Binary   Type = "application/octet-stream"
)

// IsText reports whether the given media type carries human-readable text that can be injected inline into a
// conversation (as a text block) instead of being uploaded and referenced as a binary attachment. This covers
// every "text/*" type as well as common structured-text types such as application/json, application/xml and
// their structured suffixes (e.g. application/vnd.api+json). Binary document formats (PDF, DOCX, images, ...)
// are not considered text.
func IsText(t Type) bool {
	s := strings.ToLower(strings.TrimSpace(string(t)))
	if s == "" {
		return false
	}

	if strings.HasPrefix(s, "text/") {
		return true
	}

	switch s {
	case string(JSON), string(XML), "application/xhtml+xml", "application/javascript", "application/x-yaml", "application/yaml", "application/toml", "application/x-sh":
		return true
	}

	// structured syntax suffixes, e.g. application/vnd.api+json, image/svg+xml
	if strings.HasSuffix(s, "+json") || strings.HasSuffix(s, "+xml") {
		return true
	}

	return false
}

// Purpose declares the intended use of an uploaded file. Some providers require it on upload (e.g. OpenAI's
// Files API mandates a purpose such as "vision" or "user_data"), while others treat it as an optional scope
// hint (e.g. Anthropic). The zero value is left to the provider, which should pick a sensible default.
type Purpose string

const (
	// PurposeUserData is a generic, flexible purpose for files that are referenced from message content.
	// It maps to OpenAI's "user_data" and is a safe default for the Anthropic Files API (no scope).
	PurposeUserData Purpose = "user_data"

	// PurposeVision marks images used for vision. It maps to OpenAI's "vision" purpose.
	PurposeVision Purpose = "vision"
)

type CreateOptions struct {
	Name string
	Open func() (io.ReadCloser, error)

	// MimeType is the media type of the uploaded content. Optional; providers may infer it from the
	// filename when empty.
	MimeType Type

	// Purpose declares the intended use of the file. Optional; when empty the provider chooses a default
	// (see [Purpose]).
	Purpose Purpose
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
