// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uindbinspector

import (
	"fmt"

	ndbinspector "go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"

	"github.com/worldiety/i18n"
)

// timeseriesStringWindow renders a string/enum column as a windowed value table
// with a dataview-style pager footer. It fetches enough rows for at most maxPages
// pages; the footer pages within the fetched window and, at the boundaries,
// advances the time cursor to the next/previous window.
func timeseriesStringWindow(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, ci ndbinspector.ColumnInfo, rangeMin, rangeMax *core.State[int64], invalidate *core.State[int]) core.View {
	cursor := rangeMin // reuse rangeMin as the moving "from" cursor
	from := cursor.Get()
	if from < ci.MinMillis {
		from = ci.MinMillis
	}

	pageSize := pageSizeFor(wnd)
	limit := pageSize * maxPages

	rows, err := uc.StringWindow(wnd.Subject(), ndbinspector.StringWindowRequest{
		Instance: instancePath, Engine: engine, Bucket: ci.Bucket, Column: ci.Column,
		MinMillis: from, MaxMillis: ci.MaxMillis, Limit: limit,
	})
	if err != nil {
		return alert.BannerError(err)
	}

	showPreview := wnd.Info().SizeClass > core.SizeClassSmall

	controls := ui.HStack(
		ui.IntField(StrFromMsShort.Get(wnd), cursor.Get(), cursor).Frame(ui.Frame{Width: ui.L200}),
		ui.Spacer(),
		flushKnife(wnd, uc, instancePath, engine, ci, invalidate),
	).FullWidth()

	table := stringTable(wnd, ci, rows, pageSize, cursor, showPreview)

	return ui.VStack(controls, ui.Space(ui.L16), table).FullWidth().Alignment(ui.Leading)
}

func stringTable(wnd core.Window, ci ndbinspector.ColumnInfo, rows []ndbinspector.StringRow, pageSize int, cursor *core.State[int64], showPreview bool) core.View {
	cols := []ui.TTableColumn{
		ui.TableColumn(ui.Text(StrChartTime.Get(wnd))),
		ui.TableColumn(ui.Text(StrColMillis.Get(wnd))),
	}
	if showPreview {
		cols = append(cols, ui.TableColumn(ui.Text(StrColValue.Get(wnd))))
	}

	pageIdx := core.StateOf[int](wnd, "ndbts-strpage")
	pageCount := (len(rows) + pageSize - 1) / pageSize
	clampPage(pageIdx, pageCount)
	if pageCount < 1 {
		pageCount = 1
	}
	start := pageIdx.Get() * pageSize
	end := min(start+pageSize, len(rows))
	visible := rows[start:end]

	tableRows := make([]ui.TTableRow, 0, len(visible)+1)
	for _, r := range visible {
		cells := []ui.TTableCell{
			ui.TableCell(ui.Text(fmtMillis(r.Millis))),
			ui.TableCell(ui.Text(fmt.Sprintf("%d", r.Millis))),
		}
		if showPreview {
			cells = append(cells, ui.TableCell(ui.Text(xstrings.EllipsisEnd(r.Value, previewMaxLen))))
		}
		tableRows = append(tableRows, ui.TableRow(cells...))
	}
	tableRows = append(tableRows, stringPagerFooter(wnd, len(cols), rows, ci, pageSize, pageIdx, pageCount, cursor))

	return ui.Table(cols...).Rows(tableRows...).Frame(ui.Frame{}.FullWidth())
}

func stringPagerFooter(wnd core.Window, colCount int, rows []ndbinspector.StringRow, ci ndbinspector.ColumnInfo, pageSize int, pageIdx *core.State[int], pageCount int, cursor *core.State[int64]) ui.TTableRow {
	var label string
	if len(rows) == 0 {
		label = StrNoValuesInWindow.Get(wnd)
	} else {
		start := pageIdx.Get() * pageSize
		end := min(start+pageSize, len(rows))
		label = StrTsPageLabel.Get(wnd,
			i18n.String("from", fmtMillis(rows[start].Millis)),
			i18n.String("to", fmtMillis(rows[end-1].Millis)),
			i18n.Int("page", pageIdx.Get()+1),
			i18n.Int("pages", pageCount),
		)
	}

	windowFull := len(rows) == pageSize*maxPages
	atStart := cursor.Get() <= ci.MinMillis

	onPrev := func() {
		// there is no cheap "previous window" for a time cursor without a scan;
		// jump back to the column start so the user is never stuck.
		cursor.Set(ci.MinMillis)
	}
	onNext := func() {
		if len(rows) > 0 {
			cursor.Set(rows[len(rows)-1].Millis + 1)
		}
	}

	return footerPager(colCount, label, len(rows) == 0, pageIdx, pageCount, atStart, windowFull, onPrev, onNext)
}
