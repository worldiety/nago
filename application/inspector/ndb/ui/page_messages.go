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
	"unicode/utf8"

	ndbinspector "go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/picker"

	"github.com/worldiety/i18n"
)

// previewMaxLen is the maximum number of characters shown in the payload
// preview column before it is ellipsized.
const previewMaxLen = 200

// PageMessages renders the Kafka-UI-style message inspector: pick an ndb
// database and engine, filter the message types via a multi-select picker, seek
// to a Seq and page through bounded windows. When several types are selected the
// messages are shown coherently in global Seq order (k-way merge). Nothing is
// ever fully materialized, so it stays responsive over streams with millions of
// messages.
func PageMessages(wnd core.Window, uc ndbinspector.UseCases) core.View {
	if !wnd.Subject().HasPermission(ndbinspector.PermNDBInspector) {
		return alert.Banner(StrNoAccessTitle.Get(wnd), StrNoAccessBody.Get(wnd))
	}

	instances, err := uc.Instances(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}
	if len(instances) == 0 {
		return ui.VStack(
			header(StrMessagesTitle.Get(wnd)),
			alert.Banner(StrNoNdbTitle.Get(wnd), StrNoNdbBody.Get(wnd)),
		).FullWidth().Alignment(ui.Leading)
	}

	invalidate := core.AutoState[int](wnd)
	selectedInstance := core.AutoState[selKey](wnd)
	selectedEngine := core.AutoState[selKey](wnd)
	selectedTypes := core.AutoState[[]selKey](wnd)
	fromSeq := core.AutoState[int64](wnd)

	if selectedInstance.Get() == "" {
		selectedInstance.Set(selKey(instances[0].Path))
	}
	selectedInstance.Observe(func(selKey) {
		selectedEngine.Set("")
		selectedTypes.Set(nil)
		fromSeq.Set(0)
	})
	instancePath := string(selectedInstance.Get())

	engines, err := uc.MessageEngines(wnd.Subject(), instancePath)
	if err != nil {
		return alert.BannerError(err)
	}

	if selectedEngine.Get() == "" && len(engines) > 0 {
		selectedEngine.Set(selKey(engines[0].Name))
	}
	selectedEngine.Observe(func(selKey) {
		selectedTypes.Set(nil)
		fromSeq.Set(0)
	})
	engine := string(selectedEngine.Get())

	instOpts := make([]dropdown.Option[selKey], 0, len(instances))
	for _, in := range instances {
		instOpts = append(instOpts, dropdown.Option[selKey]{Value: selKey(in.Path), Label: in.Name})
	}
	engOpts := make([]dropdown.Option[selKey], 0, len(engines))
	for _, e := range engines {
		// include the engine implementation kind (the Go short-name of the
		// engine type, e.g. "msgstore") so heterogeneous instances are clear.
		engOpts = append(engOpts, dropdown.Option[selKey]{
			Value: selKey(e.Name),
			Label: fmt.Sprintf("%s (%s)", e.Name, e.Kind),
		})
	}

	var types []ndbinspector.TypeInfo
	if engine != "" {
		if types, err = uc.Types(wnd.Subject(), instancePath, engine); err != nil {
			return alert.BannerError(err)
		}
	}
	// keep the selection valid if a stream vanished (e.g. after delete)
	pruneSelection(selectedTypes, types)

	var right core.View
	switch {
	case len(engines) == 0:
		right = alert.Banner(StrNoMsgEngineTitle.Get(wnd), StrNoMsgEngineBody.Get(wnd))
	default:
		right = messageWindow(wnd, uc, instancePath, engine, selectedTypesOf(selectedTypes, types), fromSeq, invalidate)
	}

	return ui.VStack(
		header(StrMessagesTitle.Get(wnd)),
		ui.Space(ui.L16),
		ui.HStack(
			ui.VStack(
				dropdown.Dropdown[selKey](StrDatabase.Get(wnd), instOpts, selectedInstance.Get()).
					InputValue(selectedInstance).
					Frame(ui.Frame{}.FullWidth()),
				ui.Space(ui.L8),
				dropdown.Dropdown[selKey](StrEngine.Get(wnd), engOpts, selectedEngine.Get()).
					InputValue(selectedEngine).
					Frame(ui.Frame{}.FullWidth()),
				ui.Space(ui.L8),
				typePicker(wnd, types, selectedTypes, fromSeq),
				ui.Space(ui.L8),
				ui.ScrollView(streamList(wnd, uc, instancePath, engine, types, invalidate)).
					Axis(ui.ScrollViewAxisVertical),
			).Alignment(ui.Top).Frame(ui.Frame{Width: ui.L400, MaxWidth: ui.L400}),
			ui.VLine().Frame(ui.Frame{}),
			ui.VStack(right).Alignment(ui.Top).FullWidth(),
		).FullWidth().Alignment(ui.Stretch),
	).FullWidth().Alignment(ui.Leading)
}

