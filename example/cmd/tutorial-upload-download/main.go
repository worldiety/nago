package main

import (
	"bytes"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		fstore := cfg.FileStore("my-files")
		must(any(nil), blob.Write(fstore, "bla.txt", bytes.NewBufferString("blub")))

		cfg.Component(".", func(wnd core.Window) core.View {
			return uilegacy.NewVBox(func(vbox *uilegacy.VBox) {
				vbox.Append(
					// configure the upload field
					uilegacy.NewFileField(func(fileField *uilegacy.FileField) {
						fileField.Multiple().Set(true)
						fileField.SetFilesReceiver(application.BlobReceiver(fstore))
					}),
					uilegacy.NewTable(func(table *uilegacy.Table) {
						table.Header().Append(uilegacy.NewTextCell("File"), uilegacy.NewTextCell("Options"))
						table.Rows().From(func(yield func(*uilegacy.TableRow) bool) {
							keys := must(blob.Keys(fstore))

							for _, key := range keys {
								yield(uilegacy.NewTableRow(func(row *uilegacy.TableRow) {
									row.Cells().Append(
										uilegacy.NewTextCell(key),
										uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
											cell.Body().Set(uilegacy.NewHBox(func(hbox *uilegacy.HBox) {
												hbox.Append(
													uilegacy.NewButton(func(btn *uilegacy.Button) {
														btn.Caption().Set("delete")
														btn.Action().Set(func() {
															must(any(nil), blob.Delete(fstore, key))
														})
													}),
													uilegacy.NewButton(func(btn *uilegacy.Button) {
														btn.Caption().Set("download")
														btn.Action().Set(func() {
															must(any(nil), wnd.SendFiles(core.FilesIter(blob.Filter(fstore, namePredicate(key)))))
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

					uilegacy.NewButton(func(btn *uilegacy.Button) {
						btn.Caption().Set("download all")
						btn.Action().Set(func() {
							must(any(nil), wnd.SendFiles(core.FilesIter(fstore)))
						})
					}),
				)
			})

		})
	}).Run()
}

// don't do this in production, check and handle gracefully, give help and apply logging
// this is just used to keep the code above short for demonstration
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
