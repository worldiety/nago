package workspace

import (
	"go.wdy.de/nago/pkg/data"
)

// BlobID to safe the binary somewhere
type BlobID string

// Filename which is unique per Workspace
type Filename string

// ID of the workspace
type ID string

// File within a workspace
type File struct {
	Name     Filename
	Ref      BlobID
	Mimetype Mimetype
	Size     int64
}

type Mimetype string

const (
	MIMETypeGoSource = "text/x-go"
	MIMETypeMarkdown = "text/markdown"
	MIMETypeTypst    = "text/x-typst"
	MIMETypeLaTeX    = "application/x-latex"
	MIMETypeSVG      = "image/svg+xml"
	MIMETypePNG      = "image/png"
	MIMETypeJPG      = "image/jpeg"
)

// Type denotes the kind of workspace, which can be freely defined.
type Type string

type Workspace struct {
	ID               ID
	Type             Type
	Name             string
	Files            []File
	AllowedMimeTypes []string
}

func (t Workspace) Identity() ID {
	return t.ID
}

type Repository data.Repository[Workspace, ID]
