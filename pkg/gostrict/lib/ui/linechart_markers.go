// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// LineChartMarkers defines the point markers of a line chart, like linechart.Markers of nago.
type LineChartMarkers struct {
	// Size is the size of a marker in pixels.
	Size int32
	// BorderColor is the color of the border of a marker.
	BorderColor Color
	// ShowNullDataPoints shows markers for null values.
	ShowNullDataPoints bool
}
