// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chart

import "go.wdy.de/nago/presentation/proto"

// DataPoint represents a single chart data entry with an X label and a Y value.
type DataPoint struct {
	X string
	Y float64
}

// GetDataPointAsProtoDataPoint converts the DataPoint into its proto.ChartDataPoint equivalent.
func (dp DataPoint) GetDataPointAsProtoDataPoint() proto.ChartDataPoint {
	return proto.ChartDataPoint{
		X: proto.Str(dp.X),
		Y: proto.Float(dp.Y),
	}
}
