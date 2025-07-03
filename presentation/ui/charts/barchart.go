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
	Y       float64
	Markers []TBarChartMarker
}

type TBarChartSeries struct {
	Label      string
	DataPoints []TBarChartDataPoint
}

type TBarChart struct {
	labels        []string
	series        []TBarChartSeries
	colors        []ui.Color
	frame         ui.Frame
	horizontal    bool
	stacked       bool
	downloadable  bool
	noDataMessage string
}

func BarChart() TBarChart {
	return TBarChart{
		horizontal:   false,
		stacked:      false,
		downloadable: true,
	}
}

func (c TBarChart) Labels(labels []string) TBarChart {
	c.labels = labels
	return c
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

func (c TBarChart) Horizontal(horizontal bool) TBarChart {
	c.horizontal = horizontal
	return c
}

func (c TBarChart) Stacked(stacked bool) TBarChart {
	c.stacked = stacked
	return c
}

func (c TBarChart) Downloadable(downloadable bool) TBarChart {
	c.downloadable = downloadable
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
	labels := make([]proto.Str, len(c.labels))
	for i, label := range c.labels {
		labels[i] = proto.Str(label)
	}

	return &proto.BarChart{
		Labels:        labels,
		Series:        c.getSeriesAsProtoSeries(),
		Colors:        protoColors,
		Frame:         c.getFrameAsProtoFrame(),
		Horizontal:    proto.Bool(c.horizontal),
		Stacked:       proto.Bool(c.stacked),
		Downloadable:  proto.Bool(c.downloadable),
		NoDataMessage: proto.Str(c.noDataMessage),
	}
}

func (c TBarChart) getSeriesAsProtoSeries() []proto.BarChartSeries {
	protoSeries := make([]proto.BarChartSeries, len(c.series))
	for i, series := range c.series {
		protoDataPoints := make([]proto.BarChartDataPoint, len(series.DataPoints))
		for j, dataPoint := range series.DataPoints {
			var protoMarkers []proto.BarChartMarker

			if !c.stacked {
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
				Y:       proto.Float(dataPoint.Y),
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
