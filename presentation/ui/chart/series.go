// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chart

import (
	"go.wdy.de/nago/presentation/proto"
)

type ChartSeriesType int

func (cst ChartSeriesType) ora() proto.ChartSeriesType {
	return proto.ChartSeriesType(cst)
}

const (
	ChartSeriesTypeLine   = ChartSeriesType(proto.ChartSeriesTypeLine)
	ChartSeriesTypeColumn = ChartSeriesType(proto.ChartSeriesTypeColumn)
	ChartSeriesTypeArea   = ChartSeriesType(proto.ChartSeriesTypeArea)
)

type Series struct {
	Label      string
	Type       ChartSeriesType
	DataPoints []DataPoint
}

func (s Series) Ora() proto.ChartSeries {
	protoDataPoints := make([]proto.ChartDataPoint, len(s.DataPoints))

	for j, dataPoint := range s.DataPoints {
		protoDataPoints[j] = dataPoint.GetDataPointAsProtoDataPoint()
	}

	return proto.ChartSeries{
		Label:      proto.Str(s.Label),
		Type:       s.Type.ora(),
		DataPoints: protoDataPoints,
	}
}
