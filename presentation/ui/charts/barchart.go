// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package charts

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
)

type TBarChartMarker struct {
	Label    string
	Value    string
	Width    int
	Height   int
	Color    ui.Color
	IsRound  bool
	IsDashed bool
}

type TBarChartDataPoint struct {
	X       string
	Y       string
	Markers []TBarChartMarker
}

type TBarChartSeries struct {
	Label      string
	DataPoints []TBarChartDataPoint
}

type TBarChart struct {
	series        []TBarChartSeries
	colors        []ui.Color
	frame         ui.Frame
	isHorizontal  bool
	isStacked     bool
	noDataMessage string
}

func BarChart() TBarChart {
	return TBarChart{
		isHorizontal: false,
		isStacked:    false,
	}
}

func (c TBarChart) Series(series []TBarChartSeries) TBarChart {
	c.series = series
	return c
}

func (c TBarChart) Colors(colors []ui.Color) TBarChart {
	c.colors = colors
	return c
}

func (c TBarChart) Frame(frame ui.Frame) TBarChart {
	c.frame = frame
	return c
}

func (c TBarChart) IsHorizontal(isHorizontal bool) TBarChart {
	c.isHorizontal = isHorizontal
	return c
}

func (c TBarChart) IsStacked(isStacked bool) TBarChart {
	c.isStacked = isStacked
	return c
}

func (c TBarChart) NoDataMessage(noDataMessage string) TBarChart {
	c.noDataMessage = noDataMessage
	return c
}

func (c TBarChart) Render(ctx core.RenderContext) core.RenderNode {
	protoColors := make([]proto.Color, len(c.colors))
	for i, color := range c.colors {
		protoColors[i] = proto.Color(color)
	}
	return &proto.BarChart{
		Series:        c.getSeriesAsProtoSeries(),
		Colors:        protoColors,
		Frame:         c.getFrameAsProtoFrame(),
		IsHorizontal:  proto.Bool(c.isHorizontal),
		IsStacked:     proto.Bool(c.isStacked),
		NoDataMessage: proto.Str(c.noDataMessage),
	}
}

func (c TBarChart) getSeriesAsProtoSeries() []proto.BarChartSeries {
	protoSeries := make([]proto.BarChartSeries, len(c.series))
	for i, series := range c.series {
		protoDataPoints := make([]proto.BarChartDataPoint, len(series.DataPoints))
		for j, dataPoint := range series.DataPoints {
			var protoMarkers []proto.BarChartMarker

			if !c.isStacked {
				protoMarkers = make([]proto.BarChartMarker, len(dataPoint.Markers))
				for k, marker := range dataPoint.Markers {
					protoMarkers[k] = proto.BarChartMarker{
						Label:    proto.Str(marker.Label),
						Value:    proto.Str(marker.Value),
						Width:    proto.Int(marker.Width),
						Height:   proto.Int(marker.Height),
						Color:    proto.Color(marker.Color),
						IsRound:  proto.Bool(marker.IsRound),
						IsDashed: proto.Bool(marker.IsDashed),
					}
				}
			} else if len(dataPoint.Markers) > 0 {
				slog.Warn(fmt.Sprintf("BarChart: markers are not supported for stacked charts"))
			}

			protoDataPoints[j] = proto.BarChartDataPoint{
				X:       proto.Str(dataPoint.X),
				Y:       proto.Str(dataPoint.Y),
				Markers: protoMarkers,
			}
		}

		protoSeries[i] = proto.BarChartSeries{
			Label:      proto.Str(series.Label),
			DataPoints: protoDataPoints,
		}
	}

	return protoSeries
}

func (c TBarChart) getFrameAsProtoFrame() proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(c.frame.MinWidth),
		MaxWidth:  proto.Length(c.frame.MaxWidth),
		MinHeight: proto.Length(c.frame.MinHeight),
		MaxHeight: proto.Length(c.frame.MaxHeight),
		Width:     proto.Length(c.frame.Width),
		Height:    proto.Length(c.frame.Height),
	}
}
