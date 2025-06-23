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
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"os"
)

func PageStaging(wnd core.Window, ucImp dataimport.UseCases) core.View {
	sid := dataimport.SID(wnd.Values()["stage"])
	optStage, err := ucImp.FindStagingByID(wnd.Subject(), sid)
	if err != nil {
		return alert.BannerError(err)
	}

	if optStage.IsNone() {
		return alert.BannerError(fmt.Errorf("stage not found: %w", os.ErrNotExist))
	}

	stage := optStage.Unwrap()

	optImp, err := ucImp.FindImporterByID(wnd.Subject(), stage.Importer)
	if err != nil {
		return alert.BannerError(err)
	}

	if optImp.IsNone() {
		return alert.BannerError(fmt.Errorf("importer not found: %w", os.ErrNotExist))
	}

	imp := optImp.Unwrap()

	exampleData := core.AutoState[[]dataimport.Entry](wnd).Init(func() []dataimport.Entry {
		page, err := ucImp.FilterEntries(wnd.Subject(), stage.ID, dataimport.FilterEntriesOptions{
			MaxResults: 3,
		})
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return page.Entries
	})

	deleteStagingPresented := core.AutoState[bool](wnd)

	page := core.AutoState[dataimport.FilterEntriesPage](wnd).Init(func() dataimport.FilterEntriesPage {
		page, err := ucImp.FilterEntries(wnd.Subject(), sid, dataimport.FilterEntriesOptions{})

		if err != nil {
			alert.ShowBannerError(wnd, err)
			return dataimport.FilterEntriesPage{}
		}

		return page
	})

	pageIdx := core.AutoState[int](wnd).Observe(func(newValue int) {
		p, err := ucImp.FilterEntries(wnd.Subject(), sid, dataimport.FilterEntriesOptions{
			Page: newValue,
		})

		if err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}

		page.Set(p)
		page.Notify()
	})

	dlgPresentedFieldMapping := core.AutoState[bool](wnd)
	dlgPresentedImport := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1("Entwürfe prüfen - "+stage.Name),
		dialogDeleteStaging(wnd, deleteStagingPresented, stage, ucImp),
		dialogDoImport(wnd, dlgPresentedImport, stage, page.Get(), pageIdx, ucImp),
		ui.HStack(

			ui.SecondaryButton(func() {
				deleteStagingPresented.Set(true)
			}).Title("Gesamten Entwurf löschen"),

			ui.SecondaryButton(func() {
				dlgPresentedFieldMapping.Set(true)
			}).Title("Felder zuordnen"),

			ui.PrimaryButton(func() {
				dlgPresentedImport.Set(true)
			}).Title("Importieren"),
		).Alignment(ui.Trailing).
			FullWidth().Gap(ui.L8),

		ui.Space(ui.L32),

		DialogFieldMapping(wnd, dlgPresentedFieldMapping, imp, stage, exampleData.Get(), ucImp),
		ViewPage(wnd, imp, stage, ucImp, pageIdx, page.Get()),
	).FullWidth().Alignment(ui.Leading)
}

func dialogDeleteStaging(wnd core.Window, presented *core.State[bool], stage dataimport.Staging, ucImp dataimport.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog("Diesen Entwurf löschen", ui.Text("Soll dieser Entwurf inklusive aller Datensätze gelöscht werden?"), presented, alert.Cancel(nil), alert.Delete(func() {
		if err := ucImp.DeleteStaging(wnd.Subject(), stage.ID); err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}

		wnd.Navigation().BackwardTo("admin/data/stagings", core.Values{"importer": string(stage.Importer)})

	}))
}

func dialogDoImport(wnd core.Window, presented *core.State[bool], stage dataimport.Staging, page dataimport.FilterEntriesPage, pageIdx *core.State[int], ucImp dataimport.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog("Diesen Entwurf importieren", ui.Text(fmt.Sprintf("Soll dieser Entwurf mit %d Einträgen jetzt importiert werden? Abgelehnte und bereits importierte Einträge werden dabei übersprungen.", page.Count)), presented, alert.Cancel(nil), alert.Custom(func(close func(closeDlg bool)) core.View {
		return ui.PrimaryButton(func() {
			close(true)
			defer func() {
				pageIdx.Notify()
				pageIdx.Invalidate()
			}()

			if err := ucImp.Import(wnd.Subject(), stage.ID, stage.Importer, dataimport.ImportOptions{
				Context: wnd.Context(),
				ImporterOptions: importer.Options{
					ContinueOnError: true,
					MergeDuplicates: true,
				},
			}); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			alert.ShowBannerMessage(wnd, alert.Message{
				Title:   "Import erfolgreich",
				Message: "Alle Einträge wurden erfolgreich importiert.",
				Intent:  alert.IntentOk,
			})
		}).Title("Importieren")
	}))
}
