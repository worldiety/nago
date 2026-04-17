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
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
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
						AttendeeState: &calendar.Chip{
							Label:     "Eingetragen",
							Icon:      flowbiteSolid.BadgeCheck,
							BgColor:   "#2BCA73",
							TextColor: ui.M8,
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
						AttendeeState: &calendar.Chip{
							Label:     "Wartelistenplatz: 5",
							Icon:      flowbiteSolid.ClipboardList,
							BgColor:   "#FBC83E",
							TextColor: ui.M8,
						},
					},

					calendar.Event{
						From: calendar.Instant{
							At: time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local),
							Offset: calendar.Offset{
								Label:    "Anfahrt",
								Icon:     flowbiteOutline.Bell,
								Duration: time.Hour * 24 * 3,
							},
						},
						To: calendar.Instant{
							At: time.Date(2025, 8, 31, 0, 0, 0, 0, time.Local),
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
								BgColor:   "#3A3257",
								TextColor: ui.M9,
							},
						},
					},
				).
					ViewPort(calendar.Year(2025)).
					Frame(ui.Frame{}.FullWidth()).
					Style(calendar.StartTimeSequence),

				//
				calStartTimeSeqTimeExample(),
			).FullWidth().Gap(ui.L16)

		})
	}).Run()
}

func calStartTimeSeqTimeExample() core.View {
	return calendar.Calendar(
		calendar.Event{
			From: calendar.Instant{
				At: time.Date(2026, 7, 11, 12, 0, 0, 0, time.Local),
			},
			To: calendar.Instant{
				At: time.Date(2026, 7, 11, 13, 0, 0, 0, time.Local),
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
					BgColor:   "#3A3257",
					TextColor: ui.M9,
				},
				{
					Label:     "Ausgebucht",
					Icon:      heroOutline.XMark,
					BgColor:   "#FE543E",
					TextColor: ui.M8,
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
					BgColor:   "#3A3257",
					TextColor: ui.M9,
				},
			},
			AttendeeState: &calendar.Chip{
				Label:     "Eingetragen",
				Icon:      flowbiteSolid.BadgeCheck,
				BgColor:   "#4FEE62",
				TextColor: ui.M8,
			},
		},
	).ViewPort(calendar.Day(2026, 7, 11)).
		FullWidth().
		Style(calendar.StartTimeSequence)
}
