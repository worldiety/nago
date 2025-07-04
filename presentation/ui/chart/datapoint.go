// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chart

import "go.wdy.de/nago/presentation/proto"

type DataPoint struct {
	X string
	Y float64
}

func (dp DataPoint) GetDataPointAsProtoDataPoint() proto.ChartDataPoint {
	return proto.ChartDataPoint{
		X: proto.Str(dp.X),
		Y: proto.Float(dp.Y),
	}
}
