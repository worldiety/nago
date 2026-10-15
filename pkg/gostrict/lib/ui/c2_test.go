// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/linechart"
	"go.wdy.de/nago/presentation/ui/progress"
)

// TestEnumsC2 compares the constants of the enums of batch c2 with the constants of nago.
func TestEnumsC2(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"ChartRoundingNone", int64(ChartRoundingNone), int64(chart.RoundingNone)},
		{"ChartRoundingRound", int64(ChartRoundingRound), int64(chart.RoundingRound)},
		{"ChartRoundingFloor", int64(ChartRoundingFloor), int64(chart.RoundingFloor)},
		{"ChartRoundingCeiling", int64(ChartRoundingCeiling), int64(chart.RoundingCeiling)},
		{"ChartSeriesTypeLine", int64(ChartSeriesTypeLine), int64(chart.ChartSeriesTypeLine)},
		{"ChartSeriesTypeColumn", int64(ChartSeriesTypeColumn), int64(chart.ChartSeriesTypeColumn)},
		{"ChartSeriesTypeArea", int64(ChartSeriesTypeArea), int64(chart.ChartSeriesTypeArea)},
		{"LineChartCurveStraight", int64(LineChartCurveStraight), int64(linechart.CurveStraight)},
		{"LineChartCurveSmooth", int64(LineChartCurveSmooth), int64(linechart.CurveSmooth)},
		{"LineChartCurveStepline", int64(LineChartCurveStepline), int64(linechart.CurveStepline)},
		{"ProgressHorizontal", int64(ProgressHorizontal), int64(progress.Horizontal)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}
