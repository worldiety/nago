// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/barchart"
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/linechart"
)

// The conversions of batch c2, see native.go.

func (d ChartDataPoint) nago() chart.DataPoint {
	return chart.DataPoint{X: d.X, Y: d.Y}
}

func c2ChartDataPointFrom(d chart.DataPoint) ChartDataPoint {
	return ChartDataPoint{X: d.X, Y: d.Y}
}

func (m BarChartMarker) nago() barchart.Marker {
	return barchart.Marker{
		Label:          m.Label,
		SeriesIndex:    int(m.SeriesIndex),
		DataPointIndex: int(m.DataPointIndex),
		Value:          m.Value,
		Width:          int(m.Width),
		Height:         int(m.Height),
		Color:          nui.Color(m.Color),
		Round:          m.Round,
		Dashed:         m.Dashed,
	}
}

func c2BarChartMarkerFrom(m barchart.Marker) BarChartMarker {
	return BarChartMarker{
		Label:          m.Label,
		SeriesIndex:    int32(m.SeriesIndex),
		DataPointIndex: int32(m.DataPointIndex),
		Value:          m.Value,
		Width:          int32(m.Width),
		Height:         int32(m.Height),
		Color:          Color(m.Color),
		Round:          m.Round,
		Dashed:         m.Dashed,
	}
}

// c2BarChartMarkers returns the markers of nago, which the bar chart holds.
func c2BarChartMarkers(markers []BarChartMarker) []barchart.Marker {
	res := make([]barchart.Marker, 0, len(markers))
	for _, m := range markers {
		res = append(res, m.nago())
	}
	return res
}

func (m LineChartMarkers) nago() linechart.Markers {
	return linechart.Markers{Size: int(m.Size), BorderColor: nui.Color(m.BorderColor), ShowNullDataPoints: m.ShowNullDataPoints}
}

func c2LineChartMarkersFrom(m linechart.Markers) LineChartMarkers {
	return LineChartMarkers{Size: int32(m.Size), BorderColor: Color(m.BorderColor), ShowNullDataPoints: m.ShowNullDataPoints}
}
