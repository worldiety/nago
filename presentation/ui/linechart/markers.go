// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package linechart

import (
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// Markers defines point markers for a line chart.
type Markers struct {
	Size               int      // marker size in px
	BorderColor        ui.Color // marker border color
	ShowNullDataPoints bool     // show markers for null values
}

// Ora converts Markers to proto.LineChartMarkers.
func (s Markers) Ora() proto.LineChartMarkers {
	return proto.LineChartMarkers{
		Size:               proto.Int(s.Size),
		BorderColor:        proto.Color(s.BorderColor),
		ShowNullDataPoints: proto.Bool(s.ShowNullDataPoints),
	}
}
