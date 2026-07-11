// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uindbinspector

import (
	"time"

	ndbinspector "go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/linechart"

	"github.com/worldiety/i18n"
)

// timeseriesChart renders a numeric (decimal) column as an M4-downsampled line
// chart over a selectable time range. The M4 width is derived from the window so
// the drawn series is bounded (at most 4*width points) regardless of how many
// raw points the range contains.
func timeseriesChart(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, ci ndbinspector.ColumnInfo, rangeMin, rangeMax *core.State[int64]) core.View {
	minMs, maxMs := rangeMin.Get(), rangeMax.Get()
	if maxMs < minMs {
		minMs, maxMs = maxMs, minMs
	}

	width := int(wnd.Info().Width) / 4
	if width < 50 {
		width = 50
	}
	if width > 800 {
		width = 800
	}

	pts, err := uc.SeriesM4(wnd.Subject(), ndbinspector.SeriesRequest{
		Instance: instancePath, Engine: engine, Bucket: ci.Bucket, Column: ci.Column,
		MinMillis: minMs, MaxMillis: maxMs, Width: width,
	})
	if err != nil {
		return alert.BannerError(err)
	}

	span := maxMs - minMs
	dps := make([]chart.DataPoint, 0, len(pts))
	for _, p := range pts {
		dps = append(dps, chart.DataPoint{
			X: fmtAxisMillis(p.Millis, span),
			Y: p.Value,
		})
	}

	c := chart.Chart{
		Frame:         ui.Frame{Height: ui.L400}.FullWidth(),
		XAxisTitle:    StrChartTime.Get(wnd),
		YAxisTitle:    StrChartValue.Get(wnd),
		NoDataMessage: StrChartNoData.Get(wnd),
	}
	series := []chart.Series{{
		Label:      ci.Column,
		Type:       chart.ChartSeriesTypeLine,
		DataPoints: dps,
	}}

	caption := StrChartCaption.Get(wnd,
		i18n.String("from", fmtMillis(minMs)),
		i18n.String("to", fmtMillis(maxMs)),
		i18n.Int("buckets", width),
		i18n.Int("points", len(dps)),
	)

	return ui.VStack(
		rangeControls(wnd, ci, rangeMin, rangeMax),
		ui.Space(ui.L16),
		ui.Text(caption).Font(ui.Small),
		ui.Space(ui.L8),
		linechart.LineChart(c).Curve(linechart.CurveSmooth).Series(series),
	).Alignment(ui.Leading).Frame(ui.Frame{Width: ui.Full, Height: ui.L560})
}

// rangeControls offers quick ranges relative to the column's max plus explicit
// min/max millisecond fields. No extra state is introduced; the range lives in
// the two int64 states owned by the page.
func rangeControls(wnd core.Window, ci ndbinspector.ColumnInfo, rangeMin, rangeMax *core.State[int64]) core.View {
	setRange := func(from, to int64) {
		if from < ci.MinMillis {
			from = ci.MinMillis
		}
		rangeMin.Set(from)
		rangeMax.Set(to)
	}
	quick := func(title string, dur time.Duration) core.View {
		return ui.SecondaryButton(func() {
			setRange(ci.MaxMillis-int64(dur/time.Millisecond), ci.MaxMillis)
		}).Title(title)
	}

	return ui.VStack(
		ui.HStack(
			ui.SecondaryButton(func() { setRange(ci.MinMillis, ci.MaxMillis) }).Title(StrRangeTotal.Get(wnd)),
			quick("1 h", time.Hour),
			quick("24 h", 24*time.Hour),
			quick("7 d", 7*24*time.Hour),
		).Gap(ui.L8).FullWidth(),
		ui.Space(ui.L8),
		ui.HStack(
			ui.IntField(StrFromMs.Get(wnd), rangeMin.Get(), rangeMin).Frame(ui.Frame{Width: ui.L200}),
			ui.IntField(StrToMs.Get(wnd), rangeMax.Get(), rangeMax).Frame(ui.Frame{Width: ui.L200}),
		).Gap(ui.L8).FullWidth().Alignment(ui.Bottom),
	).FullWidth().Alignment(ui.Leading)
}
