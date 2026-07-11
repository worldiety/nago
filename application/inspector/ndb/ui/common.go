// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uindbinspector

import (
	"fmt"
	"time"

	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

// selKey is a string alias satisfying the dropdown/picker ~string constraint.
type selKey string

// maxPages is the number of pages the in-window pager can navigate before the
// underlying cursor advances to the next window (limit = pageSize * maxPages).
const maxPages = 5

// Fixed page-state ids for the per-page count caches.
const (
	countCacheMessages = "ndbmsg-count"
	countCacheColumns  = "ndbts-count"
)

func header(title string) core.View {
	return ui.HStack(ui.Text(title).Font(ui.Title), ui.Spacer()).FullWidth()
}

// pageSizeFor computes how many rows fit into the current window height, using
// the same calculation as the blob store inspector. It never returns less than
// one and requires no state.
func pageSizeFor(wnd core.Window) int {
	const (
		lineHeight    = 69.0
		overheadLines = 4
	)
	return max(1, int((float64(wnd.Info().Height)-(overheadLines*lineHeight)-96)/lineHeight))
}

// footerPager renders a dataview-style table footer row (label + prev/next
// chevrons on a card-footer background) that pages within a fetched window and,
// on the boundaries, invokes onPrevWindow / onNextWindow to shift the underlying
// cursor. It is always rendered so navigation stays possible even for an empty
// window.
//
// windowFull must be true when the current fetch filled its whole limit (i.e. a
// further window may exist).
func footerPager(colCount int, label string, empty bool, pageIdx *core.State[int], pageCount int, atCursorStart, windowFull bool, onPrevWindow, onNextWindow func()) ui.TTableRow {
	canPrev := pageIdx.Get() > 0 || !atCursorStart
	canNext := pageIdx.Get() < pageCount-1 || windowFull

	prev := func() {
		if pageIdx.Get() > 0 {
			pageIdx.Set(pageIdx.Get() - 1)
			return
		}
		onPrevWindow()
		pageIdx.Set(maxPages - 1)
	}
	next := func() {
		if pageIdx.Get() < pageCount-1 {
			pageIdx.Set(pageIdx.Get() + 1)
			return
		}
		onNextWindow()
		pageIdx.Set(0)
	}

	return ui.TableRow(
		ui.TableCell(
			ui.HStack(
				ui.Text(label),
				ui.Spacer(),
				ui.TertiaryButton(prev).PreIcon(icons.ChevronLeft).Enabled(canPrev),
				ui.TertiaryButton(next).PreIcon(icons.ChevronRight).Enabled(canNext),
			).Gap(ui.L8).FullWidth(),
		).ColSpan(colCount),
	).BackgroundColor(ui.ColorCardFooter)
}

// clampPage clamps a page index state into [0, pageCount).
func clampPage(pageIdx *core.State[int], pageCount int) {
	if pageCount < 1 {
		pageCount = 1
	}
	if pageIdx.Get() >= pageCount {
		pageIdx.Set(pageCount - 1)
	}
	if pageIdx.Get() < 0 {
		pageIdx.Set(0)
	}
}

// countCache is a per-page cache of expensive entry counts, keyed by an entity
// key (e.g. a message type or a "bucket/column"). It lives in page state under a
// fixed id, so a count is scanned at most once per page visit and reused until
// the page (and thus the state) is left. Missing entries are filled lazily by
// cachedCount; resetCountCache drops the cache after a destructive action so the
// affected counts are re-scanned on the next render.
type countCache map[string]int64

// cachedCount returns the count for key, scanning via fetch and caching the
// result the first time. stateID must be a fixed, page-stable id so the cache
// survives re-renders but is discarded when the page is left. Errors are cached
// as -1 so a failing scan is not retried on every render.
//
// The cache map is created via Init so that resetCountCache (which invalidates
// the state) causes a fresh, empty map on the next render — hence a re-scan.
func cachedCount(wnd core.Window, stateID, key string, fetch func() (int64, error)) int64 {
	m := core.StateOf[countCache](wnd, stateID).Init(func() countCache {
		return countCache{}
	}).Get()
	if v, ok := m[key]; ok {
		return v
	}
	n, err := fetch()
	if err != nil {
		n = -1
	}
	m[key] = n
	return n
}

// resetCountCache invalidates the count cache under stateID and triggers a
// re-render, so its entries are re-scanned. Call it after a destructive action
// that changes entry counts (delete, flush).
func resetCountCache(wnd core.Window, stateID string) {
	core.StateOf[countCache](wnd, stateID).Reset()
}

// countLabel renders a cached count as a short label; a negative (error) count
// shows a placeholder.
func countLabel(n int64) string {
	if n < 0 {
		return "?"
	}
	return fmt.Sprintf("%d", n)
}

// fmtMillis renders a unix-milli timestamp as a compact UTC string.
func fmtMillis(ms int64) string {
	return time.UnixMilli(ms).UTC().Format("2006-01-02 15:04:05.000")
}

// fmtAxisMillis renders a unix-milli timestamp for a chart x-axis label,
// adapting the format to the span: time-only for spans below one day, else
// date+time.
func fmtAxisMillis(ms, spanMillis int64) string {
	t := time.UnixMilli(ms).UTC()
	if spanMillis < int64(24*time.Hour/time.Millisecond) {
		return t.Format("15:04:05.000")
	}
	return t.Format("2006-01-02 15:04")
}
