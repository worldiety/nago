// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"reflect"
	"testing"

	nui "go.wdy.de/nago/presentation/ui"
)

// The files of batch b4 declare no enums, so the tests compare the views of the mirror with the views of nago, which
// are built by the same calls. This checks the arguments and their conversions, including the variadic ones.
//
// L0 and L1 are not used: the length constants of length.go in the template differ from those of nago in them.

// TestTableB4 compares a table of the mirror with the table of nago.
func TestTableB4(t *testing.T) {
	got := Core(Table(
		TableColumn(Text("Name")).Width(L120).Alignment(Leading).Key("sort", "name").Span(2).HoveredBackgroundColor(M3),
		TableColumn(Text("Age")).BackgroundColor(M2).Padding(Padding{}.All(L8)).Border(Border{}.Width(L2)),
	).Rows(
		TableRow(TableCell(Text("Max")).ColSpan(2).RowSpan(1).Key("cell", "max"), TableCell(Text("42")).Alignment(Trailing).Padding(Padding{}.All(L4))).Height(L48).Key("row", "max").BackgroundColor(M1).HoveredBackgroundColor(M6),
		TableRow().Append(TableCell(Text("Eva")).BackgroundColor(M1).Border(Border{}.Radius(L8))),
	).Frame(Frame{Width: Full}).BackgroundColor(M2).Border(Border{}.Radius(L20)).RowDividerColor(M5).HeaderDividerColor(M4).CellPadding(Padding{}.Horizontal(L24)))

	want := nui.Table(
		nui.TableColumn(nui.Text("Name")).Width(nui.L120).Alignment(nui.Leading).Key("sort", "name").Span(2).HoveredBackgroundColor(nui.M3),
		nui.TableColumn(nui.Text("Age")).BackgroundColor(nui.M2).Padding(nui.Padding{}.All(nui.L8)).Border(nui.Border{}.Width(nui.L2)),
	).Rows(
		nui.TableRow(nui.TableCell(nui.Text("Max")).ColSpan(2).RowSpan(1).Key("cell", "max"), nui.TableCell(nui.Text("42")).Alignment(nui.Trailing).Padding(nui.Padding{}.All(nui.L4))).Height(nui.L48).Key("row", "max").BackgroundColor(nui.M1).HoveredBackgroundColor(nui.M6),
		nui.TableRow().Append(nui.TableCell(nui.Text("Eva")).BackgroundColor(nui.M1).Border(nui.Border{}.Radius(nui.L8))),
	).Frame(nui.Frame{Width: nui.Full}).BackgroundColor(nui.M2).Border(nui.Border{}.Radius(nui.L20)).RowDividerColor(nui.M5).HeaderDividerColor(nui.M4).CellPadding(nui.Padding{}.Horizontal(nui.L24))

	if !reflect.DeepEqual(got, want) {
		t.Errorf("table = %+v, want %+v", got, want)
	}

	// Columns replaces the columns, while the rows are kept.
	gotColumns := Core(Table().Rows(TableRow(TableCell(Text("a")))).Columns(TableColumn(Text("b"))))
	wantColumns := nui.Table().Rows(nui.TableRow(nui.TableCell(nui.Text("a")))).Columns(nui.TableColumn(nui.Text("b")))
	if !reflect.DeepEqual(gotColumns, wantColumns) {
		t.Errorf("columns = %+v, want %+v", gotColumns, wantColumns)
	}
}

