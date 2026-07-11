// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uindbinspector

import (
	ndbinspector "go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"

	"github.com/worldiety/i18n"
)

// columnKnife groups the destructive per-column maintenance tools: flush
// (compaction) and delete the whole data range. Both are permission-gated by the
// use cases and confirmed via a dialog.
func columnKnife(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, ci ndbinspector.ColumnInfo, invalidate *core.State[int]) core.View {
	delPresented := core.StateOf[bool](wnd, "ts-delrange-"+instancePath+"-"+engine+"-"+ci.Key())
	return ui.HStack(
		alert.Dialog(StrDeleteRangeTitle.Get(wnd),
			ui.Text(StrDeleteRangeBody.Get(wnd,
				i18n.String("column", ci.Key()),
				i18n.String("from", fmtMillis(ci.MinMillis)),
				i18n.String("to", fmtMillis(ci.MaxMillis)),
			)),
			delPresented,
			alert.Cancel(nil),
			alert.Delete(func() {
				if err := uc.DeleteSeriesRange(wnd.Subject(), instancePath, engine, ci.Bucket, ci.Column, ci.MinMillis, ci.MaxMillis); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				resetCountCache(wnd, countCacheColumns)
				invalidate.Set(invalidate.Get() + 1)
			}),
		),
		ui.TertiaryButton(func() {
			if err := uc.FlushColumn(wnd.Subject(), instancePath, engine, ci.Bucket, ci.Column); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			resetCountCache(wnd, countCacheColumns)
			invalidate.Set(invalidate.Get() + 1)
		}).PreIcon(icons.Refresh).AccessibilityLabel(StrCompact.Get(wnd)),
		ui.TertiaryButton(func() { delPresented.Set(true) }).PreIcon(icons.TrashBin).AccessibilityLabel(StrDeleteRange.Get(wnd)),
	)
}

// flushKnife is the flush (compaction) action shown in the string-window
// controls of the currently selected column.
func flushKnife(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, ci ndbinspector.ColumnInfo, invalidate *core.State[int]) core.View {
	return ui.SecondaryButton(func() {
		if err := uc.FlushColumn(wnd.Subject(), instancePath, engine, ci.Bucket, ci.Column); err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}
		resetCountCache(wnd, countCacheColumns)
		invalidate.Set(invalidate.Get() + 1)
	}).PreIcon(icons.Refresh).Title(StrCompact.Get(wnd))
}
