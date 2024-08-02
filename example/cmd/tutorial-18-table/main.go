package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Table(
					ui.TableColumn(ui.Text("col 1")).
						Width(ora.L20),
					ui.TableColumn(ui.Text("col 2")).
						Action(func() {
							fmt.Println("clicked header cell 2")
						}).
						HoveredBackgroundColor("I0"),
					ui.TableColumn(ui.Text("col 3")),
				).Rows(
					ui.TableRow(
						ui.TableCell(ui.Text("row 1 col 1")).
							Action(func() {
								fmt.Println("clicked cell 1/1")
							}).
							HoveredBackgroundColor("I0"),
						ui.TableCell(ui.Text("row 1 col 2")).BackgroundColor("I0"),
						ui.TableCell(ui.Text("row 1 col 3")),
					).Action(func() {
						fmt.Println("clicked row 1")
					}),
					ui.TableRow(
						ui.TableCell(ui.Text("row 2 col 1")),
						ui.TableCell(ui.Text("row 2 col 2")),
						ui.TableCell(ui.Text("row 2 col 3")),
					).
						Height(ora.L80).
						BackgroundColor("M2").
						HoveredBackgroundColor("I0"),
					ui.TableRow(
						ui.TableCell(ui.Text("row 3 col 1")),
						ui.TableCell(ui.Text("row 3 col 2+3").Color("#ffffff")).
							BackgroundColor("A0").
							Alignment(ora.Center).
							ColSpan(2),
					),
					ui.TableRow(
						ui.TableCell(ui.Text("row 4+5 col 1")).
							RowSpan(2).
							Border(ora.Border{}.
								Color("M0").
								Width(ora.L1)),
						ui.TableCell(ui.Text("row 4 col 2")),
						ui.TableCell(ui.Text("row 4 col 3")),
					),
					ui.TableRow(
						ui.TableCell(ui.Text("row 5 col 2")),
						ui.TableCell(ui.Text("row 5 col 3")),
					),
				).BackgroundColor("#ffffff").
					CellPadding(ora.Padding{}.Horizontal(ora.L24).Vertical(ora.L16)).
					Frame(ora.Frame{Width: ora.L480}).
					RowDividerColor("M5").
					Border(ora.Border{}.Radius(ora.L20)),

				ui.Text("cell alignments"),
				ui.Table().Rows(
					ui.TableRow(
						ui.TableCell(ui.Text("top-leading")).
							Alignment(ora.TopLeading),
						ui.TableCell(ui.Text("top")).
							Alignment(ora.Top),
						ui.TableCell(ui.Text("top-trailing")).
							Alignment(ora.TopTrailing),
					).Height(ora.L80),

					ui.TableRow(
						ui.TableCell(ui.Text("leading")).
							Alignment(ora.Leading),
						ui.TableCell(ui.Text("center")).
							Alignment(ora.Center),
						ui.TableCell(ui.Text("trailing")).
							Alignment(ora.Trailing),
					).Height(ora.L80),

					ui.TableRow(
						ui.TableCell(ui.Text("bottom-leading")).
							Alignment(ora.BottomLeading),
						ui.TableCell(ui.Text("bottom")).
							Alignment(ora.Bottom),
						ui.TableCell(ui.Text("Bottom-trailing")).
							Alignment(ora.BottomTrailing),
					).Height(ora.L80),
				).Frame(ora.Frame{Width: ora.L480}).
					RowDividerColor("#000000").
					BackgroundColor("I0"),
			).BackgroundColor("M3").
				Frame(ora.Frame{}.MatchScreen())
		})
	}).Run()
}
