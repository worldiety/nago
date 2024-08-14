package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"go.wdy.de/nago/web/vuejs"
	"io"
	"iter"
	"slices"
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

func generateNavigationComponent(wnd core.Window) *uilegacy.NavigationComponent {
	return uilegacy.NewNavigationComponent(func(navigationComponent *uilegacy.NavigationComponent) {
		navigationComponent.Alignment().Set(ora.AlignmentLeft)
		navigationComponent.Logo().Set(icon.OraLogo)
		navigationComponent.Menu().Append(uilegacy.NewMenuEntry(func(menuEntry *uilegacy.MenuEntry) {
			menuEntry.Title().Set("Menüpunkt A")
			menuEntry.Icon().Set(icon.PackageOutlined)
			menuEntry.IconActive().Set(icon.PackageFilled)
			menuEntry.Badge().Set("2")
			menuEntry.Menu().Append(uilegacy.NewMenuEntry(func(subEntry *uilegacy.MenuEntry) {
				subEntry.Title().Set("Subpunkt 1")
				subEntry.Link("hello", wnd, map[string]string{"menu_entry": "sub_1"})
			}))
		}))
		navigationComponent.Menu().Append(uilegacy.NewMenuEntry(func(menuEntry *uilegacy.MenuEntry) {
			menuEntry.Title().Set("Ich bin ein sehr langer Menüpunkt B")
			menuEntry.Icon().Set(icon.PackageOutlined)
			menuEntry.IconActive().Set(icon.PackageFilled)
			menuEntry.Link("1234", wnd, map[string]string{"menu_entry": "B"})
			menuEntry.Menu().Append(uilegacy.NewMenuEntry(func(subEntry *uilegacy.MenuEntry) {
				subEntry.Title().Set("Subpunkt 1")
			}))
			menuEntry.Menu().Append(uilegacy.NewMenuEntry(func(subEntry *uilegacy.MenuEntry) {
				subEntry.Title().Set("Subpunkt 2")
				subEntry.Link("hello", wnd, map[string]string{"menu_entry": "sub_2"})
				subEntry.Menu().Append(uilegacy.NewMenuEntry(func(subSubEntry *uilegacy.MenuEntry) {
					subSubEntry.Title().Set("Subsubpunkt I")
					subSubEntry.Link("hello", wnd, map[string]string{"menu_entry": "subsub_I"})
				}))
				subEntry.Menu().Append(uilegacy.NewMenuEntry(func(subSubEntry *uilegacy.MenuEntry) {
					subSubEntry.Title().Set("Subsubpunkt II")
				}))
			}))
			menuEntry.Menu().Append(uilegacy.NewMenuEntry(func(subEntry *uilegacy.MenuEntry) {
				subEntry.Title().Set("Subpunkt 3")
				subEntry.Menu().Append(uilegacy.NewMenuEntry(func(subSubEntry *uilegacy.MenuEntry) {
					subSubEntry.Title().Set("Subsubpunkt III")
					subSubEntry.Link("hello", wnd, map[string]string{"menu_entry": "subsub_III"})
				}))
			}))
		}))
		navigationComponent.Menu().Append(uilegacy.NewMenuEntry(func(menuEntry *uilegacy.MenuEntry) {
			menuEntry.Title().Set("Menüpunkt C")
			menuEntry.Icon().Set(icon.PackageOutlined)
			menuEntry.IconActive().Set(icon.PackageFilled)
			menuEntry.Link("hello", wnd, map[string]string{"menu_entry": "C"})
		}))
	})
}

