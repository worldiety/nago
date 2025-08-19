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

// ora converts ChartSeriesType into its proto.ChartSeriesType equivalent.
func (cst ChartSeriesType) ora() proto.ChartSeriesType {
	return proto.ChartSeriesType(cst)
}

// Supported chart series types.
const (
	ChartSeriesTypeLine   = ChartSeriesType(proto.ChartSeriesTypeLine)
	ChartSeriesTypeColumn = ChartSeriesType(proto.ChartSeriesTypeColumn)
	ChartSeriesTypeArea   = ChartSeriesType(proto.ChartSeriesTypeArea)
)

// Series represents a labeled set of data points for a chart, with a specific series type.
type Series struct {
	Label      string
	Type       ChartSeriesType
	DataPoints []DataPoint
}

// Ora converts the Series into its proto.ChartSeries equivalent.
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
