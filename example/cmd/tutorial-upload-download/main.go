package main

import (
	"bytes"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"io/fs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		fstore := cfg.FileStore("my-files")
		must(any(nil), blob.Write(fstore, "bla.txt", bytes.NewBufferString("blub")))

		cfg.Component(".", func(wnd core.Window) core.Component {
			return ui.NewVBox(func(vbox *ui.VBox) {
				vbox.Append(
					// configure the upload field
					ui.NewFileField(func(fileField *ui.FileField) {
						fileField.Multiple().Set(true)
						fileField.SetFilesReceiver(func(fsys fs.FS) error {
							return blob.Import(fstore, fsys)
						})
					}),
					ui.NewTable(func(table *ui.Table) {
						table.Header().Append(ui.NewTextCell("File"), ui.NewTextCell("Options"))
						table.Rows().From(func(yield func(*ui.TableRow) bool) {
							keys := must(blob.Keys(fstore))

							for _, key := range keys {
								yield(ui.NewTableRow(func(row *ui.TableRow) {
									row.Cells().Append(
										ui.NewTextCell(key),
										ui.NewTableCell(func(cell *ui.TableCell) {
											cell.Body().Set(ui.NewHBox(func(hbox *ui.HBox) {
												hbox.Append(
													ui.NewButton(func(btn *ui.Button) {
														btn.Caption().Set("delete")
														btn.Action().Set(func() {
															must(any(nil), blob.Delete(fstore, key))
														})
													}),
													ui.NewButton(func(btn *ui.Button) {
														btn.Caption().Set("download")
														btn.Action().Set(func() {
															must(any(nil), wnd.SendFiles(blob.NewFS(blob.Filter(fstore, namePredicate(key)))))
														})
													}),
												)
											}))
										}),
									)
								}))
							}

						})
					}),

					ui.NewButton(func(btn *ui.Button) {
						btn.Caption().Set("download all")
						btn.Action().Set(func() {
							must(any(nil), wnd.SendFiles(blob.NewFS(fstore)))
						})
					}),
				)
			})

		})
	}).Run()
}

// don't do this in production, check and handle gracefully, give help and apply logging
func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func namePredicate(name string) func(blob.Entry) bool {
	return func(entry blob.Entry) bool {
		return entry.Key == name
	}
}
