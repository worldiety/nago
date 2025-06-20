// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"fmt"
	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/hero"
	"os"
)

func PageSelectParser(wnd core.Window, ucImp dataimport.UseCases) core.View {
	id := importer.ID(wnd.Values()["importer"])
	optImp, err := ucImp.FindImporterByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optImp.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	imp := optImp.Unwrap()

	parsers, err := xslices.Collect2(ucImp.FindParsers(wnd.Subject()))
	if err != nil {
		return alert.BannerError(err)
	}

	pendingStaging := core.AutoState[dataimport.SID](wnd)
	presentImportDone := core.AutoState[bool](wnd)

	return ui.VStack(
		hero.Hero(imp.Configuration().Name+" - Datenquellen auswählen").
			Subtitle(fmt.Sprintf("Es stehen insgesamt %d Datenquellen zur Verfügung mit denen Daten aus externen Quellen in einen Import-Entwurf übernommen werden können. "+
				"Nach dem Parsen der Daten stehen die ausgelesenen Daten als Entwurf zur weiteren Kontrolle und Bearbeitung bereit bevor sie zum Import übergeben werden.", len(parsers))).
			Teaser(ui.ImageIcon(imp.Configuration().Image)),
		dialogDataImported(wnd, presentImportDone, pendingStaging.Get(), imp.Identity()),
		ui.Space(ui.L32),
		cardlayout.Layout(
			ui.ForEach(parsers, func(p parser.Parser) core.View {
				cfg := p.Configuration()
				var btnAction string
				var withUpload bool
				if cfg.FromUpload.Enabled {
					btnAction = "Datei(en) hochladen"
					withUpload = true
				} else if cfg.FromBuildIn.Enabled {
					btnAction = "Import auslösen"
				}

				return cardlayout.Card(cfg.Name).Body(
					ui.VStack(
						ui.ImageIcon(cfg.Image).Frame(ui.Frame{}.Size(ui.L200, ui.L200)),
						ui.Text(cfg.Description),
					),
				).Footer(ui.SecondaryButton(func() {
					if pendingStaging.Get() == "" {
						// security note: do not extend this by accepting existing staging id by query
						// because our use cases have no security concept, which otherwise protects against
						// inserting malicious data in foreign staging data sets.
						staging, err := ucImp.CreateStaging(wnd.Subject(), dataimport.StagingCreationData{
							Name:     "Mein Import Entwurf",
							Comment:  "",
							Importer: imp.Identity(),
						})
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						pendingStaging.Set(staging.ID)
					}

					if withUpload {
						wnd.ImportFiles(core.ImportFilesOptions{
							Multiple:         true,
							MaxBytes:         cfg.FromUpload.MaxUploadSize,
							AllowedMimeTypes: cfg.FromUpload.MimeTypes,
							OnCompletion: func(files []core.File) {
								// note, that we get into trouble, when navigating to early, because at the backend side, we don't know how many files will be sent by the web frontend,
								// but it is not wrong either, let us call this a 'feature'. Let us show a dialog and let the user decide, if a navigation
								// shall be done.
								// If we navigate, this completion handler is removed and we won't receiver any files after the first one.
								for _, file := range files {
									reader, err := file.Open()
									if err != nil {
										alert.ShowBannerError(wnd, err)
										return
									}

									defer reader.Close()

									stats, err := ucImp.Parse(wnd.Subject(), pendingStaging.Get(), p.Identity(), parser.Options{}, reader)
									if err != nil {
										alert.ShowBannerError(wnd, err)
										return
									}

									alert.ShowBannerMessage(wnd, alert.Message{
										Title:   fmt.Sprintf("%d Einträge gespeichert.", stats.Count),
										Message: fmt.Sprintf("Die Datei %s wurde verarbeitet.", file.Name()),
										Intent:  alert.IntentOk,
									})
								}

								presentImportDone.Set(true)
							},
						})
					} else {
						stats, err := ucImp.Parse(wnd.Subject(), pendingStaging.Get(), p.Identity(), parser.Options{}, nil)
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						alert.ShowBannerMessage(wnd, alert.Message{
							Title:   fmt.Sprintf("%d Einträge als Entwürfe gespeichert", stats.Count),
							Message: "Der Parsing-Prozess ist abgeschlossen.",
							Intent:  alert.IntentOk,
						})

						presentImportDone.Set(true)
					}
				}).Title(btnAction))
			})...,
		),
	).Alignment(ui.Leading).
		FullWidth()
}

func dialogDataImported(wnd core.Window, presented *core.State[bool], staging dataimport.SID, importer importer.ID) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog("Daten importiert", ui.Text("Die Datei wurde verarbeitet."),
		presented,
		alert.Larger(),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.SecondaryButton(func() {
				close(true)
			}).Title("Weitere Daten hinzufügen")
		}),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo("admin/data/staging", core.Values{"stage": string(staging)})
			}).Title("Import-Entwürfe ansehen")
		}),
	)
}