// typePicker is the multi-select filter over the available message types. It is
// the primary way to choose which streams appear in the list. When the selection
// changes, the seek cursor resets to the beginning.
func typePicker(wnd core.Window, types []ndbinspector.TypeInfo, selected *core.State[[]selKey], fromSeq *core.State[int64]) core.View {
	if len(types) == 0 {
		return ui.Text(StrNoStreams.Get(wnd)).Font(ui.Small)
	}
	values := make([]selKey, 0, len(types))
	for _, ti := range types {
		values = append(values, selKey(ti.Type))
	}
	selected.Observe(func([]selKey) { fromSeq.Set(0) })

	return picker.Picker[selKey](StrMessageTypes.Get(wnd), values, selected).
		MultiSelect(true).
		Stringer(func(k selKey) string { return string(k) }).
		SupportingText(StrMessageTypesHint.Get(wnd)).
		Frame(ui.Frame{}.FullWidth())
}

// selectedTypesOf resolves the picker selection to concrete ndb type ids. An
// empty selection means "all types".
func selectedTypesOf(selected *core.State[[]selKey], types []ndbinspector.TypeInfo) []ndb.TypeID {
	sel := selected.Get()
	if len(sel) == 0 {
		return nil // all
	}
	valid := map[selKey]bool{}
	for _, ti := range types {
		valid[selKey(ti.Type)] = true
	}
	out := make([]ndb.TypeID, 0, len(sel))
	for _, k := range sel {
		if valid[k] {
			out = append(out, ndb.TypeID(k))
		}
	}
	return out
}

// pruneSelection drops selected types that no longer exist so the picker never
// carries stale entries.
func pruneSelection(selected *core.State[[]selKey], types []ndbinspector.TypeInfo) {
	sel := selected.Get()
	if len(sel) == 0 {
		return
	}
	valid := map[selKey]bool{}
	for _, ti := range types {
		valid[selKey(ti.Type)] = true
	}
	pruned := sel[:0:0]
	for _, k := range sel {
		if valid[k] {
			pruned = append(pruned, k)
		}
	}
	if len(pruned) != len(sel) {
		selected.Set(pruned)
	}
}

// streamList shows per-stream statistics and a per-stream delete knife tool. It
// is informational; selection happens through the type picker above.
func streamList(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, types []ndbinspector.TypeInfo, invalidate *core.State[int]) core.View {
	if engine == "" {
		return nil
	}
	if len(types) == 0 {
		return nil
	}
	rows := make([]core.View, 0, len(types))
	for _, ti := range types {
		rows = append(rows, streamStatRow(wnd, uc, instancePath, engine, ti, invalidate))
	}
	return ui.VStack(rows...).FullWidth().Alignment(ui.Leading).Gap(ui.L8)
}

