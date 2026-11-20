// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"testing"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// focusTestWindow and focusTestContext render a view without an application, like the tests of navsplitview. A
// callback gets a pointer of its own, so an action is visible in the rendered stack.
type focusTestWindow struct{ core.Window }

func (focusTestWindow) Info() core.WindowInfo { return core.WindowInfo{SizeClass: core.SizeClassLarge} }

type focusTestContext struct{ next proto.Ptr }

func (c *focusTestContext) Window() core.Window { return focusTestWindow{} }
func (c *focusTestContext) MountCallback(f func()) proto.Ptr {
	if f == nil {
		return 0
	}
	c.next++
	return c.next
}

// stackFocusable mirrors the rule of the frontend (web/vuejs/src/components/UiStack.vue, computed focusable): a stack
// takes part in the tab order, if it has an action, a focused border, a focused outline or a focused background.
func stackFocusable(s *proto.Stack) bool {
	return s.Action != 0 || !s.Border.Focused.IsZero() || !s.Outline.Focused.IsZero() || s.BackgroundColorStates.Focus != ""
}

func renderStack(t *testing.T, v core.View) *proto.Stack {
	t.Helper()
	s, ok := v.Render(&focusTestContext{}).(*proto.Stack)
	if !ok {
		t.Fatalf("%T renders no stack", v)
	}
	return s
}

// TestStackFocusWithinBorder: FocusWithinBorder puts the border into BorderStates.FocusWithin of the rendered stack,
// beside the border of Border, and leaves the stack out of the tab order. The same border given to Border yields the
// same protocol border in Initial, so FocusWithin carries the border as given. Each counter row changes one call: with
// FocusedBorder instead the border is in Focused and the stack becomes focusable, with Border alone FocusWithin stays
// empty, and an action makes the stack focusable although FocusWithin is set, so the check of the tab order can fail.
func TestStackFocusWithinBorder(t *testing.T) {
	frame := ui.Border{}.Width(ui.L1).Radius(ui.L16)
	focus := frame.Color(ui.I0)
	line := frame.Color(ui.M5)

	// the protocol borders of line and focus as Border renders them
	wantLine := renderStack(t, ui.VStack().Border(line)).Border.Initial
	wantFocus := renderStack(t, ui.VStack().Border(focus)).Border.Initial
	if wantFocus.TopColor != proto.Color(ui.I0) || wantFocus.TopWidth != proto.Length(ui.L1) || wantLine.TopColor != proto.Color(ui.M5) {
		t.Fatalf("premise: Border renders %+v and %+v", wantLine, wantFocus)
	}

	rows := []struct {
		name        string
		view        core.View
		focusWithin bool // FocusWithin carries focus
		initial     bool // Initial carries line
		focused     bool // Focused carries focus
		focusable   bool
	}{
		{name: "FocusWithinBorder", view: ui.VStack().ID("s").FocusWithinBorder(focus), focusWithin: true},
		{name: "FocusWithinBorder and Border", view: ui.VStack().ID("s").FocusWithinBorder(focus).Border(line), focusWithin: true, initial: true},
		{name: "counter row: FocusedBorder", view: ui.VStack().ID("s").FocusedBorder(focus).Border(line), initial: true, focused: true, focusable: true},
		{name: "counter row: Border only", view: ui.VStack().ID("s").Border(line), initial: true},
		{name: "counter row: FocusWithinBorder with an action", view: ui.VStack().ID("s").Action(func() {}).FocusWithinBorder(focus), focusWithin: true, focusable: true},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			s := renderStack(t, r.view)
			check := func(state string, got proto.Border, set bool, want proto.Border) {
				t.Helper()
				if set && got != want {
					t.Errorf("%s: %+v, want %+v", state, got, want)
				}
				if !set && !got.IsZero() {
					t.Errorf("%s: %+v, want none", state, got)
				}
			}
			check("FocusWithin", s.Border.FocusWithin, r.focusWithin, wantFocus)
			check("Initial", s.Border.Initial, r.initial, wantLine)
			check("Focused", s.Border.Focused, r.focused, wantFocus)
			if !s.Border.Hovered.IsZero() || !s.Border.Active.IsZero() {
				t.Errorf("Hovered %+v, Active %+v, want none", s.Border.Hovered, s.Border.Active)
			}
			if got := stackFocusable(s); got != r.focusable {
				t.Errorf("focusable = %v, want %v", got, r.focusable)
			}
		})
	}
}
