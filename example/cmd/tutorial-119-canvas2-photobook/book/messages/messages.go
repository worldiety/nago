// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package messages declares the messages between the photo book in the browser and its backend.
package messages

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Photo is a photo on the double page in millimeters: the center, the size and the rotation in radians. The
// photos are in the order of their layers, the last one is on top.
type Photo struct {
	URL      string
	X, Y     float64
	W, H     float64
	Rotation float64
}

// Hello asks the backend for the layout, when the editor starts.
type Hello struct{}

// Layout replaces the layout of the editor, e.g. after another window has changed it.
type Layout struct {
	Photos list.List[Photo]
}

// Save stores the layout, after the user has changed it.
type Save struct {
	Photos list.List[Photo]
}

// Add puts an uploaded image with its size in pixels on the double page.
type Add struct {
	URL           string
	Width, Height float64
}
