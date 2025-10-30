// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"os"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrDocuments          = i18n.MustString("nago.ai.admin.library.documents", i18n.Values{language.English: "Documents", language.German: "Dokumente"})
	StrDocument           = i18n.MustString("nago.ai.admin.library.document", i18n.Values{language.English: "Document", language.German: "Dokument"})
	StrLibraryUpdated     = i18n.MustString("nago.ai.admin.library.updated", i18n.Values{language.English: "Library updated", language.German: "Bibliothek aktualisiert"})
	StrLibraryUpdatedDesc = i18n.MustString("nago.ai.admin.library.updated_desc", i18n.Values{language.English: "The Library has been updated.", language.German: "Die Bibliothek wurde erfolgreich aktualisiert."})
)

func PageLibrary(wnd core.Window, uc ai.UseCases) core.View {
	optProv, err := uc.FindProviderByID(wnd.Subject(), provider.ID(wnd.Values()["provider"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optProv.IsNone() {
		return alert.BannerError(fmt.Errorf("provider not found: %s: %w", wnd.Values()["provider"], os.ErrNotExist))
	}

	prov := optProv.Unwrap()
	optLibs := prov.Libraries()
	if optLibs.IsNone() {
		return alert.BannerError(fmt.Errorf("provider does not support libraries: %s", wnd.Values()["provider"]))
	}

	libs := optLibs.Unwrap()
	libID := library.ID(wnd.Values()["library"])
	lib := core.AutoState[library.Library](wnd).AsyncInit(func() library.Library {
		optLib, err := libs.FindByID(wnd.Subject(), libID)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return library.Library{}
		}

		if optLib.IsNone() {
			alert.ShowBannerError(wnd, fmt.Errorf("library not found: %s.%s: %w", prov.Identity(), wnd.Values()["library"], os.ErrNotExist))
			return library.Library{}
		}

		return optLib.Unwrap()
	})

	return ui.VStack(
		breadcrumb.Breadcrumbs(
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/provider", wnd.Values())
			}).Title(StrLibraries.Get(wnd)),
			ui.TertiaryButton(nil).Title(StrLibrary.Get(wnd)),
		),
		ui.H1(StrLibrary.Get(wnd)),
		ui.IfFunc(lib.Valid(), func() core.View {
			return formLibSettings(wnd, libs, lib)
		}),
		docTable(wnd, prov, libs, libID),
	).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
}

func formLibSettings(wnd core.Window, libs provider.Libraries, lib *core.State[library.Library]) core.View {
	type EditForm struct {
		Name        string `label:"nago.common.label.name"`
		Description string `label:"nago.common.label.description" lines:"3"`
	}
	cfg := core.AutoState[EditForm](wnd).Init(func() EditForm {
		return EditForm{
			Name:        lib.Get().Name,
			Description: lib.Get().Description,
		}
	})

	return ui.VStack(
		form.Card(
			form.Auto(form.AutoOptions{}, cfg).FullWidth(),
		),
		ui.HLine(),
		ui.HStack(ui.SecondaryButton(func() {
			info, err := libs.Update(wnd.Subject(), lib.Get().ID, library.UpdateOptions{
				Name:        cfg.Get().Name,
				Description: cfg.Get().Description,
			})
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}
			lib.Set(info)

			alert.ShowBannerMessage(wnd, alert.Message{
				Title:   StrLibraryUpdated.Get(wnd),
				Message: StrLibraryUpdatedDesc.Get(wnd),
				Intent:  alert.IntentOk,
			})

		}).Title(rstring.ActionSave.Get(wnd))).FullWidth().Alignment(ui.Trailing),
	).FullWidth()
}

func docTable(wnd core.Window, prov provider.Provider, libs provider.Libraries, libId library.ID) core.View {

	loadedDocs := core.AutoState[[]document.Document](wnd).AsyncInit(func() []document.Document {
		v, err := xslices.Collect2(libs.Library(libId).All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	return ui.VStack(
		ui.H2(StrDocuments.Get(wnd)),
		ui.If(!loadedDocs.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),
		//dialogNewLibrary(wnd, libs, createPresented, loadedLibs),

		ui.IfFunc(loadedDocs.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedDocs.Get(), []dataview.Field[dataview.Element[document.Document]]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj dataview.Element[document.Document]) core.View {
						return ui.Text(obj.Value.Name)
					},
				},
				{
					Name: rstring.LabelSummary.Get(wnd),
					Map: func(obj dataview.Element[document.Document]) core.View {
						return ui.Text(obj.Value.Summary)
					},
				},
				{
					Name: rstring.LabelState.Get(wnd),
					Map: func(obj dataview.Element[document.Document]) core.View {
						return ui.Text(string(obj.Value.ProcessingStatus))
					},
				},
			}).
				Action(func(e dataview.Element[document.Document]) {
					wnd.Navigation().ForwardTo("admin/ai/library/document", wnd.Values().Put("document", string(e.Value.ID)))
				}).
				NextActionIndicator(true).
				NewActionView(
					ui.Menu(
						ui.PrimaryButton(nil).Title(rstring.ActionNew.Get(wnd)).PreIcon(icons.Plus),
						ui.MenuGroup(
							ui.MenuItem(func() {
								wnd.ImportFiles(core.ImportFilesOptions{
									Multiple: true,
									OnCompletion: func(files []core.File) {
										defer func() {
											loadedDocs.Reset()
										}()

										for _, file := range files {
											reader, err := file.Open()
											if err != nil {
												alert.ShowBannerError(wnd, fmt.Errorf("failed to open uploaded file %s: %w", file.Name(), err))
												return
											}

											_, err = libs.Library(libId).Create(wnd.Subject(), document.CreateOptions{
												Filename: file.Name(),
												Reader:   reader,
											})

											_ = reader.Close()

											if err != nil {
												alert.ShowBannerError(wnd, fmt.Errorf("failed to upload %s: %w", file.Name(), err))
												return
											}
										}

									},
								})
							}, ui.Text(rstring.ActionFileUpload.Get(wnd))),
						),
					),
				).SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
					for _, i := range selected {
						if idx, ok := i.Int(); ok {
							if err := libs.Library(libId).Delete(wnd.Subject(), loadedDocs.Get()[idx].ID); err != nil {
								return err
							}
						}
					}

					loadedDocs.Reset()

					return nil
				}),
			)
		}),
	).FullWidth().Alignment(ui.Leading)

}
