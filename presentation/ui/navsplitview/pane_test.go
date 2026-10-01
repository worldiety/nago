// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package navsplitview_test

import (
	"testing"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/navsplitview"
)

type testWindow struct {
	core.Window
	sizeClass core.WindowSizeClass
}

func (w testWindow) Values() core.Values   { return core.Values{} }
func (w testWindow) Info() core.WindowInfo { return core.WindowInfo{SizeClass: w.sizeClass} }

type testRenderContext struct{ wnd core.Window }

func (c testRenderContext) Window() core.Window            { return c.wnd }
func (c testRenderContext) MountCallback(func()) proto.Ptr { return 0 }

func testNav(id navsplitview.ViewID) core.View {
	return ui.Text(string(id))
}

func TestTwoColumnPanesMayShrinkBelowTheirContent(t *testing.T) {
	ctx := testRenderContext{wnd: testWindow{sizeClass: core.SizeClassLarge}}
	node := navsplitview.TwoColumn(navsplitview.NavFn(testNav)).Default("content", "detail").Render(ctx)

	assertPanes(t, node, map[int]ui.Alignment{0: ui.Center, 2: ui.Center})
}

func TestThreeColumnPanesMayShrinkBelowTheirContent(t *testing.T) {
	ctx := testRenderContext{wnd: testWindow{sizeClass: core.SizeClassXL}}
	node := navsplitview.ThreeColumn(navsplitview.NavFn(testNav)).
		Default("sidebar", "content", "detail").
		AlignmentSidebar(ui.TopLeading).
		Render(ctx)

	assertPanes(t, node, map[int]ui.Alignment{0: ui.TopLeading, 2: ui.Center, 4: ui.Center})
}

func assertPanes(t *testing.T, node core.RenderNode, cells map[int]ui.Alignment) {
	t.Helper()

	grid, ok := node.(*proto.Grid)
	if !ok {
		t.Fatalf("expected a grid, got %T", node)
	}

	for idx, alignment := range cells {
		pane, ok := grid.Cells[idx].Body.(*proto.Stack)
		if !ok {
			t.Fatalf("cell %d: expected a stack around the view, got %T", idx, grid.Cells[idx].Body)
		}

		if pane.Frame.MinWidth != "0px" || pane.Frame.Width != proto.Length(ui.Full) || pane.Frame.Height != proto.Length(ui.Full) {
			t.Errorf("cell %d: pane must fill the cell and may shrink below its content: %+v", idx, pane.Frame)
		}

		if grid.Cells[idx].Alignment != proto.Alignment(ui.Stretch) || pane.Alignment != proto.Alignment(alignment) {
			t.Errorf("cell %d: pane must stretch and place the view with alignment %v", idx, alignment)
		}
	}
}
