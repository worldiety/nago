// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uindbinspector

import (
	"fmt"
	"log/slog"
	"time"

	ndbinspector "go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"

	"github.com/worldiety/i18n"
)

// PageTimeseries is the time-series inspector: pick an ndb database, a tsdb
// engine and a column, choose a time range and view the series as an
// M4-downsampled line chart (numeric columns) or a windowed value table
// (string/enum columns). Reads are bounded / downsampled, so it stays responsive
// over columns holding billions of points.
func PageTimeseries(wnd core.Window, uc ndbinspector.UseCases) core.View {
	startMs := time.Now()
	defer func() {
		stop := time.Now()
		slog.Info("PageTimeseries took", "duration", stop.Sub(startMs))
	}()
	if !wnd.Subject().HasPermission(ndbinspector.PermNDBInspector) {
		return alert.Banner(StrNoAccessTitle.Get(wnd), StrNoAccessBody.Get(wnd))
	}

	instances, err := uc.Instances(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}
	if len(instances) == 0 {
		return ui.VStack(
			header(StrTimeseriesTitle.Get(wnd)),
			alert.Banner(StrNoNdbTitle.Get(wnd), StrNoNdbBody.Get(wnd)),
		).FullWidth().Alignment(ui.Leading)
	}

	invalidate := core.AutoState[int](wnd)
	selInstance := core.AutoState[selKey](wnd)
	selEngine := core.AutoState[selKey](wnd)
	selColumn := core.AutoState[selKey](wnd)
	rangeMin := core.AutoState[int64](wnd)
	rangeMax := core.AutoState[int64](wnd)

	if selInstance.Get() == "" {
		selInstance.Set(selKey(instances[0].Path))
	}
	selInstance.Observe(func(selKey) {
		selEngine.Set("")
		selColumn.Set("")
		rangeMin.Set(0)
		rangeMax.Set(0)
	})
	instancePath := string(selInstance.Get())

	engines, err := uc.SeriesEngines(wnd.Subject(), instancePath)
	if err != nil {
		return alert.BannerError(err)
	}
	if selEngine.Get() == "" && len(engines) > 0 {
		selEngine.Set(selKey(engines[0].Name))
	}
	selEngine.Observe(func(selKey) {
		selColumn.Set("")
		rangeMin.Set(0)
		rangeMax.Set(0)
	})
	engine := string(selEngine.Get())

	instOpts := make([]dropdown.Option[selKey], 0, len(instances))
	for _, in := range instances {
		instOpts = append(instOpts, dropdown.Option[selKey]{Value: selKey(in.Path), Label: in.Name})
	}
	engOpts := make([]dropdown.Option[selKey], 0, len(engines))
	for _, e := range engines {
		engOpts = append(engOpts, dropdown.Option[selKey]{Value: selKey(e.Name), Label: fmt.Sprintf("%s (%s)", e.Name, e.Kind)})
	}

	var columns []ndbinspector.ColumnInfo
	if engine != "" {
		if columns, err = uc.Columns(wnd.Subject(), instancePath, engine); err != nil {
			return alert.BannerError(err)
		}
	}
	if selColumn.Get() == "" && len(columns) > 0 {
		selColumn.Set(selKey(columns[0].Key()))
	}
	selColumn.Observe(func(selKey) { rangeMin.Set(0); rangeMax.Set(0) })

	colOpts := make([]dropdown.Option[selKey], 0, len(columns))
	for _, ci := range columns {
		colOpts = append(colOpts, dropdown.Option[selKey]{
			Value: selKey(ci.Key()),
			Label: fmt.Sprintf("%s (%s)", ci.Key(), ci.Scheme),
		})
	}

	current, hasCurrent := findColumn(columns, string(selColumn.Get()))
	// default the range to the column's full data range on first view.
	if hasCurrent && current.HasData && rangeMin.Get() == 0 && rangeMax.Get() == 0 {
		rangeMin.Set(current.MinMillis)
		rangeMax.Set(current.MaxMillis)
	}

	var right core.View
	switch {
	case len(engines) == 0:
		right = alert.Banner(StrNoTsEngineTitle.Get(wnd), StrNoTsEngineBody.Get(wnd))
	case !hasCurrent:
		right = ui.Text(StrSelectColumnHint.Get(wnd))
	case !current.HasData:
		right = alert.Banner(StrNoDataTitle.Get(wnd), StrNoDataBody.Get(wnd, i18n.String("column", current.Key())))
	case current.Numeric():
		right = timeseriesChart(wnd, uc, instancePath, engine, current, rangeMin, rangeMax)
	default:
		right = timeseriesStringWindow(wnd, uc, instancePath, engine, current, rangeMin, rangeMax, invalidate)
	}

	return ui.VStack(
		header(StrTimeseriesTitle.Get(wnd)),
		ui.Space(ui.L16),
		ui.HStack(
			ui.VStack(
				dropdown.Dropdown[selKey](StrDatabase.Get(wnd), instOpts, selInstance.Get()).
					InputValue(selInstance).Frame(ui.Frame{}.FullWidth()),
				ui.Space(ui.L8),
				dropdown.Dropdown[selKey](StrEngine.Get(wnd), engOpts, selEngine.Get()).
					InputValue(selEngine).Frame(ui.Frame{}.FullWidth()),
				ui.Space(ui.L8),
				columnPicker(wnd, colOpts, selColumn),
				ui.Space(ui.L8),
				ui.ScrollView(columnStatList(wnd, uc, instancePath, engine, columns, selColumn, invalidate)).
					Axis(ui.ScrollViewAxisVertical),
			).Alignment(ui.Top).Frame(ui.Frame{Width: ui.L400, MaxWidth: ui.L400}),
			ui.VLine().Frame(ui.Frame{}),
			ui.VStack(right).Alignment(ui.Top).FullWidth(),
		).FullWidth().Alignment(ui.Stretch),
	).FullWidth().Alignment(ui.Leading)
}