func streamStatRow(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, ti ndbinspector.TypeInfo, invalidate *core.State[int]) core.View {
	pending := ""
	if ti.HasPending {
		pending = " +"
	}
	// The exact message count needs a full replay; cache it per page under a
	// fixed id so the scan runs at most once per visit. Destructive actions call
	// resetCountCache to re-scan the affected counts.
	key := instancePath + "|" + engine + "|" + string(ti.Type)
	count := cachedCount(wnd, countCacheMessages, key, func() (int64, error) {
		return uc.CountType(wnd.Subject(), instancePath, engine, ti.Type)
	})

	return ui.HStack(
		ui.VStack(
			ui.Text(string(ti.Type)).Font(ui.BodyLarge),
			ui.Text(StrMsgStatRow.Get(wnd,
				i18n.Int("min", int(ti.MinSeq)),
				i18n.Int("max", int(ti.MaxSeq)),
				i18n.String("pending", pending),
				i18n.String("count", countLabel(count)),
				i18n.Int("segments", ti.Segments),
				i18n.String("size", xstrings.FormatByteSize(wnd.Locale(), ti.Bytes, 1)),
			)).Font(ui.Small),
		).Alignment(ui.Leading),
		ui.Spacer(),
		deleteTypeButton(wnd, uc, instancePath, engine, string(ti.Type), invalidate),
	).Alignment(ui.Center).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8)).
		Frame(ui.Frame{}.FullWidth())
}

func messageWindow(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, selTypes []ndb.TypeID, fromSeq *core.State[int64], invalidate *core.State[int]) core.View {
	if engine == "" {
		return ui.Text(StrSelectEngineHint.Get(wnd))
	}

	// Page size is derived from the available window height exactly like the
	// blob store inspector, so it needs no state. limit fetches enough rows for
	// at most maxPages pages, so the in-window pager can navigate that many pages
	// before the seq cursor has to advance to the next window.
	pageSize := pageSizeFor(wnd)
	limit := pageSize * maxPages

	rows, err := uc.Window(wnd.Subject(), ndbinspector.WindowRequest{
		Instance: instancePath,
		Engine:   engine,
		Types:    selTypes,
		MinSeq:   ndb.Seq(fromSeq.Get()),
		Limit:    limit,
	})
	if err != nil {
		return alert.BannerError(err)
	}

	// column layout depends on selection and screen size:
	//  - the type column is only shown when more than one type is selected
	//    (with an empty/all selection the stream can be mixed, so also show it),
	//  - the preview column is hidden on small screens.
	showType := len(selTypes) != 1
	showPreview := wnd.Info().SizeClass > core.SizeClassSmall

	controls := ui.HStack(
		ui.IntField(StrFromSeq.Get(wnd), fromSeq.Get(), fromSeq).Frame(ui.Frame{Width: ui.L160}),

		ui.Spacer(),
		ui.SecondaryButton(func() {
			if err := uc.RebuildTimeIndex(wnd.Subject(), instancePath, engine); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			invalidate.Set(invalidate.Get() + 1)
		}).PreIcon(icons.Refresh).Title(StrTimeIndex.Get(wnd)),
	).FullWidth()

	// The table (incl. its pager footer) is always rendered – even for an empty
	// window – so the user can always page forward/back with the footer pager.
	table := messageTable(wnd, uc, instancePath, engine, rows, pageSize, fromSeq, showType, showPreview, invalidate)

	return ui.VStack(controls, ui.Space(ui.L16), table).FullWidth().Alignment(ui.Leading)
}

