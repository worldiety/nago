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
	"slices"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/libsync"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/pager"
	"go.wdy.de/nago/presentation/ui/picker"
	"golang.org/x/text/language"
)

var (
	StrDocuments          = i18n.MustString("nago.ai.admin.library.documents", i18n.Values{language.English: "Documents", language.German: "Dokumente"})
	StrDocument           = i18n.MustString("nago.ai.admin.library.document", i18n.Values{language.English: "Document", language.German: "Dokument"})
	StrLibraryUpdated     = i18n.MustString("nago.ai.admin.library.updated", i18n.Values{language.English: "Library updated", language.German: "Bibliothek aktualisiert"})
	StrLibraryUpdatedDesc = i18n.MustString("nago.ai.admin.library.updated_desc", i18n.Values{language.English: "The Library has been updated.", language.German: "Die Bibliothek wurde erfolgreich aktualisiert."})
	StrNoSyncLib          = i18n.MustString("nago.ai.admin.library.no_synclib", i18n.Values{language.English: "There is no library synchronization configured.", language.German: "Es ist keine automatische Bibliothekssynchronisation aktiviert."})
	StrDeleteSyncJob      = i18n.MustString("nago.ai.admin.library.delete_sync_job", i18n.Values{language.English: "Do your really want to delete the synchronization job? Already uploaded documents are not removed.", language.German: "Soll der Synchronizationsauftrag mit der KI Bibliothek wirklich entfernt werden? Bereits bestehende Dokumente werden nicht entfernt."})
	StrSyncSuccess        = i18n.MustString("nago.ai.admin.library.sync_success", i18n.Values{language.English: "Synchronization completed successfully.", language.German: "Synchronisation erfolgreich durchgeführt."})
)

func PageLibrary(wnd core.Window, stores blob.Stores, readDrives drive.ReadDrives, stat drive.Stat, uc ai.UseCases, ucLibSync libsync.UseCases) core.View {
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
		ui.Space(ui.L48),
		ui.H2(rstring.LabelSynchronization.Get(wnd)),
		formLibSync(wnd, stores, readDrives, stat, uc, ucLibSync, prov, libID),
		ui.Space(ui.L48),
		docTable(wnd, prov, libs, libID),
	).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
}

func formLibSync(wnd core.Window, stores blob.Stores, readDrives drive.ReadDrives, stat drive.Stat, uc ai.UseCases, ucLibSync libsync.UseCases, prov provider.Provider, lib library.ID) core.View {
	optJob, err := ucLibSync.FindByID(wnd.Subject(), lib)
	if err != nil {
		return alert.BannerError(err)
	}

	if optJob.IsNone() {
		invalidate := core.AutoState[bool](wnd)
		return form.Card(
			ui.Text(StrNoSyncLib.Get(wnd)),
			ui.HStack(ui.PrimaryButton(func() {
				_, err := ucLibSync.Create(wnd.Subject(), libsync.Job{
					ID:            lib,
					Provider:      prov.Identity(),
					PullPauseTime: 0,
				})

				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				invalidate.Invalidate() // force redraw for invisible state change
			}).Title(rstring.ActionAdd.Get(wnd))).FullWidth().Alignment(ui.Trailing),
		)
	}

	job := optJob.Unwrap()

	deletePresented := core.AutoState[bool](wnd)
	addStorePresented := core.AutoState[bool](wnd)
	addDrivePresented := core.AutoState[bool](wnd)
	syncEnabled := core.AutoState[bool](wnd).Init(func() bool {
		return len(job.Sources) > 0
	})

	return ui.VStack(
		alert.Dialog(rstring.ActionDelete.Get(wnd), ui.Text(StrDeleteSyncJob.Get(wnd)), deletePresented, alert.Cancel(nil), alert.Delete(func() {
			if err := ucLibSync.Delete(wnd.Subject(), lib); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

		})),
		dialogAddStore(wnd, stores, ucLibSync, addStorePresented, lib),
		dialogAddDrive(wnd, readDrives, ucLibSync, addDrivePresented, lib),
		dataview.FromSlice(wnd, job.Sources, []dataview.Field[dataview.Element[libsync.Source]]{
			{
				Name: rstring.LabelType.Get(wnd),
				Map: func(obj dataview.Element[libsync.Source]) core.View {
					switch {
					case obj.Value.Drive.Valid:
						return ui.Text("Drive")
					case obj.Value.Store.Valid:
						optInfo, _ := stores.Stat(obj.Value.Store.Name)
						return ui.Text("Store-" + optInfo.UnwrapOr(blob.StoreInfo{Name: obj.Value.Store.Name}).Type.String())
					default:
						return ui.Text("???")
					}
				},
			},

			{
				Name: rstring.LabelName.Get(wnd),
				Map: func(obj dataview.Element[libsync.Source]) core.View {
					switch {
					case obj.Value.Drive.Valid:
						optStat, _ := stat(user.SU(), obj.Value.Drive.Root)
						file := optStat.UnwrapOr(drive.File{})
						if file.Parent == "" {
							return ui.Text("root")
						}

						return ui.Text(file.Name())
					case obj.Value.Store.Valid:
						optInfo, _ := stores.Stat(obj.Value.Store.Name)
						title := optInfo.UnwrapOr(blob.StoreInfo{Name: obj.Value.Store.Name}).Title
						if title == "" {
							title = obj.Value.Store.Name
						}

						return ui.Text(title).Resolve(true)
					default:
						return ui.Text("???")
					}
				},
			},

			{
				Name: rstring.LabelDescription.Get(wnd),
				Map: func(obj dataview.Element[libsync.Source]) core.View {
					switch {
					case obj.Value.Drive.Valid:
						return ui.Text(string(obj.Value.Drive.Root))

					case obj.Value.Store.Valid:
						optInfo, _ := stores.Stat(obj.Value.Store.Name)
						return ui.Text(optInfo.UnwrapOr(blob.StoreInfo{}).Description).Resolve(true)
					default:
						return ui.Text("???")
					}
				},
			},
		}).ModelOptions(pager.ModelOptions{StatePrefix: "libsync-src"}).
			NewActionView(ui.Menu(
				ui.PrimaryButton(nil).Title(rstring.ActionNew.Get(wnd)).PreIcon(icons.Plus),
				ui.MenuGroup(
					ui.MenuItem(func() {
						addStorePresented.Set(true)
					}, ui.Text("Store")),
					ui.MenuItem(func() {
						addDrivePresented.Set(true)
					}, ui.Text("Drive")),
				),
			)).SelectOptions(
			dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
				for _, idx := range selected {
					if i, ok := idx.Int(); ok {
						if err := ucLibSync.RemoveSource(wnd.Subject(), lib, job.Sources[i]); err != nil {
							return err
						}
					}
				}

				return nil
			}),
		),

		ui.Space(ui.L16),
		ui.HStack(
			ui.SecondaryButton(func() {
				deletePresented.Set(true)
			}).Title(rstring.ActionDelete.Get(wnd)),
			ui.PrimaryButton(func() {
				syncEnabled.Set(false)
				xsync.Go(func() error {
					return ucLibSync.Synchronize(wnd.Subject(), lib)
				}, func(err error) {
					wnd.Post(func() {
						syncEnabled.Set(true)
					})
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					alert.ShowBannerMessage(wnd, alert.Message{
						Title:   rstring.LabelSynchronization.Get(wnd),
						Message: StrSyncSuccess.Get(wnd),
						Intent:  alert.IntentOk,
					})
				})

			}).Title(rstring.ActionSynchronizeNow.Get(wnd)).Enabled(syncEnabled.Get()),
		).FullWidth().Alignment(ui.Trailing).Gap(ui.L8),
	).FullWidth()
}

