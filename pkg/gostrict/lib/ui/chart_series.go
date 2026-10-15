// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// ChartSeriesType tells, how a series of a chart is drawn, like chart.ChartSeriesType of nago.
type ChartSeriesType int32

const (
	ChartSeriesTypeLine   ChartSeriesType = 0
	ChartSeriesTypeColumn ChartSeriesType = 1
	ChartSeriesTypeArea   ChartSeriesType = 2
)
