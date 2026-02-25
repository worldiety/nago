// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/web/vuejs"
)

type SomeThing struct {
	_text     struct{} `label:"I'm some arbitrary text."`
	Name      string   `id:"abc1234"` // text with single line field
	Desc      string   `lines:"3"`    // same but with 3 lines text field
	Secret    string   `style:"secret"`
	SomeFloat float64
	Duration  time.Duration `style:"decomposed" days:"false" hours:"true" minutes:"true" seconds:"false"`
	Clock     time.Duration `style:"clock" days:"false"`
	Ack       bool
	Integer   int        `value:"42"`
	_hline    struct{}   `label:"---"` // this is a horizontal line (hline)
	When      xtime.Date `supportingText:"When did this happen?"`
	Who       user.ID    `source:"nago.users"` // single select on source
	Who2      user.ID    `source:"nago.users"` // single select on source 2
	Others    []user.ID  `source:"nago.users"` // multi select on source

	_sectionText    struct{} `label:"This is a section in a card." section:"My Section"`
	FavoriteColor   ui.Color `section:"My Section"`
	ImgStyleDefault image.ID `section:"My Section"`
	ImgStyleAvatar  image.ID `section:"My Section" style:"avatar"`
	ImgStyleIcon    image.ID `section:"My Section" style:"icon"`

	SomeStrings []string `section:"Other section"` // each line is a string

}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		option.MustZero(cfg.StandardSystems())
		uid := option.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {
			thingState := core.AutoState[SomeThing](wnd).Init(func() SomeThing {
				return SomeThing{
					Who:    uid,
					Others: []user.ID{uid},
				}
			})
			return ui.VStack(
				form.Auto(form.AutoOptions{Window: wnd}, thingState),
				ui.PrimaryButton(func() {
					fmt.Printf("Thing: %#v\n", thingState.Get())
				}).Title("print"),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())
		}))

	}).Run()
}
