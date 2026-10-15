// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

// Transformation moves, scales and rotates a view. See TStack.Transformation.
type Transformation struct {
	TranslateX Length
	TranslateY Length
	TranslateZ Length
	ScaleX     float64
	ScaleY     float64
	ScaleZ     float64
	// RotateZ is the rotation in degrees.
	RotateZ float64
}
