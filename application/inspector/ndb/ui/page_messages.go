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
)

// selKey is a string alias satisfying dropdown's ~string constraint.
type selKey string

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
		return alert.Banner("Kein Zugriff", "Es fehlt die Berechtigung nago.ndb.inspector.")
	}

	instances, err := uc.Instances(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}
	if len(instances) == 0 {
		return ui.VStack(
			header("ndb Nachrichten"),
			alert.Banner("Keine ndb Datenbank", "Es wurde keine ndb Datenbank registriert."),
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
		right = alert.Banner("Keine Message-Datenbank", "In dieser ndb Datenbank wurde keine msgstore-Engine gefunden.")
	default:
		right = messageWindow(wnd, uc, instancePath, engine, selectedTypesOf(selectedTypes, types), fromSeq, invalidate)
	}

	return ui.VStack(
		header("ndb Nachrichten"),
		ui.Space(ui.L16),
		ui.HStack(
			ui.VStack(
				dropdown.Dropdown[selKey]("Datenbank", instOpts, selectedInstance.Get()).
					InputValue(selectedInstance).
					Frame(ui.Frame{}.FullWidth()),
				ui.Space(ui.L8),
				dropdown.Dropdown[selKey]("Engine", engOpts, selectedEngine.Get()).
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

func header(title string) core.View {
	return ui.HStack(ui.Text(title).Font(ui.Title), ui.Spacer()).FullWidth()
}

// typePicker is the multi-select filter over the available message types. It is
// the primary way to choose which streams appear in the list. When the selection
// changes, the seek cursor resets to the beginning.
func typePicker(wnd core.Window, types []ndbinspector.TypeInfo, selected *core.State[[]selKey], fromSeq *core.State[int64]) core.View {
	if len(types) == 0 {
		return ui.Text("Keine Streams vorhanden.").Font(ui.Small)
	}
	values := make([]selKey, 0, len(types))
	for _, ti := range types {
		values = append(values, selKey(ti.Type))
	}
	selected.Observe(func([]selKey) { fromSeq.Set(0) })

	return picker.Picker[selKey]("Nachrichtentypen", values, selected).
		MultiSelect(true).
		Stringer(func(k selKey) string { return string(k) }).
		SupportingText("Mehrfachauswahl – leer = alle Typen").
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
	return ui.HStack(
		ui.VStack(
			ui.Text(string(ti.Type)).Font(ui.BodyLarge),
			ui.Text(fmt.Sprintf("Seq %d–%d%s · %d Segmente · %s",
				ti.MinSeq, ti.MaxSeq, pending, ti.Segments,
				xstrings.FormatByteSize(wnd.Locale(), ti.Bytes, 1))).Font(ui.Small),
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
		return ui.Text("Wähle eine Engine aus, um Nachrichten anzuzeigen.")
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
		ui.IntField("Ab Seq", fromSeq.Get(), fromSeq).Frame(ui.Frame{Width: ui.L160}),

		ui.Spacer(),
		ui.SecondaryButton(func() {
			if err := uc.RebuildTimeIndex(wnd.Subject(), instancePath, engine); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			invalidate.Set(invalidate.Get() + 1)
		}).PreIcon(icons.Refresh).Title("Zeitindex"),
	).FullWidth()

	// The table (incl. its pager footer) is always rendered – even for an empty
	// window – so the user can always page forward/back with the footer pager.
	table := messageTable(wnd, uc, instancePath, engine, rows, pageSize, fromSeq, showType, showPreview, invalidate)

	return ui.VStack(controls, ui.Space(ui.L16), table).FullWidth().Alignment(ui.Leading)
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

// maxPages is the number of pages the in-window pager can navigate before the
// seq cursor advances to the next window (limit = pageSize * maxPages).
const maxPages = 5

func messageTable(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, rows []ndbinspector.MessageRow, pageSize int, fromSeq *core.State[int64], showType, showPreview bool, invalidate *core.State[int]) core.View {
	cols := []ui.TTableColumn{
		ui.TableColumn(ui.Text("Seq")),
		ui.TableColumn(ui.Text("Zeit")),
	}
	if showType {
		cols = append(cols, ui.TableColumn(ui.Text("Typ")))
	}
	cols = append(cols,
		ui.TableColumn(ui.Text("Trace")),
		ui.TableColumn(ui.Text("Größe")),
	)
	if showPreview {
		cols = append(cols, ui.TableColumn(ui.Text("Vorschau")))
	}
	cols = append(cols, ui.TableColumn(ui.Text("")))

	// page within the fetched window
	pageIdx := core.StateOf[int](wnd, "ndbmsg-page")
	pageCount := (len(rows) + pageSize - 1) / pageSize
	if pageCount < 1 {
		pageCount = 1
	}
	if pageIdx.Get() >= pageCount {
		pageIdx.Set(pageCount - 1)
	}
	if pageIdx.Get() < 0 {
		pageIdx.Set(0)
	}
	start := pageIdx.Get() * pageSize
	end := min(start+pageSize, len(rows))
	visible := rows[start:end]

	tableRows := make([]ui.TTableRow, 0, len(visible)+1)
	for _, r := range visible {
		tableRows = append(tableRows, messageRow(wnd, uc, instancePath, engine, r, showType, showPreview, invalidate))
	}
	// dataview-style pager footer, always present so navigation is always possible.
	tableRows = append(tableRows, pagerFooter(len(cols), rows, pageSize, pageIdx, pageCount, fromSeq))

	return ui.Table(cols...).Rows(tableRows...).Frame(ui.Frame{}.FullWidth())
}

// pagerFooter mimics the dataview optic: a footer row with an entries label, a
// spacer and prev/next chevrons on a card-footer background. It pages within the
// fetched window; on the boundaries it advances/retreats the seq cursor so the
// user can always navigate, even when the current window is empty.
func pagerFooter(colCount int, rows []ndbinspector.MessageRow, pageSize int, pageIdx *core.State[int], pageCount int, fromSeq *core.State[int64]) ui.TTableRow {
	var label string
	if len(rows) == 0 {
		label = "Keine Nachrichten in diesem Fenster"
	} else {
		start := pageIdx.Get() * pageSize
		end := min(start+pageSize, len(rows))
		label = fmt.Sprintf("Seq %d–%d · Seite %d von %d",
			rows[start].Seq, rows[end-1].Seq, pageIdx.Get()+1, pageCount)
	}

	canPrev := pageIdx.Get() > 0 || fromSeq.Get() > 0
	// A next window exists if the fetch filled the whole limit; otherwise this
	// was the last (partial) window.
	windowFull := len(rows) == pageSize*maxPages
	canNext := pageIdx.Get() < pageCount-1 || windowFull

	prev := func() {
		if pageIdx.Get() > 0 {
			pageIdx.Set(pageIdx.Get() - 1)
			return
		}
		// step the seq window back by one full fetch and land on its last page.
		n := fromSeq.Get() - int64(pageSize*maxPages)
		if n < 0 {
			n = 0
		}
		fromSeq.Set(n)
		pageIdx.Set(maxPages - 1)
	}
	next := func() {
		if pageIdx.Get() < pageCount-1 {
			pageIdx.Set(pageIdx.Get() + 1)
			return
		}
		// advance the seq window to just after the last visible message.
		if len(rows) > 0 {
			fromSeq.Set(int64(rows[len(rows)-1].Seq) + 1)
		}
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
		cells = append(cells, ui.TableCell(ui.Text(payloadPreview(r.Payload))))
	}
	cells = append(cells, ui.TableCell(deleteSeqButton(wnd, uc, instancePath, engine, r.Type, r.Seq, invalidate)))

	return ui.TableRow(cells...).Action(func() { detailPresented.Set(true) })
}

func messageDetailDialog(wnd core.Window, r ndbinspector.MessageRow, presented *core.State[bool]) core.View {
	meta := ui.Table(
		ui.TableColumn(ui.Text("Feld")),
		ui.TableColumn(ui.Text("Wert")),
	).Rows(
		metaRow("Seq", fmt.Sprintf("%d", r.Seq)),
		metaRow("Typ", string(r.Type)),
		metaRow("Zeit", time.Unix(0, r.TimeNano).Format(time.RFC3339Nano)),
		metaRow("TraceID", r.TraceID),
		metaRow("Encoding", fmt.Sprintf("%d", r.Encoding)),
		metaRow("Tombstone", fmt.Sprintf("%v", r.Tomb)),
		metaRow("Größe", xstrings.FormatByteSize(wnd.Locale(), int64(r.Size), 2)),
	).Frame(ui.Frame{}.FullWidth())

	body := ui.VStack(
		meta,
		ui.Space(ui.L8),
		payloadView(r.Payload),
	).Alignment(ui.Leading).FullWidth()

	return alert.Dialog(fmt.Sprintf("Nachricht %d", r.Seq), body, presented, alert.Larger(), alert.Ok(), alert.Closeable())
}

func metaRow(field, value string) ui.TTableRow {
	return ui.TableRow(
		ui.TableCell(ui.Text(field).Font(ui.BodyLarge)),
		ui.TableCell(ui.Text(value)),
	)
}

func payloadView(payload []byte) core.View {
	if len(payload) == 0 {
		return ui.Text("<leer>").Font(ui.Small)
	}
	if !utf8.Valid(payload) {
		return ui.Text(fmt.Sprintf("<binäre Daten, %d Bytes>", len(payload))).Font(ui.Small)
	}
	return ui.CodeEditor(string(payload)).Frame(ui.Frame{Height: ui.L320}.FullWidth()).Language("json")
}

func deleteSeqButton(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine string, typeID ndb.TypeID, seq ndb.Seq, invalidate *core.State[int]) core.View {
	presented := core.StateOf[bool](wnd, fmt.Sprintf("delseq-%s-%s-%d", instancePath, engine, seq))
	return ui.HStack(
		alert.Dialog("Nachricht löschen",
			ui.Text(fmt.Sprintf("Seq %d aus Stream %q als gelöscht markieren (Tombstone)?", seq, typeID)),
			presented,
			alert.Cancel(nil),
			alert.Delete(func() {
				if err := uc.DeleteSeq(wnd.Subject(), instancePath, engine, typeID, seq); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				invalidate.Set(invalidate.Get() + 1)
			}),
		),
		ui.TertiaryButton(func() { presented.Set(true) }).PreIcon(icons.TrashBin).AccessibilityLabel("Löschen"),
	)
}

func deleteTypeButton(wnd core.Window, uc ndbinspector.UseCases, instancePath, engine, typeID string, invalidate *core.State[int]) core.View {
	presented := core.StateOf[bool](wnd, "deltype-"+instancePath+"-"+engine+"-"+typeID)
	return ui.HStack(
		alert.Dialog("Stream löschen",
			ui.Text(fmt.Sprintf("Den gesamten Stream %q unwiderruflich löschen?", typeID)),
			presented,
			alert.Cancel(nil),
			alert.Delete(func() {
				if err := uc.DeleteType(wnd.Subject(), instancePath, engine, ndb.TypeID(typeID)); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				invalidate.Set(invalidate.Get() + 1)
			}),
		),
		ui.TertiaryButton(func() { presented.Set(true) }).PreIcon(icons.TrashBin).AccessibilityLabel("Stream löschen"),
	)
}

func payloadPreview(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	if !utf8.Valid(payload) {
		return fmt.Sprintf("<binär, %d Bytes>", len(payload))
	}
	return xstrings.EllipsisEnd(string(payload), previewMaxLen)
}

func shortTrace(t string) string {
	return xstrings.EllipsisEnd(t, 12)
}