func messageTable(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, rows []ndbinspector.MessageRow, pageSize int, fromSeq *core.State[int64], showType, showPreview bool, invalidate *core.State[int]) core.View {
	cols := []ui.TTableColumn{
		ui.TableColumn(ui.Text(StrColSeq.Get(wnd))),
		ui.TableColumn(ui.Text(StrColTime.Get(wnd))),
	}
	if showType {
		cols = append(cols, ui.TableColumn(ui.Text(StrColType.Get(wnd))))
	}
	cols = append(cols,
		ui.TableColumn(ui.Text(StrColTrace.Get(wnd))),
		ui.TableColumn(ui.Text(StrColSize.Get(wnd))),
	)
	if showPreview {
		cols = append(cols, ui.TableColumn(ui.Text(StrColPreview.Get(wnd))))
	}
	cols = append(cols, ui.TableColumn(ui.Text("")))

	// page within the fetched window
	pageIdx := core.StateOf[int](wnd, "ndbmsg-page")
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
		tableRows = append(tableRows, messageRow(wnd, uc, instancePath, engine, r, showType, showPreview, invalidate))
	}
	// dataview-style pager footer, always present so navigation is always possible.
	tableRows = append(tableRows, messagePagerFooter(wnd, len(cols), rows, pageSize, pageIdx, pageCount, fromSeq))

	return ui.Table(cols...).Rows(tableRows...).Frame(ui.Frame{}.FullWidth())
}

// messagePagerFooter builds the footer with a message-specific label and the
// seq-cursor window shifts, delegating the shared optic to footerPager.
func messagePagerFooter(wnd core.Window, colCount int, rows []ndbinspector.MessageRow, pageSize int, pageIdx *core.State[int], pageCount int, fromSeq *core.State[int64]) ui.TTableRow {
	var label string
	if len(rows) == 0 {
		label = StrNoMessagesInWindow.Get(wnd)
	} else {
		start := pageIdx.Get() * pageSize
		end := min(start+pageSize, len(rows))
		label = StrMsgPageLabel.Get(wnd,
			i18n.Int("min", int(rows[start].Seq)),
			i18n.Int("max", int(rows[end-1].Seq)),
			i18n.Int("page", pageIdx.Get()+1),
			i18n.Int("pages", pageCount),
		)
	}

	windowFull := len(rows) == pageSize*maxPages
	onPrev := func() {
		n := fromSeq.Get() - int64(pageSize*maxPages)
		if n < 0 {
			n = 0
		}
		fromSeq.Set(n)
	}
	onNext := func() {
		if len(rows) > 0 {
			fromSeq.Set(int64(rows[len(rows)-1].Seq) + 1)
		}
	}

	return footerPager(colCount, label, len(rows) == 0, pageIdx, pageCount,
		fromSeq.Get() == 0, windowFull, onPrev, onNext)
}

func messageRow(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, r ndbinspector.MessageRow, showType, showPreview bool, invalidate *core.State[int]) ui.TTableRow {
	ts := time.Unix(0, r.TimeNano).Format("2006-01-02 15:04:05.000")
	seqLabel := fmt.Sprintf("%d", r.Seq)
	if r.Tomb {
		seqLabel += " ⌫"
	}

	detailPresented := core.StateOf[bool](wnd, fmt.Sprintf("msgdlg-%s-%s-%d", instancePath, engine, r.Seq))

	cells := []ui.TTableCell{
		ui.TableCell(ui.HStack(
			messageDetailDialog(wnd, r, detailPresented),
			ui.Text(seqLabel),
		)),
		ui.TableCell(ui.Text(ts)),
	}
	if showType {
		cells = append(cells, ui.TableCell(ui.Text(string(r.Type))))
	}
	cells = append(cells,
		ui.TableCell(ui.Text(shortTrace(r.TraceID))),
		ui.TableCell(ui.Text(xstrings.FormatByteSize(wnd.Locale(), int64(r.Size), 1))),
	)
	if showPreview {
		cells = append(cells, ui.TableCell(ui.Text(payloadPreview(wnd, r.Payload))))
	}
	cells = append(cells, ui.TableCell(deleteSeqButton(wnd, uc, instancePath, engine, r.Type, r.Seq, invalidate)))

	return ui.TableRow(cells...).Action(func() { detailPresented.Set(true) })
}