func columnPicker(wnd core.Window, opts []dropdown.Option[selKey], selected *core.State[selKey]) core.View {
	if len(opts) == 0 {
		return ui.Text(StrNoColumns.Get(wnd)).Font(ui.Small)
	}
	return dropdown.Dropdown[selKey](StrColumn.Get(wnd), opts, selected.Get()).
		InputValue(selected).
		Frame(ui.Frame{}.FullWidth())
}

// columnStatList shows per-column metadata plus a per-column flush/delete-range
// knife tool. Selection happens through the column dropdown above.
func columnStatList(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, columns []ndbinspector.ColumnInfo, selected *core.State[selKey], invalidate *core.State[int]) core.View {
	if len(columns) == 0 {
		return nil
	}
	rows := make([]core.View, 0, len(columns))
	for _, ci := range columns {
		rows = append(rows, columnStatRow(wnd, uc, instancePath, engine, ci, invalidate))
	}
	return ui.VStack(rows...).FullWidth().Alignment(ui.Leading).Gap(ui.L8)
}

func columnStatRow(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, ci ndbinspector.ColumnInfo, invalidate *core.State[int]) core.View {
	span := StrEmptyRange.Get(wnd)
	if ci.HasData {
		span = StrRangeSpan.Get(wnd, i18n.String("from", fmtMillis(ci.MinMillis)), i18n.String("to", fmtMillis(ci.MaxMillis)))
	}
	// The exact point count needs a full scan; cache it per page under a fixed
	// id so the scan runs at most once per visit. Destructive actions call
	// resetCountCache to re-scan the affected counts.
	key := instancePath + "|" + engine + "|" + ci.Key()
	count := cachedCount(wnd, countCacheColumns, key, func() (int64, error) {
		return uc.CountColumn(wnd.Subject(), instancePath, engine, ci.Bucket, ci.Column)
	})

	return ui.HStack(
		ui.VStack(
			ui.Text(ci.Key()).Font(ui.BodyLarge),
			ui.Text(StrColStatRow.Get(wnd,
				i18n.String("scheme", ci.Scheme.String()),
				i18n.String("count", countLabel(count)),
				i18n.Int("chunks", ci.Chunks),
				i18n.String("size", xstrings.FormatByteSize(wnd.Locale(), ci.Bytes, 1)),
			)).Font(ui.Small),
			ui.Text(span).Font(ui.Small),
		).Alignment(ui.Leading),
		ui.Spacer(),
		columnKnife(wnd, uc, instancePath, engine, ci, invalidate),
	).Alignment(ui.Center).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(ui.Frame{}.FullWidth())
}

func findColumn(columns []ndbinspector.ColumnInfo, key string) (ndbinspector.ColumnInfo, bool) {
	for _, ci := range columns {
		if ci.Key() == key {
			return ci, true
		}
	}
	return ndbinspector.ColumnInfo{}, false
}