type driveEntry struct {
	Name   string
	User   user.ID
	Root   drive.FID
	Global bool
}

func (d driveEntry) String() string {
	str := d.Name + " (" + string(d.Root) + ")"
	if d.Global {
		str = "Global: " + str
	}

	return str
}

func dialogAddDrive(wnd core.Window, readDrives drive.ReadDrives, ucLibSync libsync.UseCases, presented *core.State[bool], lib library.ID) core.View {
	if !presented.Get() {
		return nil
	}

	drives, err := xslices.Collect2(readDrives(wnd.Subject(), wnd.Subject().ID()))
	if err != nil {
		return alert.BannerError(err)
	}

	var items []driveEntry
	for _, drv := range drives {
		items = append(items, driveEntry{
			Name:   drv.Name,
			Global: drv.Namespace == drive.NamespaceGlobal,
			Root:   drv.Root,
			User:   wnd.Subject().ID(),
		})
	}
	
	slices.SortFunc(items, func(a, b driveEntry) int {
		return strings.Compare(a.Name, b.Name)
	})

	selected := core.AutoState[[]driveEntry](wnd)

	return alert.Dialog(
		rstring.ActionAdd.Get(wnd),
		picker.Picker[driveEntry](rstring.LabelSource.Get(wnd), items, selected).
			MultiSelect(true).
			DialogOptions(alert.Large()).
			Frame(ui.Frame{}.FullWidth()),
		presented,
		alert.Cancel(nil),
		alert.Add(func() (close bool) {
			for _, info := range selected.Get() {
				var src libsync.Source
				src.Drive.Valid = true
				src.Drive.Root = info.Root
				if err := ucLibSync.AddSource(wnd.Subject(), lib, src); err != nil {
					alert.ShowBannerError(wnd, err)
					return false
				}
			}

			return true
		}),
		alert.Large(),
	)
}

func dialogAddStore(wnd core.Window, stores blob.Stores, ucLibSync libsync.UseCases, presented *core.State[bool], lib library.ID) core.View {
	if !presented.Get() {
		return nil
	}

	var items []blob.StoreInfo

	for storeName, err := range stores.All() {
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		optIfo, err := stores.Stat(storeName)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		ifo := optIfo.UnwrapOr(blob.StoreInfo{
			Name: storeName,
		})

		items = append(items, ifo)
	}

	slices.SortFunc(items, func(a, b blob.StoreInfo) int {
		return strings.Compare(a.String(), b.String())
	})

	selected := core.AutoState[[]blob.StoreInfo](wnd)

	return alert.Dialog(
		rstring.ActionAdd.Get(wnd),
		picker.Picker[blob.StoreInfo](rstring.LabelSource.Get(wnd), items, selected).
			MultiSelect(true).
			DialogOptions(alert.Large()).
			Frame(ui.Frame{}.FullWidth()),
		presented,
		alert.Cancel(nil),
		alert.Add(func() (close bool) {
			for _, info := range selected.Get() {
				var src libsync.Source
				src.Store.Valid = true
				src.Store.Name = info.Name
				if err := ucLibSync.AddSource(wnd.Subject(), lib, src); err != nil {
					alert.ShowBannerError(wnd, err)
					return false
				}
			}

			return true
		}),
		alert.Large(),
	)
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
						text := obj.Value.Name
						if !strings.Contains(text, " ") {
							return ui.Text(xstrings.EllipsisEnd(text, 30))
						}

						return ui.Text(text)
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
