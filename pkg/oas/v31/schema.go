// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

type Schema struct {
	GoPkgName string `json:"-"`
	GoName    string `json:"-"`

	Description          string             `json:"description,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Minimum              float64            `json:"minimum,omitempty"`
	Maximum              float64            `json:"maximum,omitempty"`
	Discriminator        string             `json:"discriminator,omitempty"`
	Example              any                `json:"example,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
}

func (s *Schema) RefName() string {
	if s.Ref != "" {
		return s.Ref
	}
	return "#/components/schemas/" + s.RefPlainName()
}

func (s *Schema) RefPlainName() string {
	return s.GoName
}
