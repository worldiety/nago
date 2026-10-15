// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// ChartDataPoint is a single value of a chart with an X label and a Y value, like chart.DataPoint of nago.
type ChartDataPoint struct {
	X string
	Y float64
}
