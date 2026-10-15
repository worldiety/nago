// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// LineChartCurve is the style of the curve, which connects the data points of a line chart, like linechart.Curve of
// nago.
type LineChartCurve int32

const (
	LineChartCurveStraight LineChartCurve = 0
	LineChartCurveSmooth   LineChartCurve = 1
	LineChartCurveStepline LineChartCurve = 2
)
