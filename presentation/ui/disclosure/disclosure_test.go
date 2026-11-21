// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package disclosure_test

import (
	"fmt"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/disclosure"
)

func configure(cfg *application.Configurator) {
	cfg.SetApplicationID("de.worldiety.disclosuretest")

	cfg.RootView("plain", func(wnd core.Window) core.View {
		return ui.VStack(
			disclosure.Disclosure(ui.Text("Region"), ui.Text("region content")).ID("region"),
			disclosure.Disclosure(ui.Text("Addresses"), ui.Text("addresses content")).Open(true).HideSeparator().FullWidth(),
		)
	})

	cfg.RootView("bound", func(wnd core.Window) core.View {
		open := core.AutoState[bool](wnd)
		return ui.VStack(
			disclosure.Disclosure(ui.Text("Region"), ui.Text("region content")).InputValue(open).Open(false),
			ui.Text(fmt.Sprintf("open: %v", open.Get())),
		)
	})
}

func TestDisclosure_Plain(t *testing.T) {
	for _, roundTrip := range []bool{false, true} {
		t.Run(fmt.Sprintf("roundTrip=%v", roundTrip), func(t *testing.T) {
			var opts []nagotest.Option
			if roundTrip {
				opts = append(opts, nagotest.WithRoundTrip())
			}

			w := nagotest.New(t, configure, opts...).Open(t, nil, "plain")

			all := w.FindAll(nagotest.Type[*proto.Disclosure]()).Exactly(2)

			region := all.At(0).Node().Component.(*proto.Disclosure)
			if region.Value || !region.Separator || region.InputValue != 0 || region.Id != "region" {
				t.Fatalf("unexpected region disclosure: %+v", region)
			}

			addresses := all.At(1).Node().Component.(*proto.Disclosure)
			if !addresses.Value || addresses.Separator || addresses.Frame.Width != proto.Length(ui.Full) {
				t.Fatalf("unexpected addresses disclosure: %+v", addresses)
			}

			// summary and content are always part of the tree, also of a closed disclosure
			all.At(0).Find(nagotest.Text("Region"))
			all.At(0).Find(nagotest.Text("region content"))
			all.At(1).Find(nagotest.Text("addresses content"))
		})
	}
}

func TestDisclosure_BoundState(t *testing.T) {
	w := nagotest.New(t, configure, nagotest.WithRoundTrip()).Open(t, nil, "bound")

	d := w.Find(nagotest.Type[*proto.Disclosure]())
	if d.Node().Component.(*proto.Disclosure).InputValue == 0 {
		t.Fatal("expected the bound state")
	}

	// the browser reports the open state of the user
	w.Type(d, "true")
	w.Find(nagotest.Text("open: true"))

	// and the state takes precedence over Open(false)
	if !w.Find(nagotest.Type[*proto.Disclosure]()).Node().Component.(*proto.Disclosure).Value {
		t.Fatal("expected the value of the state")
	}
}
