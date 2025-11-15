// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

// A Point defines an arbitrary data point.
type Point[T, D Number] struct {
	X T `json:"x"`
	Y D `json:"y"`
}