// TestGridB4 compares a grid of the mirror with the grid of nago.
func TestGridB4(t *testing.T) {
	got := Core(Grid(
		GridCell(Text("a")).ColStart(1).ColEnd(2).RowStart(1).RowEnd(2).Padding(Padding{}.All(L4)).Alignment(Center).BackgroundColor(ColorCardBody),
		GridCell(Text("b")).ColSpan(2).RowSpan(1).Alignment(Trailing),
	).Append(GridCell(Text("c")).ColStart(3)).Columns(3).Rows(2).Heights(L40, L80).Widths(L120).Gap(L8).RowGap(L4).ColGap(L12).FullWidth().Padding(Padding{}.All(L8)).Border(Border{}.Radius(L8)).Visible(true).AccessibilityLabel("grid").Frame(Frame{MinWidth: L80}).WithFrame(func(f Frame) Frame { return f.Grow() }).Font(BodyMedium).BackgroundColor(M1))

	gridCells := nui.Grid(
		nui.GridCell(nui.Text("a")).ColStart(1).ColEnd(2).RowStart(1).RowEnd(2).Padding(nui.Padding{}.All(nui.L4)).Alignment(nui.Center).BackgroundColor(nui.ColorCardBody),
		nui.GridCell(nui.Text("b")).ColSpan(2).RowSpan(1).Alignment(nui.Trailing),
	).Append(nui.GridCell(nui.Text("c")).ColStart(3)).Columns(3).Rows(2).Heights(nui.L40, nui.L80).Widths(nui.L120).Gap(nui.L8).RowGap(nui.L4).ColGap(nui.L12).FullWidth()
	// the methods of the mirror return a grid, while the methods of nago return a DecoredView, which has no Font
	var want nui.DecoredView = gridCells.Padding(nui.Padding{}.All(nui.L8)).(nui.TGrid)
	want = want.Border(nui.Border{}.Radius(nui.L8))
	want = want.Visible(true)
	want = want.AccessibilityLabel("grid")
	want = want.Frame(nui.Frame{MinWidth: nui.L80})
	want = want.WithFrame(func(f nui.Frame) nui.Frame { return f.Grow() })
	want = want.(nui.TGrid).Font(nui.BodyMedium)
	want = want.(nui.TGrid).BackgroundColor(nui.M1)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("grid = %+v, want %+v", got, want)
	}
}

// TestMenuB4 compares a menu of the mirror with the menu of nago, including an empty group and custom content.
func TestMenuB4(t *testing.T) {
	got := Core(Menu(Text("anchor"),
		MenuGroup(MenuItem(nil, Text("one")), MenuItem(nil, Text("two"))),
		MenuGroup().CustomContent(Text("custom")),
		MenuGroup(),
	).Frame(Frame{Width: L200}).Offset(L4))

	want := nui.Menu(nui.Text("anchor"),
		nui.MenuGroup(nui.MenuItem(nil, nui.Text("one")), nui.MenuItem(nil, nui.Text("two"))),
		nui.MenuGroup().CustomContent(nui.Text("custom")),
		nui.MenuGroup(),
	).Frame(nui.Frame{Width: nui.L200}).Offset(nui.L4)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("menu = %+v, want %+v", got, want)
	}
}

// TestDialogB4 compares a dialog of the mirror with the dialog of nago.
func TestDialogB4(t *testing.T) {
	got := Core(Dialog(Text("body")).Title(Text("title")).PreBody(Text("pre")).TitleX(Text("x")).Footer(Text("footer")).Alignment(Center).ModalPadding(Padding{}.All(L8)).Frame(Frame{Width: L400, MaxHeight: L560}).WithFrame(func(f Frame) Frame { return f.Large() }).DisableBoxLayout(false))

	want := nui.Dialog(nui.Text("body")).Title(nui.Text("title")).PreBody(nui.Text("pre")).TitleX(nui.Text("x")).Footer(nui.Text("footer")).Alignment(nui.Center).ModalPadding(nui.Padding{}.All(nui.L8)).Frame(nui.Frame{Width: nui.L400, MaxHeight: nui.L560}).WithFrame(func(f nui.Frame) nui.Frame { return f.Large() }).DisableBoxLayout(false)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("dialog = %+v, want %+v", got, want)
	}
}

// TestModalB4 compares the modals of the mirror with the modals of nago.
func TestModalB4(t *testing.T) {
	modal := Core(Modal(Text("m")).Top(L8).AllowBackgroundScrolling(true).OnDismissRequest(nil))
	wantModal := nui.Modal(nui.Text("m")).Top(nui.L8).AllowBackgroundScrolling(true).OnDismissRequest(nil)
	if !reflect.DeepEqual(modal, wantModal) {
		t.Errorf("modal = %+v, want %+v", modal, wantModal)
	}

	overlay := Core(Overlay(Text("o")).Top(L16).Left(L8).Right(L8).Bottom(L4).AllowBackgroundScrolling(false))
	wantOverlay := nui.Overlay(nui.Text("o")).Top(nui.L16).Left(nui.L8).Right(nui.L8).Bottom(nui.L4).AllowBackgroundScrolling(false)
	if !reflect.DeepEqual(overlay, wantOverlay) {
		t.Errorf("overlay = %+v, want %+v", overlay, wantOverlay)
	}

	notifications := Core(NotificationsOverlay(Text("n")).Bottom(L4))
	wantNotifications := nui.NotificationsOverlay(nui.Text("n")).Bottom(nui.L4)
	if !reflect.DeepEqual(notifications, wantNotifications) {
		t.Errorf("notifications = %+v, want %+v", notifications, wantNotifications)
	}
}
