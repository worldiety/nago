package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"io"
	"log/slog"
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

//go:embed example.jpeg
var exampleImg []byte

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		//cfg.KeycloakAuthentication()
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

		cfg.Page("hello", func(wire ui.Wire) *ui.Page {
			return ui.NewPage(wire, func(page *ui.Page) {

				type myParams struct {
					A int    `name:"a"`
					B string `name:"b"`
				}
				test, _ := ui.UnmarshalValues[myParams](wire.Values())
				page.Body().Set(
					ui.NewVBox(func(vbox *ui.VBox) {
						vbox.Append(
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("zurück")
								btn.Action().Set(func() {
									page.History().Back()
								})
							}),

							ui.MakeText(fmt.Sprintf("A=%v", test.A)),
							ui.MakeText(fmt.Sprintf("B=%v", test.B)),
						)
					}),
				)
			})
		})

		counter := 0
		cfg.Page("1234", func(w ui.Wire) *ui.Page {
			logging.FromContext(w.Context()).Info("user", slog.Any("user", w.User()))
			logging.FromContext(w.Context()).Info("remote", slog.String("addr", w.Remote().Addr()), slog.String("forwd", w.Remote().ForwardedFor()))

			page := ui.NewPage(w, nil)
			page.Body().Set(
				ui.NewScaffold(func(scaffold *ui.Scaffold) {
					scaffold.TopBar().Left.Set(ui.MakeText("hello app"))
					scaffold.TopBar().Mid.Set(ui.MakeText("GED+DH"))
					scaffold.TopBar().Right.Set(ui.NewButton(func(btn *ui.Button) {
						btn.Caption().Set("gehe zu")
						btn.Action().Set(func() {
							page.Modals().Append(
								ui.NewDialog(func(dlg *ui.Dialog) {
									dlg.Title().Set("super dialog")
									dlg.Icon().Set(`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75l3 3m0 0l3-3m-3 3v-7.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
</svg>
`)
									dlg.Body().Set(ui.MakeText("hello super text"))
									dlg.Actions().Append(
										ui.NewButton(func(btn *ui.Button) {
											btn.Caption().Set("schließen")
											btn.Action().Set(func() {
												page.Modals().Remove(dlg)
											})
										}),

										ui.NewButton(func(btn *ui.Button) {
											btn.Caption().Set("öffnen")
											btn.Style().Set("destructive")
											btn.Action().Set(func() {
												page.History().Open("hello", ui.Values{
													"a": "1234",
													"b": "456",
												})
											})
										}),

										ui.NewButton(func(btn *ui.Button) {
											btn.Caption().Set("ugly stack it")
											btn.Action().Set(func() {
												page.Modals().Append(
													ui.NewDialog(func(dlg *ui.Dialog) {
														dlg.Title().Set("got you over it")
													}),
												)
											})
										}),
									)
								}),
							)

						})
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

					var yieldToggleVal bool
					var myMagicTF *ui.TextField
					scaffold.Body().Set(
						ui.NewVBox(func(vbox *ui.VBox) {
							vbox.Append(ui.NewHBox(func(hBox *ui.HBox) {
								hBox.Children().From(func(yield func(ui.LiveComponent)) {
									yield(ui.NewToggle(func(tgl *ui.Toggle) {
										tgl.Label().Set("Ein Toggle.")
										tgl.Checked().Set(yieldToggleVal)
										tgl.OnCheckedChanged().Set(func() {
											yieldToggleVal = tgl.Checked().Get()
											fmt.Println("yield toggle to", tgl.Checked().Get())
										})
									}))
								})
							}))

							vbox.Append(ui.NewNumberField(func(numberField *ui.NumberField) {
								numberField.Value().Set(123)
								numberField.Simple().Set(true)
								numberField.Label().Set("Nummernfeld für Ganzzahlen")
								numberField.Placeholder().Set("Bitte eine Ganzzahl eingeben...")
								numberField.OnValueChanged().Set(func() {
									if numberField.Value().Get() == 3 {
										numberField.Error().Set("Wert darf nicht 3 sein")
									} else {
										numberField.Error().Set("")
									}
								})
							}))

							vbox.Append(ui.NewSlider(func(slider *ui.Slider) {
								slider.Label().Set("Slider")
								slider.Hint().Set("Das ist ein Hinweis")
								slider.Min().Set(-10)
								slider.Max().Set(25)
								slider.Value().Set(21.49)
								slider.Stepsize().Set(.75)
								slider.Initialized().Set(true)
								slider.OnChanged().Set(func() {
									slider.Initialized().Set(true)
								})
							}))

							vbox.Append(ui.NewDatepicker(func(datepicker *ui.Datepicker) {
								datepicker.Label().Set("Datepicker-Label")
								datepicker.Error().Set("Das ist auch eine Fehlermeldung")
								datepicker.OnClicked().Set(func() {
									datepicker.Expanded().Set(!datepicker.Expanded().Get())
								})
								datepicker.SelectedDay().Set(7)
								datepicker.SelectedMonthIndex().Set(2)
								datepicker.SelectedYear().Set(2024)
							}))

							vbox.Append(ui.NewDropdown(func(dropdown *ui.Dropdown) {
								dropdown.Multiselect().Set(true)
								dropdown.Expanded().Set(false)
								dropdown.Label().Set("Multiselect")
								dropdown.Error().Set("Das ist eine Fehlermeldung")
								dropdown.Hint().Set("Das ist ein Hinweis")
								dropdown.OnClicked().Set(func() {
									dropdown.Expanded().Set(!dropdown.Expanded().Get())
								})

								dropdown.Items().Append(
									ui.NewDropdownItem(func(item *ui.DropdownItem) {
										item.Content().Set("Option A")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									ui.NewDropdownItem(func(item *ui.DropdownItem) {
										item.Content().Set("Option BC")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									ui.NewDropdownItem(func(item *ui.DropdownItem) {
										item.Content().Set("Option DEF")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),
								)
							}))

							vbox.Append(ui.NewDropdown(func(dropdown *ui.Dropdown) {
								dropdown.Multiselect().Set(false)
								dropdown.Expanded().Set(false)
								dropdown.Label().Set("Dropdown")
								dropdown.Hint().Set("Das ist ein anderer Hinweis")
								dropdown.OnClicked().Set(func() {
									dropdown.Expanded().Set(!dropdown.Expanded().Get())
								})

								dropdown.Items().Append(
									ui.NewDropdownItem(func(item *ui.DropdownItem) {
										item.Content().Set("Option G")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									ui.NewDropdownItem(func(item *ui.DropdownItem) {
										item.Content().Set("Option HI")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									ui.NewDropdownItem(func(item *ui.DropdownItem) {
										item.Content().Set("Option JKL")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),
								)
							}))

							vbox.Append(ui.MakeText(w.User().UserID() + ":" + w.User().Name() + "->" + w.User().Email()))

							vbox.Append(
								ui.NewTextField(func(t *ui.TextField) {
									t.Simple().Set(false)
									t.Error().Set("Fehler")
									t.Label().Set("Vorname")
									t.Placeholder().Set("Bitte eingeben...")
									t.Hint().Set("dieses Feld ist ohne Fehler")
								}),

								ui.NewTextField(func(t *ui.TextField) {
									t.Simple().Set(true)
									myMagicTF = t
									t.Label().Set("Nachname")
									t.OnTextChanged().Set(func() {
										fmt.Printf("ontext changed to '%v'\n", t.Value().Get())
										t.Error().Set("Malte sagt doch gut: " + t.Value().Get())
										if t.Value().Get() == "magic" {
											vbox.Append(ui.NewTextField(func(t *ui.TextField) {
												t.Label().Set("magic field")
												t.Disabled().Set(true)
											}))
										}
									})
								}),

								ui.NewToggle(func(tgl *ui.Toggle) {

									tgl.Label().Set("anschalten")
									tgl.Checked().Set(false)
									//	tgl.Disabled().Set(true)
									tgl.OnCheckedChanged().Set(func() {
										fmt.Println("toggle changed to ", tgl.Checked().Get())
										myMagicTF.Disabled().Set(tgl.Checked().Get())
									})
								}),

								ui.NewFileField(func(fileField *ui.FileField) {
									fileField.Label().Set("Dein Zeug zum upload")
									fileField.Hint().Set("Klick or Drag'n drop zum Upload")
									//fileField.Accept().Set(".gif")
									fileField.Multiple().Set(true)
									fileField.OnUploadReceived(func(files []ui.FileUpload) {
										for _, file := range files {
											f, _ := file.Open()
											defer f.Close()
											buf, _ := io.ReadAll(f)
											fmt.Println(file.Name(), file.Size(), len(buf))
											page.Modals().Append(
												ui.NewDialog(func(dlg *ui.Dialog) {
													dlg.Title().Set("hey")
													dlg.Body().Set(ui.MakeText("hello Alex, die Datei ist sicher angekommen: " + file.Name()))
													dlg.Actions().Append(
														ui.NewButton(func(btn *ui.Button) {
															btn.Caption().Set("ganz toll")
															btn.Action().Set(func() {
																page.Modals().Remove(dlg)
															},
															)
														}),
													)
												}),
											)
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
									btn.Caption().Set("tertiary")
									btn.Style().Set(ui.TertiaryIntent)
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

									table.Header().Append(
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

										table.Rows().From(func(yield func(*ui.TableRow)) {
											for c := 0; c < 4; c++ {
												yield(ui.NewTableRow(func(row *ui.TableRow) {
													for c := 0; c < 4; c++ {
														row.Cells().Append(ui.NewTableCell(func(cell *ui.TableCell) {
															cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
																btn.Caption().Set(fmt.Sprintf("row=%d col=%d", i, c))
																btn.Action().Set(func() {
																	fmt.Println("clicked it")
																})
															}))
														}))
													}
												}))
											}
										})
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

								ui.MakeText("response grid:"),
								ui.NewDivider(nil),
								ui.NewGrid(func(grid *ui.Grid) {
									grid.Columns().Set(1)
									grid.ColumnsSmallOrLarger().Set(2)
									grid.ColumnsMediumOrLarger().Set(3)
									grid.AppendCells(
										ui.NewGridCell(func(cell *ui.GridCell) {
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

								ui.NewStepper(func(stepper *ui.Stepper) {
									stepper.SelectedIndex().Set(1)

									stepper.Steps().Append(
										ui.NewStepInfo(func(step *ui.StepInfo) {
											step.Number().Set("1")
											step.Caption().Set("First step")
											step.Details().Set("comes always first")
										}),

										ui.NewStepInfo(func(step *ui.StepInfo) {
											step.Number().Set("2")
											step.Caption().Set("seconds step")
											step.Details().Set("comes always second")
										}),

										ui.NewStepInfo(func(step *ui.StepInfo) {
											step.Number().Set("33")
											step.Caption().Set("third step")
											step.Details().Set("comes always last")
										}),
									)
								}),

								ui.NewTextArea(func(textArea *ui.TextArea) {
									textArea.Label().Set("dein Roman")
									textArea.Value().Set("lorem")
									textArea.Hint().Set("egal was du machst")
									textArea.Rows().Set(10)
									textArea.OnTextChanged().Set(func() {
										textArea.Error().Set("dein Fehler: " + textArea.Value().Get())
									})
								}),

								ui.NewImage(func(img *ui.Image) {
									img.URL().Set("https://www.worldiety.de/_nuxt/images/news_wzo_einzug2-bc96a5.webp")
								}),

								ui.NewImage(func(img *ui.Image) {
									img.Source(func() (io.Reader, error) {
										return bytes.NewBuffer(exampleImg), nil
									})
								}),

								ui.NewHBox(func(hBox *ui.HBox) {
									hBox.Alignment().Set("flex-center")
									hBox.Append(
										ui.NewChip(func(chip *ui.Chip) {
											chip.Caption().Set("default")
											chip.Action().Set(func() {
												fmt.Println("chip click")
											})
											chip.OnClose().Set(func() {
												hBox.Children().Remove(chip)
											})

										}),

										ui.NewChip(func(chip *ui.Chip) {
											chip.Caption().Set("red")
											chip.Color().Set("red")
										}),
										ui.NewChip(func(chip *ui.Chip) {
											chip.Caption().Set("green")
											chip.Color().Set("green")
										}),
										ui.NewChip(func(chip *ui.Chip) {
											chip.Caption().Set("yellow")
											chip.Color().Set("yellow")
										}),

										ui.NewCard(func(card *ui.Card) {
											card.Append(
												ui.NewVBox(func(vbox *ui.VBox) {
													vbox.Append(
														ui.NewText(func(text *ui.Text) {
															text.Color().Set("#FF0000")
															text.Size().Set("2xl")
															text.Value().Set("Super card")
														}),
														ui.MakeText("standard text kram"),
														ui.NewDivider(nil),
														ui.MakeText("bblabla"),
													)
												}),
											)
										}),
									)
								}),
							)

							vbox.Append(ui.NewWebView(func(view *ui.WebView) {
								view.Value().Set(`
<h1>hello <em>world</em></h1>
<p>a paragraph</p>

`)
							}))
						}),
					)
				}),
			)

			return page
		})

	}).Run()
}
