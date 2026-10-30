// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package accessibility_test

import (
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accessibility"
)

func configure(cfg *application.Configurator) {
	cfg.SetApplicationID("de.worldiety.accessibility")

	cfg.RootView("absent", func(wnd core.Window) core.View {
		return ui.VStack(
			ui.Text("a"),
			accessibility.Accessible[core.View](nil).Label("nil"),
			accessibility.Accessible(ui.Lazy(func() core.View { return nil })).Label("renders nil"),
			ui.Text("b"),
			// the dialog wraps its optional TitleX into an Accessible
			ui.Dialog(ui.Text("body")).Title(ui.Text("title")),
		)
	})

	cfg.RootView("present", func(wnd core.Window) core.View {
		return accessibility.Accessible(ui.Text("a")).Role(accessibility.RoleButton).Label("x")
	})
}

func TestAccessibleOfAbsentContentIsAbsent(t *testing.T) {
	w := nagotest.New(t, configure).Open(t, nil, "absent")

	// a panic during render shows an error page instead of the tree
	w.Find(nagotest.Text("title"))
	w.FindAll(nagotest.Type[*proto.Accessible]()).None()

	stack := w.Tree().(*proto.Stack)
	if len(stack.Children) != 3 {
		t.Errorf("expected the absent contents to be skipped like nil views, got %d children", len(stack.Children))
	}
}

func TestAccessibleOfContentAppliesAttributes(t *testing.T) {
	w := nagotest.New(t, configure).Open(t, nil, "present")

	node := w.Find(nagotest.Type[*proto.Accessible]()).Node().Component.(*proto.Accessible)
	if node.Role != "button" || node.Label != "x" {
		t.Errorf("unexpected accessible: %#v", node)
	}
	w.Find(nagotest.Text("a"))
}
