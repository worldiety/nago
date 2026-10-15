// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/chart"
)

func (s IndexState) nago() *core.State[int] {
	return get[*core.State[int]](s.ref)
}

func (s ColorState) nago() *core.State[nui.Color] {
	return get[*core.State[nui.Color]](s.ref)
}

func (c Chart) nago(labels []string) chart.Chart {
	return chart.Chart{
		Labels:        labels,
		Frame:         c.Frame.nago(),
		Downloadable:  c.Downloadable,
		NoDataMessage: c.NoDataMessage,
		XAxisTitle:    c.XAxisTitle,
		YAxisTitle:    c.YAxisTitle,
		LabelRounding: chart.RoundingType(c.LabelRounding),
	}
}

func chartSeriesOf(label string, typ ChartSeriesType, points []ChartDataPoint) chart.Series {
	s := chart.Series{Label: label, Type: chart.ChartSeriesType(typ)}
	for _, p := range points {
		s.DataPoints = append(s.DataPoints, p.nago())
	}
	return s
}

func chartSeriesList(series []TChartSeries) []chart.Series {
	var res []chart.Series
	for _, s := range series {
		res = append(res, get[chart.Series](s.ref))
	}
	return res
}
