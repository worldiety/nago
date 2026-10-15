// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package geom declares the records, which the functions of the header package host take and return.
package geom

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Rect is a rectangle.
type Rect struct {
	X, Y, W, H float64
}

// Stats summarizes a list of numbers.
type Stats struct {
	Count int32
	Sum   float64
	Label string
	Parts list.List[Rect]
}
