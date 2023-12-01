package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type PID string

type Person struct {
	ID        PID
	Firstname string
}

func (p Person) Identity() PID {
	return p.ID
}

type PetID int

type Pet struct {
	ID   PetID
	Name string
}

func (p Pet) Identity() PetID {
	return p.ID
}

type IntlIke int

type EntByInt struct {
	ID IntlIke
}

func (e EntByInt) Identity() IntlIke {
	return e.ID
}

type MyForm struct {
	Name    ui.TextField
	Check   ui.SwitchField
	DueDate ui.DateField
	Comment ui.TextAreaField
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		cfg.KeycloakAuthentication()
		persons := kv.NewCollection[Person, PID](cfg.Store("test2-db"), "persons")
		err := persons.Save(
			Person{
				ID:        "1",
				Firstname: "Frodo",
			},
			Person{
				ID:        "2",
				Firstname: "Sam",
			},
			Person{
				ID:        "3",
				Firstname: "Pippin",
			},
		)
		if err != nil {
			panic(err)
		}

		pets := kv.NewCollection[Pet, PetID](cfg.Store("test2-db"), "pets")
		pets.Save(
			Pet{
				ID:   1,
				Name: "Katze",
			},
			Pet{
				ID:   2,
				Name: "Hund",
			},
			Pet{
				ID:   3,
				Name: "Esel",
			},
			Pet{
				ID:   4,
				Name: "Stadtmusikant",
			},
		)

		testPet, err2 := pets.Find(3)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println(testPet)

		cfg.Serve(vuejs.Dist())

		counter := 0
		cfg.LivePage("1234", func(w ui.Wire) *ui.LivePage {
			page := ui.NewLivePage(w)
			page.SetBody(
				ui.NewScaffold(func(scaffold *ui.Scaffold) {
					scaffold.TopBar().Left.Set(ui.MakeText("hello app"))
					scaffold.TopBar().Mid.Set(ui.MakeText("GED+DH"))
					scaffold.TopBar().Right.Set(ui.NewButton(func(btn *ui.Button) {
						btn.Caption().Set("user")
					}))
					scaffold.Breadcrumbs().Append(
						ui.NewButton(func(btn *ui.Button) {
							btn.Caption().Set("homer")
						}),
						ui.NewButton(func(btn *ui.Button) {
							btn.Caption().Set("simpson")
							btn.PreIcon().Set(`<svg class="rtl:rotate-180 w-3 h-3 text-gray-400 mx-1" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 6 10">
                <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 9 4-4-4-4"/>
              </svg>`)
						}),
					)

					scaffold.Menu().Append(
						ui.NewButton(func(btn *ui.Button) {
							btn.Caption().Set("hello")
							btn.Action().Set(func() {
								fmt.Println("clicked hello")
							})
						}),
						ui.NewButton(func(btn *ui.Button) {
							btn.Caption().Set("world")
							btn.Action().Set(func() {
								fmt.Println("clicked world")
							})
							btn.PreIcon().Set(`<svg class="" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 22 21">
              <path d="M16.975 11H10V4.025a1 1 0 0 0-1.066-.998 8.5 8.5 0 1 0 9.039 9.039.999.999 0 0 0-1-1.066h.002Z"/>
              <path d="M12.5 0c-.157 0-.311.01-.565.027A1 1 0 0 0 11 1.02V10h8.975a1 1 0 0 0 1-.935c.013-.188.028-.374.028-.565A8.51 8.51 0 0 0 12.5 0Z"/>
            </svg>`)
						}),
					)

					scaffold.Body().Set(
						ui.NewVBox().With(func(vbox *ui.VBox) {
							vbox.Append(
								ui.NewTextField().With(func(t *ui.TextField) {
									t.Label().Set("Vorname")
									t.Hint().Set("dieses Feld ist ohne Fehler")
								}),

								ui.NewTextField().With(func(t *ui.TextField) {
									t.Label().Set("Nachname")
									t.OnTextChanged().Set(func() {
										fmt.Printf("ontext changed to '%v'\n", t.Value().Value())
										t.Error().Set("Malte sagt doch gut: " + t.Value().Get())
										if t.Value().Get() == "magic" {
											vbox.Append(ui.NewTextField().With(func(t *ui.TextField) {
												t.Label().Set("magic field")
												t.Disabled().Set(true)
											}))
										}
									})
								}),

								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("hello world")
									btn.Action().Set(func() {
										counter++
										btn.Caption().Set(fmt.Sprintf("clicked %d", counter))
									})
								}),
								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("primary")
									btn.Style().Set(ui.PrimaryIntent)
								}),
								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("secondary")
									btn.Style().Set(ui.SecondaryIntent)
								}),
								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("destructive")
									btn.Style().Set(ui.Destructive)
								}),

								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("subtile")
									btn.Style().Set(ui.SubtileIntent)
								}),
								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("preicon")
									btn.PreIcon().Set(`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 21">
<path d="M15 12a1 1 0 0 0 .962-.726l2-7A1 1 0 0 0 17 3H3.77L3.175.745A1 1 0 0 0 2.208 0H1a1 1 0 0 0 0 2h.438l.6 2.255v.019l2 7 .746 2.986A3 3 0 1 0 9 17a2.966 2.966 0 0 0-.184-1h2.368c-.118.32-.18.659-.184 1a3 3 0 1 0 3-3H6.78l-.5-2H15Z"/>
</svg>`)
									btn.Style().Set(ui.PrimaryIntent)
								}),

								ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("post")
									btn.PostIcon().Set(`<svg class="rtl:rotate-180" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
</svg>`)
									btn.Style().Set(ui.PrimaryIntent)
								}),

								ui.NewButton(func(btn *ui.Button) {
									btn.PreIcon().Set(`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 18">
<path d="M3 7H1a1 1 0 0 0-1 1v8a2 2 0 0 0 4 0V8a1 1 0 0 0-1-1Zm12.954 0H12l1.558-4.5a1.778 1.778 0 0 0-3.331-1.06A24.859 24.859 0 0 1 6 6.8v9.586h.114C8.223 16.969 11.015 18 13.6 18c1.4 0 1.592-.526 1.88-1.317l2.354-7A2 2 0 0 0 15.954 7Z"/>
</svg>`)
									btn.Style().Set(ui.PrimaryIntent)
								}),

								ui.NewDivider(nil),

								ui.NewHBox(func(box *ui.HBox) {
									box.Append(
										ui.NewButton(func(btn *ui.Button) {
											btn.Style().Set(ui.PrimaryIntent)
											btn.Caption().Set("col 1")
										}),

										ui.NewButton(func(btn *ui.Button) {
											btn.Style().Set(ui.PrimaryIntent)
											btn.Caption().Set("col 2")
										}),

										ui.NewButton(func(btn *ui.Button) {
											btn.Style().Set(ui.PrimaryIntent)
											btn.Caption().Set("col 3")
										}),
									)
								}),

								ui.NewTable(func(table *ui.Table) {

									table.AppendColumns(
										ui.NewTableCell(func(cell *ui.TableCell) {
											cell.Body().Set(ui.NewText(func(text *ui.Text) {
												text.Value().Set("hello world")
												text.Color().Set("text-pink-700")
												text.Size().Set("text-6xl")
												text.OnHoverStart().Set(func() {
													text.Size().Set("40px")
													text.Value().Set("in")
													fmt.Println("in")
												})
												text.OnHoverEnd().Set(func() {
													text.Size().Set("")
													text.Value().Set("out")
													fmt.Println("out")
												})
												text.OnClick().Set(func() {

													fmt.Println("text was clicked")
												})
											}))
										}),
										ui.NewTableCell(func(cell *ui.TableCell) {
											cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
												btn.Caption().Set("column 2")
											}))
										}),
										ui.NewTableCell(func(cell *ui.TableCell) {
											cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
												btn.Caption().Set("column 3")
											}))
										}),
										ui.NewTableCell(func(cell *ui.TableCell) {
											cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
												btn.Caption().Set("column 4")
											}))
										}),
									)

									for i := 0; i < 10; i++ {

										table.AppendRow(
											ui.NewTableRow(func(row *ui.TableRow) {
												for c := 0; c < 4; c++ {
													row.AppendCell(ui.NewTableCell(func(cell *ui.TableCell) {
														cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
															btn.Caption().Set(fmt.Sprintf("row=%d col=%d", i, c))
														}))
													}))
												}
											}),
										)
									}
								}),

								ui.NewGrid(func(grid *ui.Grid) {
									grid.Columns().Set(3)
									grid.Rows().Set(3)
									grid.AppendCells(
										ui.NewGridCell(func(cell *ui.GridCell) {
											cell.ColStart().Set(1)
											cell.ColEnd().Set(3)
											cell.RowStart().Set(1)
											cell.RowEnd().Set(3)
											cell.Body().Set(ui.MakeText("01"))
										}),
										ui.NewGridCell(func(cell *ui.GridCell) {
											cell.Body().Set(ui.MakeText("02"))
										}),
										ui.NewGridCell(func(cell *ui.GridCell) {
											cell.Body().Set(ui.MakeText("03"))
										}),
										ui.NewGridCell(func(cell *ui.GridCell) {
											cell.Body().Set(ui.MakeText("04"))
										}),
										ui.NewGridCell(func(cell *ui.GridCell) {
											cell.Body().Set(ui.MakeText("05"))
										}),
									)
								}),
							)
						}),
					)
				}),
			)

			return page
		})

	}).Run()
}
