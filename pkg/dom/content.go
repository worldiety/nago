// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package dom provides a fully type-safe in-memory HTML5 DOM API.
// Content category interfaces enforce permitted children at compile time.
package dom

import "io"

// Node is the base interface for all DOM nodes.
// The unexported node() method prevents external implementations and
// ensures only types defined in this package satisfy the interface.
type Node interface {
	node() // sentinel – prevents external implementations

	// Attribute access
	SetAttr(key, value string)
	GetAttr(key string) string
	RemoveAttr(key string)

	// Tree navigation
	Parent() Node

	// Text helpers
	SetTextContent(s string)
	TextContent() string

	// Raw HTML helpers (children are replaced on set)
	SetInnerHTML(raw string)
	InnerHTML() string

	// Rendering
	Render(w io.Writer) error

	// internal helpers used by element.go / render.go
	appendChildNode(n Node)
	children() []Node
	tagName() string
	isVoid() bool
}

// ── Content category marker interfaces ────────────────────────────────────────
// Each marker interface carries only the unexported sentinel so that only
// types in this package can satisfy them. AppendChild methods on container
// elements use these interfaces to enforce permitted-content rules at
// compile time.

// MetadataContent is content that sets up the presentation or behaviour of
// the rest of the content, or sets up the relationship of the document with
// other documents. Permitted inside <head>.
type MetadataContent interface {
	Node
	metadataContent()
}

// FlowContent is content that makes up the normal flow of a document.
// Most block-level and many inline elements are flow content.
type FlowContent interface {
	Node
	flowContent()
}

// SectioningContent defines the scope of headings and footers.
type SectioningContent interface {
	FlowContent
	sectioningContent()
}

// HeadingContent defines the header of a section.
type HeadingContent interface {
	FlowContent
	headingContent()
}

// PhrasingContent is the text of the document and the elements that mark
// up that text at the intra-paragraph level. It is a subset of flow content.
type PhrasingContent interface {
	FlowContent
	phrasingContent()
}

// EmbeddedContent imports another resource or inserts content from another
// vocabulary into the document.
type EmbeddedContent interface {
	PhrasingContent
	embeddedContent()
}

// InteractiveContent is content that is specifically intended for user
// interaction.
type InteractiveContent interface {
	PhrasingContent
	interactiveContent()
}

// ── Specialised child-constraint interfaces ────────────────────────────────────

// SelectContent is permitted directly inside <select>.
type SelectContent interface {
	Node
	selectContent()
}

// TableContent is permitted directly inside <table>.
type TableContent interface {
	Node
	tableContent()
}

// TableSectionContent is permitted directly inside <thead>, <tbody>, <tfoot>.
// Only <tr> satisfies this interface.
type TableSectionContent interface {
	Node
	tableSectionContent()
}

// TableRowContent is permitted directly inside <tr>.
// Only <th> and <td> satisfy this interface.
type TableRowContent interface {
	Node
	tableRowContent()
}

// ColGroupContent is permitted directly inside <colgroup>.
// Only <col> satisfies this interface.
type ColGroupContent interface {
	Node
	colGroupContent()
}

// ListContent is permitted directly inside <ul> and <ol>.
// Only <li> satisfies this interface.
type ListContent interface {
	Node
	listContent()
}

// DlContent is permitted directly inside <dl>.
// Only <dt> and <dd> satisfy this interface.
type DlContent interface {
	Node
	dlContent()
}

// RubyContent is permitted directly inside <ruby>.
type RubyContent interface {
	Node
	rubyContent()
}

// FigureContent is permitted directly inside <figure>.
// <figcaption> and any FlowContent satisfy this interface.
type FigureContent interface {
	Node
	figureContent()
}

// MediaContent is permitted directly inside <video> and <audio>
// before the fallback content: <source> and <track>.
type MediaContent interface {
	Node
	mediaContent()
}

// PictureContent is permitted directly inside <picture>: <source> and <img>.
type PictureContent interface {
	Node
	pictureContent()
}

// FieldsetContent is permitted directly inside <fieldset>:
// at most one <legend> followed by flow content.
type FieldsetContent interface {
	Node
	fieldsetContent()
}