func messageDetailDialog(wnd core.Window, r ndbinspector.MessageRow, presented *core.State[bool]) core.View {
	meta := ui.Table(
		ui.TableColumn(ui.Text(StrColField.Get(wnd))),
		ui.TableColumn(ui.Text(StrColValue.Get(wnd))),
	).Rows(
		metaRow("Seq", fmt.Sprintf("%d", r.Seq)),
		metaRow(StrColType.Get(wnd), string(r.Type)),
		metaRow(StrColTime.Get(wnd), time.Unix(0, r.TimeNano).Format(time.RFC3339Nano)),
		metaRow("TraceID", r.TraceID),
		metaRow("Encoding", fmt.Sprintf("%d", r.Encoding)),
		metaRow("Tombstone", fmt.Sprintf("%v", r.Tomb)),
		metaRow(StrColSize.Get(wnd), xstrings.FormatByteSize(wnd.Locale(), int64(r.Size), 2)),
	).Frame(ui.Frame{}.FullWidth())

	body := ui.VStack(
		meta,
		ui.Space(ui.L8),
		payloadView(wnd, r.Payload),
	).Alignment(ui.Leading).FullWidth()

	return alert.Dialog(StrMessageX.Get(wnd, i18n.Int("seq", int(r.Seq))), body, presented, alert.Larger(), alert.Ok(), alert.Closeable())
}

func metaRow(field, value string) ui.TTableRow {
	return ui.TableRow(
		ui.TableCell(ui.Text(field).Font(ui.BodyLarge)),
		ui.TableCell(ui.Text(value)),
	)
}

func payloadView(wnd core.Window, payload []byte) core.View {
	if len(payload) == 0 {
		return ui.Text(StrEmpty.Get(wnd)).Font(ui.Small)
	}
	if !utf8.Valid(payload) {
		return ui.Text(StrBinaryDataX.Get(wnd, i18n.Int("bytes", len(payload)))).Font(ui.Small)
	}
	return ui.CodeEditor(string(payload)).Frame(ui.Frame{Height: ui.L320}.FullWidth()).Language("json")
}

func deleteSeqButton(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, typeID ndb.TypeID, seq ndb.Seq, invalidate *core.State[int]) core.View {
	presented := core.StateOf[bool](wnd, fmt.Sprintf("delseq-%s-%s-%d", instancePath, engine, seq))
	return ui.HStack(
		alert.Dialog(StrDeleteMessage.Get(wnd),
			ui.Text(StrDeleteMessageBody.Get(wnd, i18n.Int("seq", int(seq)), i18n.String("type", string(typeID)))),
			presented,
			alert.Cancel(nil),
			alert.Delete(func() {
				if err := uc.DeleteSeq(wnd.Subject(), instancePath, engine, typeID, seq); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				resetCountCache(wnd, countCacheMessages)
				invalidate.Set(invalidate.Get() + 1)
			}),
		),
		ui.TertiaryButton(func() { presented.Set(true) }).PreIcon(icons.TrashBin).AccessibilityLabel(StrDelete.Get(wnd)),
	)
}

func deleteTypeButton(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine, typeID string, invalidate *core.State[int]) core.View {
	presented := core.StateOf[bool](wnd, "deltype-"+instancePath+"-"+engine+"-"+typeID)
	return ui.HStack(
		alert.Dialog(StrDeleteStream.Get(wnd),
			ui.Text(StrDeleteStreamBody.Get(wnd, i18n.String("type", typeID))),
			presented,
			alert.Cancel(nil),
			alert.Delete(func() {
				if err := uc.DeleteType(wnd.Subject(), instancePath, engine, ndb.TypeID(typeID)); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				resetCountCache(wnd, countCacheMessages)
				invalidate.Set(invalidate.Get() + 1)
			}),
		),
		ui.TertiaryButton(func() { presented.Set(true) }).PreIcon(icons.TrashBin).AccessibilityLabel(StrDeleteStream.Get(wnd)),
	)
}

func payloadPreview(wnd core.Window, payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	if !utf8.Valid(payload) {
		return StrBinaryDataX.Get(wnd, i18n.Int("bytes", len(payload)))
	}
	return xstrings.EllipsisEnd(string(payload), previewMaxLen)
}

func shortTrace(t string) string {
	return xstrings.EllipsisEnd(t, 12)
}
