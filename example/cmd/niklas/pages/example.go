package pages

import (
	"fmt"
	"go.wdy.de/nago/presentation/ui"
	"io"
)

func Example(wire ui.Wire) *ui.Page {
	return ui.NewPage(wire, func(page *ui.Page) {
		page.Body().Set(
			ui.NewScaffold(func(scaffold *ui.Scaffold) {
				scaffold.TopBar().Mid.Set(ui.MakeText("moin niklas"))
				scaffold.Body().Set(
					ui.NewVBox(func(vbox *ui.VBox) {
						vbox.Append(
							ui.MakeText("Ein paar Buttons:"),
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("hier klicken")
								btn.Style().Set("primary")
								btn.Action().Set(func() {
									page.Modals().Append(
										ui.NewDialog(func(dlg *ui.Dialog) {
											dlg.Title().Set("Geklickt!")
											dlg.Body().Set(ui.MakeText("Ein Button wurde geklickt 😊"))

											dlg.Actions().Append(
												ui.NewButton(func(btn *ui.Button) {
													btn.Caption().Set("schnell weg")
													btn.Style().Set("destructive")
													btn.Action().Set(func() {
														page.Modals().Remove(dlg)
													})
												}),

												ui.NewButton(func(btn *ui.Button) {
													btn.Caption().Set("Sag Hallo zu Morty!")
													btn.Style().Set("secondary")
													btn.Action().Set(func() {
														page.History().Open("morty", ui.Values{
															"a": "b",
														})
													})
												}),
											)

										}),
									)
								})
							}),
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("hier nicht klicken")
								btn.Style().Set(ui.SubtileIntent)
								btn.Action().Set(func() {
									page.Modals().Append(
										ui.NewDialog(func(dlg *ui.Dialog) {
											dlg.Title().Set(":(")
										}),
									)
								})
							}),
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("Ora Primary Button")
								btn.Style().Set(ui.PrimaryIntent)
								btn.Action().Set(func() {
									page.Modals().Append(
										ui.NewDialog(func(dlg *ui.Dialog) {
											dlg.Title().Set("Ein geklickter Button!")
											dlg.Body().Set(ui.MakeText("Der Ora Primary Button 😊"))

											dlg.Actions().Append(
												ui.NewButton(func(btn *ui.Button) {
													btn.Caption().Set("zurück")
													btn.Style().Set("destructive")
													btn.Action().Set(func() {
														page.Modals().Remove(dlg)
													})
												}),
											)
										}),
									)
								})
							}),
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("Ora Secondary Button")
								btn.Style().Set(ui.SecondaryIntent)
								btn.Action().Set(func() {
									page.Modals().Append(
										ui.NewDialog(func(dlg *ui.Dialog) {
											dlg.Title().Set("Geklickt!")
											dlg.Body().Set(ui.MakeText("Der Ora Secondary Button 🤩"))

											dlg.Actions().Append(
												ui.NewButton(func(btn *ui.Button) {
													btn.Caption().Set("zurück")
													btn.Style().Set("destructive")
													btn.Action().Set(func() {
														page.Modals().Remove(dlg)
													})
												}),
											)
										}),
									)
								})
							}),
							ui.NewButton(func(btn *ui.Button) {
								btn.Caption().Set("Ora Tertiary Button")
								btn.Style().Set(ui.TertiaryIntent)
								btn.Action().Set(func() {
									page.Modals().Append(
										ui.NewDialog(func(dlg *ui.Dialog) {
											dlg.Title().Set("Geklickt!")
											dlg.Body().Set(ui.MakeText("Der Ora Tertiary Button 🥳"))

											dlg.Actions().Append(
												ui.NewButton(func(btn *ui.Button) {
													btn.Caption().Set("zurück")
													btn.Style().Set("destructive")
													btn.Action().Set(func() {
														page.Modals().Remove(dlg)
													})
												}),
											)
										}),
									)
								})
							}),
							ui.MakeText("Eine Tabelle:"),
							ui.NewTable(func(table *ui.Table) {
								table.Header().Append(
									ui.NewTableCell(func(cell *ui.TableCell) {
										cell.Body().Set(ui.MakeText("Gut"))
									}),
									ui.NewTableCell(func(cell *ui.TableCell) {
										cell.Body().Set(ui.MakeText("Auch gut"))
									}),
									ui.NewTableCell(func(cell *ui.TableCell) {
										cell.Body().Set(ui.MakeText("Ebenfalls gut"))
									}),
									ui.NewTableCell(func(cell *ui.TableCell) {
										cell.Body().Set(ui.MakeText("Ebenso gut"))
									}),
								)

								table.Rows().From(func(yield func(row *ui.TableRow)) {
									yield(ui.NewTableRow(func(row *ui.TableRow) {
										row.Cells().Append(
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("eher nicht gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("nicht so gut"))
											}),
										)
									}))
									yield(ui.NewTableRow(func(row *ui.TableRow) {
										row.Cells().Append(
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("gut"))
											}),
											ui.NewTableCell(func(cell *ui.TableCell) {
												cell.Body().Set(ui.MakeText("auch nicht gut"))
											}),
										)
									}))

								})

							}),

							ui.NewFileField(func(fileField *ui.FileField) {
								fileField.Label().Set("Dateien zum Upload einfügen")
								fileField.Hint().Set("Klick oder drag 'n' drop zum Upload")
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
												dlg.Body().Set(ui.MakeText("Die Datei ist sicher angekommen: " + file.Name()))
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
						)
					}),
				)
			}),
		)
	})
}
