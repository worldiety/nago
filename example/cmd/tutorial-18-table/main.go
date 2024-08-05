package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return VStack(
				Table(
					TableColumn(Text("col 1")).
						Width(L20),
					TableColumn(Text("col 2")).
						Action(func() {
							fmt.Println("clicked header cell 2")
						}).
						HoveredBackgroundColor("I0"),
					TableColumn(Text("col 3")),
				).Rows(
					TableRow(
						TableCell(Text("row 1 col 1")).
							Action(func() {
								fmt.Println("clicked cell 1/1")
							}).
							HoveredBackgroundColor("I0"),
						TableCell(Text("row 1 col 2")).BackgroundColor("I0"),
						TableCell(Text("row 1 col 3")),
					).Action(func() {
						fmt.Println("clicked row 1")
					}),
					TableRow(
						TableCell(Text("row 2 col 1")),
						TableCell(Text("row 2 col 2")),
						TableCell(Text("row 2 col 3")),
					).
						Height(L80).
						BackgroundColor("M2").
						HoveredBackgroundColor("I0"),
					TableRow(
						TableCell(Text("row 3 col 1")),
						TableCell(Text("row 3 col 2+3").Color("#ffffff")).
							BackgroundColor("A0").
							Alignment(Center).
							ColSpan(2),
					),
					TableRow(
						TableCell(Text("row 4+5 col 1")).
							RowSpan(2).
							Border(Border{}.
								Color("M0").
								Width(L1)),
						TableCell(Text("row 4 col 2")),
						TableCell(Text("row 4 col 3")),
					),
					TableRow(
						TableCell(Text("row 5 col 2")),
						TableCell(Text("row 5 col 3")),
					),
				).BackgroundColor("#ffffff").
					CellPadding(Padding{}.Horizontal(L24).Vertical(L16)).
					Frame(Frame{Width: L480}).
					RowDividerColor("M5").
					Border(Border{}.Radius(L20)),

				Text("cell alignments"),
				Table().Rows(
					TableRow(
						TableCell(Text("top-leading")).
							Alignment(TopLeading),
						TableCell(Text("top")).
							Alignment(Top),
						TableCell(Text("top-trailing")).
							Alignment(TopTrailing),
					).Height(L80),

					TableRow(
						TableCell(Text("leading")).
							Alignment(Leading),
						TableCell(Text("center")).
							Alignment(Center),
						TableCell(Text("trailing")).
							Alignment(Trailing),
					).Height(L80),

					TableRow(
						TableCell(Text("bottom-leading")).
							Alignment(BottomLeading),
						TableCell(Text("bottom")).
							Alignment(Bottom),
						TableCell(Text("Bottom-trailing")).
							Alignment(BottomTrailing),
					).Height(L80),
				).Frame(Frame{Width: L480}).
					RowDividerColor("#000000").
					BackgroundColor("I0"),
			).BackgroundColor("M3").
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