//go:embed example.jpeg
var exampleImg []byte

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetThemes(ora.Themes{
			Dark: ora.GenerateTheme(
				ora.PrimaryColor(ora.MustParseHSL("#F7A823")),
				ora.SecondaryColor(ora.MustParseHSL("#00FF00")),
				ora.TertiaryColor(ora.MustParseHSL("#0000FF")),
			),
			Light: ora.GenerateTheme(
				ora.PrimaryColor(ora.MustParseHSL("#F7A823")),
				ora.SecondaryColor(ora.MustParseHSL("#00FF00")),
				ora.TertiaryColor(ora.MustParseHSL("#0000FF")),
			),
		})

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

		cfg.Component("hello", func(wnd core.Window) core.View {
			return uilegacy.NewPage(func(page *uilegacy.Page) {

				type myParams struct {
					A int    `name:"a"`
					B string `name:"b"`
				}
				test, _ := core.UnmarshalValues[myParams](wnd.Values())
				page.Body().Set(
					uilegacy.NewScaffold(func(scaffold *uilegacy.Scaffold) {
						scaffold.NavigationComponent().Set(generateNavigationComponent(wnd))

						scaffold.Body().Set(
							uilegacy.NewVBox(func(vbox *uilegacy.VBox) {
								vbox.Append(
									uilegacy.NewButton(func(btn *uilegacy.Button) {
										btn.Caption().Set("zurück")
										btn.Action().Set(func() {
											wnd.Navigation().Back()
										})
									}),

									uilegacy.MakeText(fmt.Sprintf("A=%v", test.A)),
									uilegacy.MakeText(fmt.Sprintf("B=%v", test.B)),
								)
							}),
						)
					}),
				)
			})
		})

		counter := 0
		cfg.Component("1234", func(wnd core.Window) core.View {
			// logging.FromContext(wnd.Context()).Info("user", slog.Any("user", wnd.User()), slog.String("session", string(wnd.SessionID())))

			page := uilegacy.NewPage(nil)

			page.Body().Set(
				uilegacy.NewScaffold(func(scaffold *uilegacy.Scaffold) {

					scaffold.NavigationComponent().Set(generateNavigationComponent(wnd))

					var myMagicTF *uilegacy.TextField
					scaffold.Body().Set(
						uilegacy.NewVBox(func(vbox *uilegacy.VBox) {

							vbox.Append(uilegacy.NewBreadcrumbs(func(breadcrumbs *uilegacy.Breadcrumbs) {
								breadcrumbs.Items().Append(uilegacy.NewBreadcrumbItem(func(item *uilegacy.BreadcrumbItem) {
									item.Label().Set("Punkt A")
								}))
								breadcrumbs.Items().Append(uilegacy.NewBreadcrumbItem(func(item *uilegacy.BreadcrumbItem) {
									item.Label().Set("Ich bin ein unnötig langer Breadcrumb-Punkt, der dafür sorgen soll, dass die Breadcrumbs umbrechen")
									item.Action().Set(func() {
										println("Breadcrumb item clicked")
									})
								}))
								breadcrumbs.Items().Append(uilegacy.NewBreadcrumbItem(func(item *uilegacy.BreadcrumbItem) {
									item.Label().Set("Punkt C")
								}))
								breadcrumbs.Items().Append(uilegacy.NewBreadcrumbItem(func(item *uilegacy.BreadcrumbItem) {
									item.Label().Set("Punkt D")
								}))
								breadcrumbs.SelectedItemIndex().Set(1)
								breadcrumbs.Icon().Set(icon.Dashboard)
							}))

							vbox.Append(uilegacy.NewProgressBar(func(progressBar *uilegacy.ProgressBar) {
								progressBar.Max().Set(113.43)
								progressBar.Value().Set(47.32)
							}))

							vbox.Append(uilegacy.NewFlexContainer(func(flexContainer *uilegacy.FlexContainer) {
								flexContainer.Orientation().Set(ora.OrientationHorizontal)
								flexContainer.ElementSize().Set(ora.ElementSizeLarge)
								flexContainer.ContentAlignment().Set(ora.ContentStart)

								flexContainer.Elements().Append(uilegacy.NewCard(func(card *uilegacy.Card) {
									card.Append(uilegacy.NewFlexContainer(func(innerFlexContainer *uilegacy.FlexContainer) {
										innerFlexContainer.Orientation().Set(ora.OrientationHorizontal)
										innerFlexContainer.ElementSize().Set(ora.ElementSizeAuto)
										innerFlexContainer.ContentAlignment().Set(ora.ContentBetween)
										innerFlexContainer.ItemsAlignment().Set(ora.ItemsCenter)

										innerFlexContainer.Elements().Append(uilegacy.NewText(func(text *uilegacy.Text) {
											text.Value().Set("Ein Text")
										}))
										innerFlexContainer.Elements().Append(uilegacy.NewButton(func(button *uilegacy.Button) {
											button.Caption().Set("Button A2")
											button.Style().Set(uilegacy.SecondaryIntent)
											button.Action().Set(func() {
												page.Modals().Append(

													uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
														dlg.Title().Set("Ein Dialog")
														dlg.Icon().Set(icon.ExclamationTriangle)
														dlg.Size().Set(ora.ElementSizeMedium)
														dlg.Body().Set(uilegacy.NewFlexContainer(func(flex *uilegacy.FlexContainer) {
															flex.Orientation().Set(ora.OrientationVertical)
															flex.ContentAlignment().Set(ora.ContentStart)
															flex.ItemsAlignment().Set(ora.ItemsStart)

															flex.Elements().Append(uilegacy.NewText(func(text *uilegacy.Text) {
																text.Value().Set("Ein Text")
															}))

															flex.Elements().Append(uilegacy.NewDropdown(func(dropdown *uilegacy.Dropdown) {
																dropdown.Multiselect().Set(true)
																dropdown.OnClicked().Set(func() {
																	dropdown.Expanded().Set(!dropdown.Expanded().Get())
																})

																dropdown.Items().Append(uilegacy.NewDropdownItem(func(dropdownItem *uilegacy.DropdownItem) {
																	dropdownItem.Content().Set("Item A Item A Item A Item A Item A Item A Item A Item A Item A")
																	dropdownItem.OnClicked().Set(func() {
																		dropdown.Toggle(dropdownItem)
																	})
																}))
																dropdown.Items().Append(uilegacy.NewDropdownItem(func(dropdownItem *uilegacy.DropdownItem) {
																	dropdownItem.Content().Set("Item B Item B Item B Item B Item B Item B Item B Item B Item B Item B")
																	dropdownItem.OnClicked().Set(func() {
																		dropdown.Toggle(dropdownItem)
																	})
																}))
															}))

															flex.Elements().Append(uilegacy.NewDatepicker(func(datepicker *uilegacy.Datepicker) {
																datepicker.RangeMode().Set(true)
																datepicker.Label().Set("Ein Datepicker")
																datepicker.OnClicked().Set(func() {
																	datepicker.Expanded().Set(true)
																})
																datepicker.OnSelectionChanged().Set(func() {
																	fmt.Println("changed date")
																	datepicker.Expanded().Set(false)
																})
															}))
														}))
														dlg.Footer().Set(uilegacy.NewFlexContainer(func(flex *uilegacy.FlexContainer) {
															flex.ContentAlignment().Set(ora.ContentBetween)

															flex.Elements().Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
																btn.Caption().Set("Schließen")
																btn.Style().Set(uilegacy.SecondaryIntent)
																btn.Action().Set(func() {
																	page.Modals().Remove(dlg)
																})
															}))
															flex.Elements().Append(uilegacy.NewFlexContainer(func(flex *uilegacy.FlexContainer) {
																flex.ContentAlignment().Set(ora.ContentEnd)

																flex.Elements().Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
																	btn.Caption().Set("Behalten")
																	btn.Style().Set(uilegacy.SecondaryIntent)
																	btn.Action().Set(func() {
																		page.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
																			dlg.Title().Set("Sekundärer Dialog")
																			dlg.Size().Set(ora.ElementSizeSmall)
																			dlg.Body().Set(uilegacy.NewText(func(text *uilegacy.Text) {
																				text.Value().Set("Mit einfachem Inhalt")
																			}))
																			dlg.Footer().Set(uilegacy.NewFlexContainer(func(flex *uilegacy.FlexContainer) {
																				flex.ContentAlignment().Set(ora.ContentEnd)
																				flex.ItemsAlignment().Set(ora.ItemsCenter)
																				flex.Elements().Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
																					btn.Caption().Set("Schließen")
																					btn.Action().Set(func() {
																						page.Modals().Remove(dlg)
																					})
																				}))
																			}))
																		}))
																	})
																}))
																flex.Elements().Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
																	btn.Caption().Set("Löschen")
																	btn.Style().Set(uilegacy.Destructive)
																}))
															}))
														}))
													}),
												)
											})
										}))
									}))
								}))
								flexContainer.Elements().Append(uilegacy.NewCard(func(card *uilegacy.Card) {
									card.Append(uilegacy.NewButton(func(button *uilegacy.Button) {
										button.Caption().Set("Button B")
										button.Style().Set(uilegacy.PrimaryIntent)
									}))
								}))
								flexContainer.Elements().Append(uilegacy.NewCard(func(card *uilegacy.Card) {
									card.Append(uilegacy.NewButton(func(button *uilegacy.Button) {
										button.Caption().Set("Button C")
										button.Style().Set(uilegacy.SecondaryIntent)
									}))
								}))
								flexContainer.Elements().Append(uilegacy.NewCard(func(card *uilegacy.Card) {
									card.Append(uilegacy.NewButton(func(button *uilegacy.Button) {
										button.Caption().Set("Button D")
										button.Style().Set(uilegacy.TertiaryIntent)
									}))
								}))
								flexContainer.Elements().Append(uilegacy.NewCard(func(card *uilegacy.Card) {
									card.Append(uilegacy.NewButton(func(button *uilegacy.Button) {
										button.Caption().Set("Button E")
										button.Style().Set(uilegacy.Destructive)
									}))
								}))
							}))

							vbox.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
								btn.Caption().Set("reload page")
								btn.Action().Set(func() {
									wnd.Navigation().Reload()
								})
							}))

							vbox.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
								btn.Caption().Set("download")
								btn.Action().Set(func() {
									err := wnd.SendFiles(core.FilesIter(mem.From(mem.Entries{
										"test.txt": []byte("hello world"),
									})))
									xdialog.ErrorView("send files failed", err)
								})
							}))

							vbox.Append(uilegacy.NewFileField(func(fileField *uilegacy.FileField) {
								fileField.Label().Set("Drag & Drop oder Dateien per Klick auswählen")
								fileField.HintRight().Set("Max. Dateigröße: 300 MB")
								fileField.MaxBytes().Set(300000000000000) // 300 MB
								fileField.HintLeft().Set("Unterstützte Dateiformate: MP4, PDF, JPG, DOCX")
								fileField.Accept().Set("video/mp4,image/jpeg,image/png,application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document")
								//fileField.Accept().Set(".gif")
								fileField.Multiple().Set(true)

								fileField.SetFilesReceiver(func(it iter.Seq2[core.File, error]) error {
									defer core.Release(it)
									var err error
									it(func(file core.File, e error) bool {
										f, err := file.Open()
										if err != nil {
											err = e
											return false
										}
										defer f.Close()
										buf, _ := io.ReadAll(f)

										fmt.Println(file.Name(), len(buf))
										return true
									})

									if err != nil {
										xdialog.ErrorView("err", err)
									}

									return nil
								})

							}))

							var toggle *uilegacy.Toggle
							vbox.Append(uilegacy.NewToggle(func(tgl *uilegacy.Toggle) {
								toggle = tgl
								tgl.Label().Set("Toggle 1")
								tgl.Checked().Set(false)
								//	tgl.Disabled().Set(true)
								tgl.OnCheckedChanged().Set(func() {
									fmt.Println("toggle 1 changed to ", tgl.Checked().Get())
								})
							}))

							vbox.Append(uilegacy.NewToggle(func(tgl *uilegacy.Toggle) {

								tgl.Label().Set("Toggle 2")
								tgl.Checked().Set(false)
								//	tgl.Disabled().Set(true)
								tgl.OnCheckedChanged().Set(func() {
									toggle.Checked().Set(false)
									toggle.OnCheckedChanged().Invoke()
									fmt.Println("toggle 2 changed to ", tgl.Checked().Get())
								})
							}))

							var pf *uilegacy.PasswordField
							vbox.Append(uilegacy.NewPasswordField(func(passwordField *uilegacy.PasswordField) {
								passwordField.Simple().Set(true)
								passwordField.Label().Set("Passwort 1")
								passwordField.Placeholder().Set("Bitte ein Passwort eingeben...")
								passwordField.OnPasswordChanged().Set(func() {
									pf.Value().Set(passwordField.Value().Get())
								})
							}))

							vbox.Append(uilegacy.NewPasswordField(func(passwordField *uilegacy.PasswordField) {
								pf = passwordField
								passwordField.Simple().Set(true)
								passwordField.Label().Set("Passwort 2")
								passwordField.Placeholder().Set("Bitte ein Passwort eingeben...")
							}))

							var nf *uilegacy.NumberField
							vbox.Append(uilegacy.NewNumberField(func(numberField *uilegacy.NumberField) {
								numberField.Value().Set("123")
								numberField.Simple().Set(true)
								numberField.Label().Set("Nummernfeld für Ganzzahlen")
								numberField.Placeholder().Set("Bitte eine Ganzzahl eingeben...")
								numberField.OnValueChanged().Set(func() {
									nf.Value().Set(numberField.Value().Get())
								})
							}))

							vbox.Append(uilegacy.NewNumberField(func(numberField *uilegacy.NumberField) {
								nf = numberField
								numberField.Value().Set("123")
								numberField.Simple().Set(true)
								numberField.Label().Set("Nummernfeld für Ganzzahlen 2")
								numberField.Placeholder().Set("Bitte eine Ganzzahl eingeben...")
							}))

							vbox.Append(uilegacy.NewSlider(func(slider *uilegacy.Slider) {
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

							vbox.Append(uilegacy.NewDatepicker(func(datepicker *uilegacy.Datepicker) {
								datepicker.Label().Set("Datepicker-Label")

								datepicker.SelectedStartDay().Set(1)
								datepicker.SelectedStartMonth().Set(3)
								datepicker.SelectedStartYear().Set(2023)
								datepicker.StartDateSelected().Set(true)

								datepicker.SelectedEndDay().Set(25)
								datepicker.SelectedEndMonth().Set(10)
								datepicker.SelectedEndYear().Set(2023)
								datepicker.EndDateSelected().Set(true)

								datepicker.OnClicked().Set(func() {
									fmt.Println("clicked datepicker")
								})
								datepicker.RangeMode().Set(true)
								datepicker.OnSelectionChanged().Set(func() {
									fmt.Println("changed date")
								})
							}))

							vbox.Append(uilegacy.NewDropdown(func(dropdown *uilegacy.Dropdown) {
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
									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Halle A 07:00 Uhr Stand 1")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Halle A 07:00 Uhr Stand 2")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Halle B 07:00 Uhr Stand 1")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),
									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Halle B 08:00 Uhr Stand 2")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),
								)
							}))

							vbox.Append(uilegacy.NewDropdown(func(dropdown *uilegacy.Dropdown) {
								dropdown.Multiselect().Set(false)
								dropdown.Expanded().Set(false)
								dropdown.Label().Set("Dropdown")
								dropdown.Hint().Set("Das ist ein anderer Hinweis")
								dropdown.OnClicked().Set(func() {
									dropdown.Expanded().Set(!dropdown.Expanded().Get())
								})

								dropdown.Items().Append(
									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Option G")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Option HI")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),

									uilegacy.NewDropdownItem(func(item *uilegacy.DropdownItem) {
										item.Content().Set("Option JKL")
										item.OnClicked().Set(func() {
											dropdown.Toggle(item)
										})
									}),
								)
							}))

							// vbox.Append(ui.MakeText(string(wnd.User().UserID()) + ":" + wnd.User().Name() + "->" + string(wnd.User().Email())))

							var otherCheckbox *uilegacy.Checkbox
							vbox.Append(
								uilegacy.NewTextField(func(t *uilegacy.TextField) {
									t.Simple().Set(false)
									t.Label().Set("Vorname")
									t.Help().Set("Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext. Das ist ein Hilfstext.")
									t.Placeholder().Set("Bitte eingeben...")
									t.Hint().Set("dieses Feld ist ohne Fehler")
									t.OnTextChanged().Set(func() {
										myMagicTF.Value().Set(t.Value().Get())
										myMagicTF.OnTextChanged().Invoke()
									})
								}),

								uilegacy.NewTextField(func(t *uilegacy.TextField) {
									t.Simple().Set(true)
									myMagicTF = t
									t.Label().Set("Nachname")
									t.OnTextChanged().Set(func() {
										fmt.Printf("ontext changed to '%v'\n", t.Value().Get())
										t.Error().Set("Malte sagt doch gut: " + t.Value().Get())
										if t.Value().Get() == "magic" {
											vbox.Append(uilegacy.NewTextField(func(t *uilegacy.TextField) {
												t.Label().Set("magic field")
												t.Disabled().Set(true)
											}))
										}
									})
								}),

								uilegacy.NewToggle(func(tgl *uilegacy.Toggle) {

									tgl.Label().Set("anschalten")
									tgl.Checked().Set(false)
									//	tgl.Disabled().Set(true)
									tgl.OnCheckedChanged().Set(func() {
										fmt.Println("toggle changed to ", tgl.Checked().Get())
										myMagicTF.Disabled().Set(tgl.Checked().Get())
									})
								}),

								uilegacy.NewTable(func(table *uilegacy.Table) {
									table.Rows().Append(uilegacy.NewTableRow(func(row *uilegacy.TableRow) {
										row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewCheckbox(func(chb *uilegacy.Checkbox) {
												chb.OnClicked().Set(func() {
													fmt.Println("Hallo aus Checkbox")
													if otherCheckbox.Selected().Get() == true {
														otherCheckbox.Selected().Set(false)
													}
												})

											}))
										}))
										row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewCheckbox(func(chb *uilegacy.Checkbox) {
												otherCheckbox = chb
											}))
										}))
									}))
									table.Rows().Append(uilegacy.NewTableRow(func(row *uilegacy.TableRow) {
										var radiobuttons []*uilegacy.Radiobutton
										var selectedButton *uilegacy.Radiobutton

										row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewRadiobutton(func(rab *uilegacy.Radiobutton) {
												radiobuttons = append(radiobuttons, rab)
												rab.OnClicked().Set(func() {
													selectedButton = rab
													fmt.Println("radiobutton 1 changed to", rab.Selected().Get())
													rab.UpdateRadioButtons(radiobuttons, selectedButton)
												})
											}))
										}))
										row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewRadiobutton(func(rab *uilegacy.Radiobutton) {
												radiobuttons = append(radiobuttons, rab)
												rab.OnClicked().Set(func() {
													selectedButton = rab
													fmt.Println("radiobutton 2 changed to", rab.Selected().Get())
													rab.UpdateRadioButtons(radiobuttons, selectedButton)
												})
											}))
										}))
									}))
								}),

								uilegacy.NewRadiobutton(func(rab *uilegacy.Radiobutton) {

								}),

								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("hello world")
									btn.Action().Set(func() {
										counter++
										btn.Caption().Set(fmt.Sprintf("clicked %d", counter))
									})
								}),
								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("primary")
									btn.Style().Set(uilegacy.PrimaryIntent)
								}),
								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("secondary")
									btn.Style().Set(uilegacy.SecondaryIntent)
								}),
								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("tertiary")
									btn.Style().Set(uilegacy.TertiaryIntent)
								}),
								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("destructive")
									btn.Style().Set(uilegacy.Destructive)
								}),

								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("subtile")
									btn.Style().Set(uilegacy.SubtileIntent)
								}),
								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("preicon")
									btn.PreIcon().Set(`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 21">
<path d="M15 12a1 1 0 0 0 .962-.726l2-7A1 1 0 0 0 17 3H3.77L3.175.745A1 1 0 0 0 2.208 0H1a1 1 0 0 0 0 2h.438l.6 2.255v.019l2 7 .746 2.986A3 3 0 1 0 9 17a2.966 2.966 0 0 0-.184-1h2.368c-.118.32-.18.659-.184 1a3 3 0 1 0 3-3H6.78l-.5-2H15Z"/>
</svg>`)
									btn.Style().Set(uilegacy.PrimaryIntent)
								}),

								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("post")
									btn.PostIcon().Set(`<svg class="rtl:rotate-180" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
</svg>`)
									btn.Style().Set(uilegacy.PrimaryIntent)
								}),

								uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.PreIcon().Set(`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 18">
<path d="M3 7H1a1 1 0 0 0-1 1v8a2 2 0 0 0 4 0V8a1 1 0 0 0-1-1Zm12.954 0H12l1.558-4.5a1.778 1.778 0 0 0-3.331-1.06A24.859 24.859 0 0 1 6 6.8v9.586h.114C8.223 16.969 11.015 18 13.6 18c1.4 0 1.592-.526 1.88-1.317l2.354-7A2 2 0 0 0 15.954 7Z"/>
</svg>`)
									btn.Style().Set(uilegacy.PrimaryIntent)
								}),

								uilegacy.NewDivider(nil),

								uilegacy.NewHBox(func(box *uilegacy.HBox) {
									box.Append(
										uilegacy.NewButton(func(btn *uilegacy.Button) {
											btn.Style().Set(uilegacy.PrimaryIntent)
											btn.Caption().Set("col 1")
										}),

										uilegacy.NewButton(func(btn *uilegacy.Button) {
											btn.Style().Set(uilegacy.PrimaryIntent)
											btn.Caption().Set("col 2")
										}),

										uilegacy.NewButton(func(btn *uilegacy.Button) {
											btn.Style().Set(uilegacy.PrimaryIntent)
											btn.Caption().Set("col 3")
										}),
									)
								}),

								uilegacy.NewTable(func(table *uilegacy.Table) {

									table.Header().Append(
										uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewText(func(text *uilegacy.Text) {
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
										uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewButton(func(btn *uilegacy.Button) {
												btn.Caption().Set("column 2")
											}))
										}),
										uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewButton(func(btn *uilegacy.Button) {
												btn.Caption().Set("column 3")
											}))
										}),
										uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewButton(func(btn *uilegacy.Button) {
												btn.Caption().Set("column 4")
											}))
										}),
									)

									for i := 0; i < 10; i++ {

										table.Rows().From(func(yield func(*uilegacy.TableRow) bool) {
											for c := 0; c < 4; c++ {
												yield(uilegacy.NewTableRow(func(row *uilegacy.TableRow) {
													for c := 0; c < 4; c++ {
														row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
															cell.Body().Set(uilegacy.NewButton(func(btn *uilegacy.Button) {
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

								uilegacy.NewGrid(func(grid *uilegacy.Grid) {
									grid.Columns().Set(3)
									grid.Rows().Set(3)
									grid.AppendCells(
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.ColStart().Set(1)
											cell.ColEnd().Set(3)
											cell.RowStart().Set(1)
											cell.RowEnd().Set(3)
											cell.Body().Set(uilegacy.MakeText("01"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("02"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("03"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("04"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("05"))
										}),
									)
								}),

								uilegacy.MakeText("response grid:"),
								uilegacy.NewDivider(nil),
								uilegacy.NewGrid(func(grid *uilegacy.Grid) {
									grid.Columns().Set(1)
									grid.ColumnsSmallOrLarger().Set(2)
									grid.ColumnsMediumOrLarger().Set(3)
									grid.AppendCells(
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("01"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("02"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("03"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("04"))
										}),
										uilegacy.NewGridCell(func(cell *uilegacy.GridCell) {
											cell.Body().Set(uilegacy.MakeText("05"))
										}),
									)
								}),

								uilegacy.NewStepper(func(stepper *uilegacy.Stepper) {
									stepper.SelectedIndex().Set(1)

									stepper.Steps().Append(
										uilegacy.NewStepInfo(func(step *uilegacy.StepInfo) {
											step.Number().Set("1")
											step.Caption().Set("First step")
											step.Details().Set("comes always first")
										}),

										uilegacy.NewStepInfo(func(step *uilegacy.StepInfo) {
											step.Number().Set("2")
											step.Caption().Set("seconds step")
											step.Details().Set("comes always second")
										}),

										uilegacy.NewStepInfo(func(step *uilegacy.StepInfo) {
											step.Number().Set("33")
											step.Caption().Set("third step")
											step.Details().Set("comes always last")
										}),
									)
								}),

								uilegacy.NewTextArea(func(textArea *uilegacy.TextArea) {
									textArea.Label().Set("dein Roman")
									textArea.Value().Set("lorem")
									textArea.Hint().Set("egal was du machst")
									textArea.Rows().Set(10)
									textArea.OnTextChanged().Set(func() {
										textArea.Error().Set("dein Fehler: " + textArea.Value().Get())
									})
								}),

								uilegacy.NewHBox(func(hBox *uilegacy.HBox) {
									hBox.Alignment().Set("flex-center")
									hBox.Append(
										uilegacy.NewChip(func(chip *uilegacy.Chip) {
											chip.Caption().Set("default")
											chip.Action().Set(func() {
												fmt.Println("chip click")
											})
											chip.OnClose().Set(func() {
												hBox.Children().Remove(chip)
											})

										}),

										uilegacy.NewChip(func(chip *uilegacy.Chip) {
											chip.Caption().Set("red")
											chip.Color().Set("red")
										}),
										uilegacy.NewChip(func(chip *uilegacy.Chip) {
											chip.Caption().Set("green")
											chip.Color().Set("green")
										}),
										uilegacy.NewChip(func(chip *uilegacy.Chip) {
											chip.Caption().Set("yellow")
											chip.Color().Set("yellow")
										}),

										uilegacy.NewCard(func(card *uilegacy.Card) {
											card.Append(
												uilegacy.NewVBox(func(vbox *uilegacy.VBox) {
													vbox.Append(
														uilegacy.NewText(func(text *uilegacy.Text) {
															text.Color().Set("#FF0000")
															text.Size().Set("2xl")
															text.Value().Set("Super card")
														}),
														uilegacy.MakeText("standard text kram"),
														uilegacy.NewDivider(nil),
														uilegacy.MakeText("bblabla"),
													)
												}),
											)
										}),
									)
								}),
							)

							vbox.Append(uilegacy.NewWebView(func(view *uilegacy.WebView) {
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
