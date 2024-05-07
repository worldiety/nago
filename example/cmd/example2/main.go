package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
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
		cfg.SetApplicationID("de.worldiety.nago.demo.kitchensink")
		//cfg.KeycloakAuthentication()
		persons := application.SloppyRepository[Person, PID](cfg)
		err := persons.SaveAll(
			slices.Values([]Person{
				{
					ID:        "1",
					Firstname: "Frodo",
				},
				{
					ID:        "2",
					Firstname: "Sam",
				},
				{
					ID:        "3",
					Firstname: "Pippin",
				},
			}),
		)
		if err != nil {
			panic(err)
		}

		pets := application.SloppyRepository[Pet, PetID](cfg)
		err = pets.SaveAll(
			slices.Values([]Pet{
				{
					ID:   1,
					Name: "Katze",
				},
				{
					ID:   2,
					Name: "Hund",
				},
				{
					ID:   3,
					Name: "Esel",
				},
				{
					ID:   4,
					Name: "Stadtmusikant",
				},
			}),
		)

		testPet, err2 := pets.FindByID(3)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println(testPet)

		cfg.Serve(vuejs.Dist())

		cfg.Component("hello", func(wnd core.Window) core.Component {
			return ui.NewPage(func(page *ui.Page) {

				type myParams struct {
					A int    `name:"a"`
					B string `name:"b"`
				}
				test, _ := core.UnmarshalValues[myParams](wnd.Values())
				page.Body().Set(
					ui.NewVBox(func(vbox *ui.VBox) {
						vbox.Append(
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("zurück")
								btn.Action().Set(func() {
									wnd.Navigation().Back()
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
		cfg.Component("1234", func(wnd core.Window) core.Component {
			logging.FromContext(wnd.Context()).Info("user", slog.Any("user", wnd.User()), slog.String("session", string(wnd.SessionID())))

			page := ui.NewPage(nil)
			page.Body().Set(
				ui.NewScaffold(func(scaffold *ui.Scaffold) {

					scaffold.NavigationComponent().Set(
						ui.NewNavigationComponent(func(navigationComponent *ui.NavigationComponent) {
							var menuEntryA *ui.MenuEntry
							var menuEntryB *ui.MenuEntry
							var menuEntryC *ui.MenuEntry

							navigationComponent.Alignment().Set(ora.AlignmentLeft)
							navigationComponent.Logo().Set(icon.OraLogo)
							navigationComponent.Menu().Append(ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
								menuEntryA = menuEntry
								menuEntry.Title().Set("Menüpunkt A")
								menuEntry.Icon().Set(icon.PackageOutlined)
								menuEntry.IconActive().Set(icon.PackageFilled)
								menuEntry.Action().Set(func() {
									wnd.Navigation().ForwardTo("hello", map[string]string{"menu_entry": "sub_1"})
								})
								menuEntry.OnFocus().Set(func() {
									menuEntryB.Expanded().Set(false)
									menuEntryC.Expanded().Set(false)
								})
								menuEntry.Menu().Append(ui.NewMenuEntry(func(subEntry *ui.MenuEntry) {
									subEntry.Title().Set("Subpunkt 1")
									subEntry.Action().Set(func() {
										wnd.Navigation().ForwardTo("hello", map[string]string{"menu_entry": "sub_1"})
									})
								}))
							}))
							navigationComponent.Menu().Append(ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
								menuEntryB = menuEntry
								menuEntry.Title().Set("Ich bin ein sehr langer Menüpunkt B")
								menuEntry.Icon().Set(icon.PackageOutlined)
								menuEntry.IconActive().Set(icon.PackageFilled)
								menuEntry.OnFocus().Set(func() {
									menuEntryA.Expanded().Set(false)
									menuEntryC.Expanded().Set(false)
								})
								menuEntry.Menu().Append(ui.NewMenuEntry(func(subEntry *ui.MenuEntry) {
									subEntry.Title().Set("Subpunkt 1")
								}))
								menuEntry.Menu().Append(ui.NewMenuEntry(func(subEntry *ui.MenuEntry) {
									subEntry.Title().Set("Subpunkt 2")
									subEntry.Action().Set(func() {
										wnd.Navigation().ForwardTo("hello", map[string]string{"menu_entry": "sub_2"})
									})
									subEntry.Menu().Append(ui.NewMenuEntry(func(subSubEntry *ui.MenuEntry) {
										subSubEntry.Title().Set("Subsubpunkt I")
										subSubEntry.Action().Set(func() {
											wnd.Navigation().ForwardTo("hello", map[string]string{"menu_entry": "subsub_I"})
										})
									}))
									subEntry.Menu().Append(ui.NewMenuEntry(func(subSubEntry *ui.MenuEntry) {
										subSubEntry.Title().Set("Subsubpunkt II")
									}))
								}))
								menuEntry.Menu().Append(ui.NewMenuEntry(func(subEntry *ui.MenuEntry) {
									subEntry.Title().Set("Subpunkt 3")
									subEntry.Menu().Append(ui.NewMenuEntry(func(subSubEntry *ui.MenuEntry) {
										subSubEntry.Title().Set("Subsubpunkt III")
										subSubEntry.Action().Set(func() {
											wnd.Navigation().ForwardTo("hello", map[string]string{"menu_entry": "subsub_III"})
										})
									}))
								}))
							}))
							navigationComponent.Menu().Append(ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
								menuEntryC = menuEntry
								menuEntry.Title().Set("Menüpunkt C")
								menuEntry.Icon().Set(icon.PackageOutlined)
								menuEntry.IconActive().Set(icon.PackageFilled)
								menuEntry.OnFocus().Set(func() {
									menuEntryA.Expanded().Set(false)
									menuEntryB.Expanded().Set(false)
								})
								menuEntry.Action().Set(func() {
									wnd.Navigation().ForwardTo("hello", map[string]string{"menu_entry": "C"})
								})
							}))
						}),
					)

					var myMagicTF *ui.TextField
					scaffold.Body().Set(
						ui.NewVBox(func(vbox *ui.VBox) {

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
								var currentStartValue = 15.1
								var currentEndValue = 32.46

								slider.Label().Set("Slider")
								slider.Hint().Set("Das ist ein Hinweis")
								slider.RangeMode().Set(false)
								slider.Min().Set(-4.43)
								slider.Max().Set(91.05)
								slider.StartValue().Set(currentStartValue)
								slider.EndValue().Set(currentEndValue)
								slider.Stepsize().Set(2.17)
								slider.StartInitialized().Set(true)
								slider.EndInitialized().Set(false)
								slider.ShowLabel().Set(true)
								slider.LabelSuffix().Set(" €")
								slider.ShowTickMarks().Set(true)
								slider.OnChanged().Set(func() {
									if slider.StartValue().Get() != currentStartValue {
										slider.StartInitialized().Set(true)
									}
									if slider.EndValue().Get() != currentEndValue {
										slider.EndInitialized().Set(true)
									}
									currentStartValue = slider.StartValue().Get()
									currentEndValue = slider.EndValue().Get()
								})
							}))

							vbox.Append(ui.NewDatepicker(func(datepicker *ui.Datepicker) {
								datepicker.Label().Set("Datepicker-Label")
								datepicker.OnClicked().Set(func() {
									datepicker.Expanded().Set(true)
								})
								datepicker.RangeMode().Set(true)
								datepicker.OnSelectionChanged().Set(func() {
									fmt.Println("changed date")
									datepicker.Expanded().Set(false)
								})
							}))

							vbox.Append(ui.NewDropdown(func(dropdown *ui.Dropdown) {
								dropdown.Multiselect().Set(true)
								dropdown.Expanded().Set(false)
								dropdown.Searchable().Set(true)
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
										item.Content().Set("Option BCD")
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

							vbox.Append(ui.MakeText(string(wnd.User().UserID()) + ":" + wnd.User().Name() + "->" + string(wnd.User().Email())))

							vbox.Append(
								ui.NewTextField(func(t *ui.TextField) {
									t.Simple().Set(false)
									t.Label().Set("Vorname")
									t.Help().Set("Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext.")
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

								ui.NewPasswordField(func(p *ui.PasswordField) {
									p.Simple().Set(false)
									p.Disabled().Set(false)
									p.Hint().Set("Optional")
									p.Label().Set("Passwort")
									p.Help().Set("Das ist ein kurzer Hilfstext.")
									p.Placeholder().Set("Bitte eingeben...")
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

										table.Rows().From(func(yield func(*ui.TableRow) bool) {
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
