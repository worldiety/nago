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
	"go.wdy.de/nago/application/color"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/calendar"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_70")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {

			isLarge := wnd.Info().SizeClass >= core.SizeClassLarge
			chipAlignment := ui.TopTrailing
			if !isLarge {
				chipAlignment = ui.BottomLeading
			}

			txtColor := ui.M8

			return ui.VStack(
				ui.Text("hello world"),
				calendar.Calendar(
					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 7, 11, 0, 0, 0, 0, time.Local),
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 11, 0, 0, 0, 0, time.Local),
						},
						Label:     "Some Event (Torben)",
						Organiser: "Torben",
						Location:  "WZO",
						Lane: calendar.Lane{
							Label: "Torben",
						},
						Category: calendar.Category{
							Label: "Kategorie 2",
							Color: "#ff0000",
						},
						IsCancelled: true,
					},

					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 8, 12, 0, 0, 0, 0, time.Local),
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 13, 0, 0, 0, 0, time.Local),
						},
						Label: "Some Event Day 1",
						Lane: calendar.Lane{
							Label: "Torben",
						},
						Category: calendar.Category{
							Label: "Kategorie 2",
							Color: "#ffff00",
						},
					},

					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 8, 14, 0, 0, 0, 0, time.Local),
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 15, 0, 0, 0, 0, time.Local),
						},
						Label: "Some Event Day 2",
						Lane: calendar.Lane{
							Label: "Torben",
						},
						Organiser: "Olaf",
						Location:  "WZO",
						Category: calendar.Category{
							Label: "Kategorie 2",
							Color: "#ffff00",
						},
						Chips: []calendar.Chip{
							{
								Label:       "Eingetragen",
								Icon:        flowbiteSolid.BadgeCheck,
								StrokeColor: ui.M8,
								BgColor:     "#2BCA73",
								TextColor:   txtColor,
								Alignment:   ui.BottomLeading,
								FullWidth:   true,
							},
						},
					},

					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 10, 11, 0, 0, 0, 0, time.Local),
						},
						To: calendar.Instant{
							At: time.Date(2025, 11, 11, 0, 0, 0, 0, time.Local),
						},
						Label: "Some Event (Torben 2)",
						Lane: calendar.Lane{
							Label: "Torben",
						},
						Category: calendar.Category{
							Label: "Kategorie 2",
							Color: "#ff0000",
						},
						Chips: []calendar.Chip{
							{
								Label:     "Wartelistenplatz: 5",
								Icon:      flowbiteSolid.ClipboardList,
								FillColor: ui.M8,
								BgColor:   "#FBC83E",
								TextColor: txtColor,
								Alignment: ui.BottomLeading,
								FullWidth: true,
							},
						},
					},

					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 2, 1, 15, 30, 0, 0, time.Local),
							Offset: calendar.Offset{
								Label:    "Anfahrt",
								Icon:     flowbiteOutline.Bell,
								Duration: time.Hour * 24 * 3,
							},
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 31, 12, 25, 0, 0, time.Local),
							Offset: calendar.Offset{
								Label:    "Abfahrt",
								Icon:     flowbiteOutline.BellActive,
								Duration: time.Hour * 24 * 6,
							},
						},
						Action: func() {
							fmt.Println("clicked this event")
						},
						Category: calendar.Category{
							Label: "Kategorie 1",
							Color: "#00ff00",
						},
						Label: "Some other Event with more text in a larger box and how does it look with padding? (Olaf)",
						Lane: calendar.Lane{
							Label: "Olaf",
						},
						Chips: []calendar.Chip{
							{
								Label:     "20 | 50",
								Icon:      flowbiteSolid.Users,
								FillColor: txtColor,
								BgColor:   ui.M1,
								TextColor: txtColor,
								Alignment: chipAlignment,
							},
						},
					},
				).
					ViewPort(calendar.Year(2025)).
					Frame(ui.Frame{}.FullWidth()).
					Style(calendar.StartTimeSequence),

				//
				calStartTimeSeqTimeExample(chipAlignment, txtColor),
			).FullWidth().Gap(ui.L16)

		})
	}).Run()
}

func calStartTimeSeqTimeExample(chipAlignment ui.Alignment, txtColor color.Color) core.View {

	return calendar.Calendar(
		calendar.Event{
			From: calendar.Instant{
				At: time.Date(2026, 7, 11, 12, 0, 0, 0, time.Local),
				Offset: calendar.Offset{
					Duration: 1 * time.Hour,
					Icon:     flowbiteOutline.Cart,
					Label:    "Test",
				},
			},
			To: calendar.Instant{
				At: time.Date(2026, 7, 11, 13, 0, 0, 0, time.Local),
				Offset: calendar.Offset{
					Duration: 165 * time.Minute,
					Icon:     flowbiteOutline.Cart,
					Label:    "Rückkehr",
				},
			},
			Label:     "Mittag",
			Organiser: "Torben",
			Location:  "WZO",
			Lane: calendar.Lane{
				Label: "Torben",
			},
			Category: calendar.Category{
				Label: "Kategorie 2",
				Color: "#ff0000",
			},
			Chips: []calendar.Chip{
				{
					Label:     "Warteliste",
					Icon:      flowbiteSolid.ClipboardList,
					FillColor: txtColor,
					BgColor:   ui.M1,
					TextColor: txtColor,
					Alignment: chipAlignment,
				},
				{
					Label:       "Ausgebucht",
					Icon:        flowbiteOutline.Ban,
					StrokeColor: txtColor,
					BgColor:     "#FE543E",
					TextColor:   txtColor,
					Alignment:   chipAlignment,
				},
			},
		},
		calendar.Event{
			From: calendar.Instant{
				At: time.Date(2026, 7, 11, 12, 0, 0, 0, time.Local),
			},
			To: calendar.Instant{
				At: time.Date(2026, 7, 11, 12, 30, 0, 0, time.Local),
			},
			Label:     "Mittag",
			Organiser: "Olaf",
			Location:  "WZO",
			Lane: calendar.Lane{
				Label: "Torben",
			},
			Category: calendar.Category{
				Label: "Kategorie 2",
				Color: "#ff0000",
			},
		},

		calendar.Event{
			From: calendar.Instant{
				At: time.Date(2026, 7, 11, 13, 0, 0, 0, time.Local),
			},
			To: calendar.Instant{
				At: time.Date(2026, 7, 11, 14, 30, 0, 0, time.Local),
			},
			Label: "Mittagschläfchen",
			Category: calendar.Category{
				Label: "Kategorie 2",
				Color: "#ffff00",
			},
			Chips: []calendar.Chip{
				{
					Label:     "20",
					Icon:      flowbiteSolid.Users,
					FillColor: txtColor,
					BgColor:   ui.M1,
					TextColor: txtColor,
					Alignment: chipAlignment,
				},
				{
					Label:     "Eingetragen",
					Icon:      flowbiteSolid.BadgeCheck,
					FillColor: txtColor,
					BgColor:   "#2BCA73",
					TextColor: txtColor,
					Alignment: ui.BottomLeading,
					FullWidth: true,
				},
			},
		},
	).ViewPort(calendar.Day(2026, 7, 11)).
		FullWidth().
		Style(calendar.StartTimeSequence)
}
