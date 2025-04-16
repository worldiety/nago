// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

type Path = string

type Paths map[Path]*PathItem

type PathItem struct {
	// Allows for a referenced definition of this path item. The value MUST be in the form of a URL, and the referenced
	//structure MUST be in the form of a Path Item Object. In case a Path Item Object field appears both in
	//the defined object and the referenced object, the behavior is undefined. See the rules for resolving Relative References.
	Ref string `json:"$ref,omitempty"`

	// An optional string summary, intended to apply to all operations in this path.
	Summary string `json:"summary,omitempty"`

	// An optional string description, intended to apply to all operations in this path.
	// [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	Get     *Operation `json:"get,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`

	Parameters []Parameter            `json:"parameters,omitempty"`
	Security   []*SecurityRequirement `json:"security,omitempty"`
}
