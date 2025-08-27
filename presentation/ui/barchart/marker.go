// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package barchart

import (
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// Marker represents an annotation for a specific data point within a bar chart.
// It can highlight or label values by attaching a text label, size, color, and
// optional styles (e.g., rounded or dashed). The marker is associated with a
// specific series and data point index.
type Marker struct {
	Label          string
	SeriesIndex    int
	DataPointIndex int
	Value          string
	Width          int
	Height         int
	Color          ui.Color
	Round          bool
	Dashed         bool
}

// Ora converts the Marker into its proto.BarChartMarker representation,
// mapping all configuration fields (position, style, and label) into the
// protocol buffer format used for rendering.
func (m Marker) Ora() proto.BarChartMarker {
	return proto.BarChartMarker{
		Label:          proto.Str(m.Label),
		SeriesIndex:    proto.Int(m.SeriesIndex),
		DataPointIndex: proto.Int(m.DataPointIndex),
		Value:          proto.Str(m.Value),
		Width:          proto.Int(m.Width),
		Height:         proto.Int(m.Height),
		Color:          proto.Color(m.Color),
		Round:          proto.Bool(m.Round),
		Dashed:         proto.Bool(m.Dashed),
	}
}
