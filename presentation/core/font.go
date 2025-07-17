// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

type FontFace struct {
	// Family is an arbitrary name for a font family to match against different variants.
	Family string `json:"family,omitempty"`

	// Style is like normal, italic, oblique
	Style string `json:"style,omitempty"`

	// Weight property of the FontFace interface retrieves or sets the weight of the font. Like 400 or 700 or bold
	// or normal.
	Weight string `json:"weight,omitempty"`

	// Source where to download the font.
	Source URI `json:"URL,omitempty"`
}

type Fonts struct {
	DefaultFont string     `json:"default,omitempty"`
	Faces       []FontFace `json:"faces,omitempty"`
}

func (ff Fonts) Contains(family string) bool {
	for _, face := range ff.Faces {
		if face.Family == family {
			return true
		}
	}
	
	return false
}
