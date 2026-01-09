// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/evs"
	cfgevs "go.wdy.de/nago/application/evs/cfg"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type Location string

type ShopEvent interface {
	Event()
	Location() Location
}

type OrderPlaced struct {
	ArticleID string
	Comment   string
	OrderID   string
	Loc       Location
}

func (e OrderPlaced) Location() Location {
	return e.Loc
}

func (OrderPlaced) Event() {}

type OrderPayed struct {
	OrderID string
	Loc     Location
}

func (e OrderPayed) Location() Location {
	return e.Loc
}

func (OrderPayed) Event() {}

type OrderCancelled struct {
	OrderID string
	Loc     Location
}

func (e OrderCancelled) Location() Location {
	return e.Loc
}

func (OrderCancelled) Event() {}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_85")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))

		modShop := option.Must(cfgevs.Enable[ShopEvent](cfg, "test.nago.app", "Shop Events", cfgevs.Options[ShopEvent]{}.WithOptions(
			cfgevs.Schema[OrderPlaced, ShopEvent]("OrderPlaced"),
			cfgevs.Schema[OrderPayed, ShopEvent]("OrderPayed"),
			cfgevs.Schema[OrderCancelled, ShopEvent]("OrderCancelled"),
			cfgevs.Index[Location, ShopEvent](func(e evs.Envelope[ShopEvent]) (Location, error) {
				return e.Data.Location(), nil
			}),
		)))

		_ = modShop // take a look for the things you can build your use cases on

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.Text("See the admin center")
		})
	}).Run()
}

// Architecture style note:
// To align your code better with Event Storming keep your use case function declaration but summarize
// your arguments into a single struct called <UseCaseName>Cmd. The Result of that function is
// the Event which has been stored. See the example below

type PayOrderCmd struct {
	OrderID string
}
type PayOrder func(subject auth.Subject, cmd PayOrderCmd) (OrderPayed, error)

func NewPayOrder(storeEvent evs.Store[ShopEvent]) PayOrder {
	return func(subject auth.Subject, cmd PayOrderCmd) (OrderPayed, error) {
		// permission checks and validation stuff and other bussiness logic
		// if err:=subject(....
		// ...

		evt := OrderPayed{OrderID: cmd.OrderID}
		if _, err := storeEvent(user.SU(), evt, evs.StoreOptions{CreatedBy: subject.ID()}); err != nil {
			return evt, err
		}

		return evt, nil
	}
}
