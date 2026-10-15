// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// BarChartMarker is an annotation of a data point of a bar chart, which highlights or labels a value, like
// barchart.Marker of nago.
type BarChartMarker struct {
	Label          string
	SeriesIndex    int32
	DataPointIndex int32
	Value          string
	Width          int32
	Height         int32
	Color          Color
	Round          bool
	Dashed         bool
}
