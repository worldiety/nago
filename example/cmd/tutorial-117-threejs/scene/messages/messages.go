// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package messages declares the messages between the scene in the browser and its backend. Both import it, so they
// can not disagree about a message.
package messages

// Clicked tells the backend, that the user has clicked into the scene.
type Clicked struct {
	X, Y float64
}

// Color sets the color of the knot.
type Color struct {
	RGB uint32
}

// Speed sets the speed of the rotation.
type Speed struct {
	Factor float64
}
