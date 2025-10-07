// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidrive

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/worldiety/i18n"
	"github.com/worldiety/i18n/date"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

var (
	StrName                 = i18n.MustString("nago.drive.table.filename", i18n.Values{language.English: "Name", language.German: "Name"})
	StrCreateDirectory      = i18n.MustString("nago.drive.create_directory", i18n.Values{language.English: "Create folder", language.German: "Ordner erstellen"})
	StrUploadFile           = i18n.MustString("nago.drive.upload_file", i18n.Values{language.English: "Upload File", language.German: "Datei hochladen"})
	StrDialogCreateDirTitle = i18n.MustString("nago.drive.dialog.create_directory_title", i18n.Values{language.English: "Create a folder", language.German: "Einen Ordner erstellen"})
	StrDialogEnterDirName   = i18n.MustString("nago.drive.dialog.enter_directory_name", i18n.Values{language.English: "Enter your folder name", language.German: "Name des Ordners eingeben"})
	StrFileSize             = i18n.MustString("nago.drive.table.filesize", i18n.Values{language.English: "File size", language.German: "Größe"})
	StrMyFiles              = i18n.MustString("nago.drive.my_files", i18n.Values{language.English: "My files", language.German: "Meine Dateien"})
	StrAnonymous            = i18n.MustString("nago.drive.anonymous", i18n.Values{language.English: "Anonymous", language.German: "Anonym"})
	StrUploadSuccessTitle   = i18n.MustString("nago.drive.upload.success_title", i18n.Values{language.English: "Upload success", language.German: "Upload erfolgreich"})
	StrUploadSuccessX       = i18n.MustVarString("nago.drive.upload.success_x", i18n.Values{language.English: "File {name} has been uploaded successfully.", language.German: "Die Datei {name} wurde erfolgreich hochgeladen."})
	StrDialogDeleteTitle    = i18n.MustString("nago.drive.dialog.delete_title", i18n.Values{language.English: "Deletes Files", language.German: "Dateien löschen"})
	StrDialogDeleteDescX    = i18n.MustQuantityString(
		"nago.drive.dialog.delete_desc",
		i18n.QValues{
			language.English: i18n.Quantities{One: "Delete the selected file?", Other: "Delete {amount} files?"},
			language.German:  i18n.Quantities{One: "Soll die ausgewählte Datei unwiderruflich gelöscht werden?", Other: "Sollen {amount} Dateien unwiderruflich gelöscht werden?"},
		},
	)

	StrDialogRenameTitle       = i18n.MustString("nago.drive.dialog.rename_title", i18n.Values{language.English: "Rename file", language.German: "Datei umbenennen"})
	StrDialogRenameDesc        = i18n.MustString("nago.drive.dialog.rename_desc", i18n.Values{language.English: "Enter your new name", language.German: "Den neuen Dateinamen eingeben"})
	StrDialogRenameInvalidName = i18n.MustString("nago.drive.dialog.rename_invalid", i18n.Values{language.English: "The file name is invalid. Keep below 255 chars and use less special characters.", language.German: "Der Dateiname ist ungültig. Namen länger als 255 Zeichen sind nicht erlaubt genauso wie verschiedene Sonderzeichen."})
)

type TDrive struct {
	breadcrumbs        []drive.FID
	frame              ui.Frame
	onNavigate         func(drive.File)
	current            *core.State[drive.FID]
	stat               func(id drive.FID) (option.Opt[drive.File], error)
	canCreateFile      func(parentDir drive.File) bool
	canCreateDirectory func(parentDir drive.File) bool
	mkdir              func(parentDir drive.File, name string) error
	actionDirectory    func(drive.File)
	root               drive.FID
	rootName           string
	importOptions      core.ImportFilesOptions
}

func Drive(fid *core.State[drive.FID]) TDrive {
	return TDrive{
		current: fid,
	}
}

func (c TDrive) Stat(stat func(id drive.FID) (option.Opt[drive.File], error)) TDrive {
	c.stat = stat
	return c
}

func (c TDrive) Frame(frame ui.Frame) TDrive {
	c.frame = frame
	return c
}

func (c TDrive) OnNavigate(f func(drive.File)) TDrive {
	c.onNavigate = f
	return c
}

func (c TDrive) DirectoryAction(f func(drive.File)) TDrive {
	c.actionDirectory = f
	return c
}

// Root defines the (virtual) root element, which marks the first element of the hierarchy.
// This allows to visualize any directory as a pseudo root element. If empty, the natural hierarchy is used.
// See also [TDrive.RootName].
func (c TDrive) Root(fid drive.FID) TDrive {
	c.root = fid
	return c
}

func (c TDrive) ImportFilesOptions(opts core.ImportFilesOptions) TDrive {
	c.importOptions = opts
	return c
}

// RootName sets the name of the root object. See also [TDrive.Root]. If empty and the root file has no
// name, [StrMyFiles] is used.
func (c TDrive) RootName(name string) TDrive {
	c.rootName = name
	return c
}

func (c TDrive) Breadcrumbs(breadcrumbs ...drive.FID) TDrive {
	c.breadcrumbs = breadcrumbs
	return c
}

func (c TDrive) autoInit(ctx core.RenderContext) TDrive {
	wnd := ctx.Window()
	if c.canCreateDirectory == nil {
		c.canCreateDirectory = func(parentDir drive.File) bool {
			return parentDir.CanWrite(ctx.Window().Subject())
		}
	}

	if c.canCreateFile == nil {
		c.canCreateFile = func(parentDir drive.File) bool {
			return parentDir.CanWrite(ctx.Window().Subject())
		}
	}

	uc, hasUC := core.FromContext[drive.UseCases](ctx.Window().Context(), "")

	if c.stat == nil {
		c.stat = func(id drive.FID) (option.Opt[drive.File], error) {
			if hasUC {
				return uc.Stat(ctx.Window().Subject(), id)
			}

			return option.None[drive.File](), fmt.Errorf("cannot stat: drive management not found")
		}
	}

	if c.mkdir == nil {
		c.mkdir = func(parentDir drive.File, name string) error {
			if hasUC {
				if _, err := uc.MkDir(ctx.Window().Subject(), parentDir.ID, name, drive.MkDirOptions{}); err != nil {
					return err
				}

				return nil
			}

			return fmt.Errorf("cannot mkdir: drive management not found")
		}
	}

	if len(c.breadcrumbs) == 0 && c.current.Get() != "" {
		fids, err := calculateBreadcrumbs(uc.Stat, wnd.Subject(), c.root, c.current.Get())
		if err != nil {
			slog.Error("failed to calculate breadcrumbs", "err", err)
		}
		c.breadcrumbs = fids
	}

	if c.actionDirectory == nil {
		c.actionDirectory = func(f drive.File) {
			c.current.Set(f.ID)
		}
	}

	if c.onNavigate == nil {
		c.onNavigate = func(f drive.File) {
			c.current.Set(f.ID)
			c.current.Notify()
		}
	}

	if c.importOptions.IsZero() {
		c.importOptions = core.ImportFilesOptions{
			Multiple: true,
		}
	}

	return c
}

func (c TDrive) Render(ctx core.RenderContext) core.RenderNode {
	c = c.autoInit(ctx)
	curDir := c.loadFile(c.current.Get()).UnwrapOr(drive.File{})
	canCreateDir := c.canCreateDirectory(curDir)
	canCreateFile := c.canCreateFile(curDir)

	wnd := ctx.Window()
	createDirPresented := core.StateOf[bool](wnd, string(curDir.ID)+"-mkdir-presented")
	uc, hasUC := core.FromContext[drive.UseCases](ctx.Window().Context(), "")
	if !hasUC {
		return alert.BannerError(fmt.Errorf("drive management not found")).Render(ctx)
	}

	pModel, err := pager.NewModel[drive.File, drive.FID](
		wnd,

		func(id drive.FID) (option.Opt[drive.File], error) {
			return c.stat(id)
		},
		func(yield func(drive.FID, error) bool) {
			file := c.loadFile(c.current.Get()).UnwrapOr(drive.File{})
			for id := range file.Entries.All() {
				if !yield(id, nil) {
					return
				}
			}
		},
		pager.ModelOptions{
			StatePrefix: string(curDir.ID),
		},
	)

	if err != nil {
		return alert.BannerError(err).Render(ctx)
	}

	deletePresented := core.StateOf[bool](wnd, string(curDir.ID)+"-delete-presented")
	renamePresented := core.StateOf[bool](wnd, string(curDir.ID)+"-rename-presented")

	selected := pModel.Selected()
	var selectedFid option.Opt[drive.FID]
	if len(selected) == 1 {
		selectedFid = option.Some(selected[0])
	}

	return ui.VStack(
		dialogDelete(wnd, deletePresented, uc.Delete, selected),
		dialogCreateDirectory(ctx.Window(), createDirPresented, func(name string) error {
			return c.mkdir(curDir, name)
		}),
		dialogRename(wnd, renamePresented, uc.Stat, uc.Rename, selectedFid),
		ui.HStack(c.viewBreadcrumbs(wnd, c.loadFiles(c.breadcrumbs...), c.onNavigate)).Alignment(ui.Leading).FullWidth(),
		ui.HStack(
			ui.Menu(
				ui.SecondaryButton(nil).PreIcon(icons.Grid).Enabled(pModel.SelectionCount > 0).Title(rstring.LabelOptions.Get(wnd)),
				ui.MenuGroup(
					ui.MenuItem(func() {
						deletePresented.Set(true)
					}, ui.HStack(ui.ImageIcon(icons.TrashBin), ui.Text(rstring.ActionDelete.Get(wnd))).Gap(ui.L8)),
					ui.MenuItem(func() {
						selected := pModel.Selected()
						if len(selected) == 1 {
							optFile, err := uc.Get(wnd.Subject(), selected[0], "")
							if err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}

							if optFile.IsNone() {
								alert.ShowBannerError(wnd, os.ErrNotExist)
							}

							file := optFile.Unwrap()
							wnd.ExportFiles(core.ExportFilesOptions{
								ID:    string(selected[0]),
								Files: []core.File{file},
							})

							return
						}

						// zip export
						zip, err := uc.Zip(wnd.Subject(), selected)
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						wnd.ExportFiles(core.ExportFilesOptions{
							ID:    string(selected[0]),
							Files: []core.File{zip},
						})

					}, ui.HStack(ui.ImageIcon(icons.Download), ui.Text(rstring.ActionDownload.Get(wnd))).Gap(ui.L8)),

					ui.MenuItem(func() {
						renamePresented.Set(true)
					}, ui.If(pModel.SelectionCount == 1, ui.HStack(ui.ImageIcon(icons.Pen), ui.Text(rstring.ActionRename.Get(wnd))).Gap(ui.L8))),
				),
			),
			ui.VLineWithColor(ui.ColorInputBorder).Frame(ui.Frame{Height: ui.L40}),
			ui.If(canCreateDir, ui.SecondaryButton(func() {
				createDirPresented.Set(true)
			}).Title(StrCreateDirectory.Get(ctx.Window()))),
			ui.If(canCreateFile, ui.SecondaryButton(func() {
				targetDir := c.current.Get() // important: freeze our target, because it may change during upload
				tmp := c.importOptions
				tmp.OnCompletion = func(files []core.File) {
					if c.importOptions.OnCompletion != nil {
						c.importOptions.OnCompletion(files)
					}

					for _, file := range files {
						reader, err := file.Open()
						if err != nil {
							alert.ShowBannerError(wnd, err)
							continue
						}

						err = uc.Put(wnd.Subject(), targetDir, file.Name(), reader, drive.PutOptions{
							OriginalFilename: file.Name(),
							SourceHint:       drive.Upload,
							KeepVersion:      true,
						})

						_ = reader.Close()

						if err != nil {
							alert.ShowBannerError(wnd, err)
							continue
						}

						alert.ShowBannerMessage(wnd, alert.Message{
							Title:   StrUploadSuccessTitle.Get(wnd),
							Message: StrUploadSuccessX.Get(wnd, i18n.String("name", file.Name())),
							Intent:  alert.IntentOk,
						})
					}
				}

				wnd.ImportFiles(tmp)
			}).Title(StrUploadFile.Get(ctx.Window()))),
		).FullWidth().
			Alignment(ui.Trailing).
			Gap(ui.L8),
		c.viewListing(ctx.Window(), pModel),
	).
		Gap(ui.L16).
		Frame(c.frame).Render(ctx)
}

func (c TDrive) loadFiles(ids ...drive.FID) []drive.File {
	res := make([]drive.File, 0, len(ids))
	for _, id := range ids {
		optFile, err := c.stat(id)
		if err != nil {
			slog.Error("failed to stat drive file", "id", id, "err", err.Error())
			continue
		}

		if optFile.IsNone() {
			slog.Error("drive file is gone", "id", id)
			continue
		}

		res = append(res, optFile.Unwrap())
	}

	return res
}

func (c TDrive) loadFile(id drive.FID) option.Opt[drive.File] {
	optFile, err := c.stat(id)
	if err != nil {
		slog.Error("failed to stat drive file", "id", id, "err", err.Error())
		return option.None[drive.File]()
	}

	if optFile.IsNone() {
		slog.Error("drive file is gone", "id", id)
		return option.None[drive.File]()
	}

	return optFile
}

func (c TDrive) viewListing(wnd core.Window, model pager.Model[drive.File, drive.FID]) core.View {

	displyName, ok := core.FromContext[user.DisplayName](wnd.Context(), "")
	if !ok {
		return alert.BannerError(fmt.Errorf("user.DisplayName not found"))
	}

	return dataview.FromModel(
		wnd,
		model,
		[]dataview.Field[drive.File]{
			{
				Name: "",
				Map: func(obj drive.File) core.View {
					if obj.IsDir() {
						return ui.ImageIcon(icons.Folder)
					}

					mt := obj.FileInfo.UnwrapOr(drive.FileInfo{})
					switch mt.MimeType {
					case "application/pdf":
						return ui.ImageIcon(icons.FilePdf)
					default:
						return ui.ImageIcon(icons.File)
					}
				},
			},
			{
				Name: StrName.Get(wnd),
				Map: func(obj drive.File) core.View {
					return ui.Text(obj.Filename)
				},
			},
			{
				Name: rstring.LabelChanged.Get(wnd),
				Map: func(obj drive.File) core.View {
					return ui.Text(date.Format(wnd.Locale(), date.Date, obj.ModTime())).AccessibilityLabel(date.Format(wnd.Locale(), date.Time, obj.ModTime()))
				},
			},
			{
				Name: rstring.LabelChangedBy.Get(wnd),
				Map: func(obj drive.File) core.View {
					if log, ok := obj.AuditLog.Last(); ok {
						if v, ok := log.Unwrap(); ok && v.ModBy() != "" {
							return ui.Text(displyName(v.ModBy()).Displayname)
						}
					}

					return ui.Text(StrAnonymous.Get(wnd))
				},
			},
			{
				Name: StrFileSize.Get(wnd),
				Map: func(obj drive.File) core.View {
					if obj.IsDir() {
						return ui.Text(rstring.LabelXItems.Get(wnd, float64(obj.Entries.Len()), i18n.Int("x", obj.Entries.Len())))
					}

					return ui.Text(xstrings.FormatByteSize(wnd.Locale(), obj.Size(), 1))
				},
			},
		},
	).Action(func(e drive.File) {
		if e.IsDir() && c.actionDirectory != nil {
			c.actionDirectory(e)
			return
		}
	})
}

func (c TDrive) viewBreadcrumbs(wnd core.Window, breadcrumbs []drive.File, onNavigate func(drive.File)) core.View {
	if len(breadcrumbs) == 0 {
		return nil
	}

	var tmp []core.View
	for idx, file := range breadcrumbs {
		var action func()
		if idx < len(breadcrumbs)-1 {
			action = func() {
				if onNavigate != nil {
					onNavigate(file)
				}
			}
		}

		title := file.Name()
		if title == "" {
			title = c.rootName
		}

		if title == "" {
			title = StrMyFiles.Get(wnd)
		}
		tmp = append(tmp, ui.TertiaryButton(action).Title(title))
	}

	return breadcrumb.Breadcrumbs(tmp...)
}

func dialogCreateDirectory(wnd core.Window, presented *core.State[bool], onCreate func(name string) error) core.View {
	if !presented.Get() {
		return nil
	}

	text := core.DerivedState[string](presented, "dir-name")
	return alert.Dialog(
		StrDialogCreateDirTitle.Get(wnd),
		ui.TextField(StrName.Get(wnd), text.Get()).InputValue(text).SupportingText(StrDialogEnterDirName.Get(wnd)),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			if err := onCreate(text.Get()); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		}),
	)
}

func calculateBreadcrumbs(stat drive.Stat, subject auth.Subject, virtualRoot drive.FID, fid drive.FID) ([]drive.FID, error) {
	var res []drive.FID
	for fid != "" {
		optFile, err := stat(subject, fid)
		if err != nil {
			return nil, err
		}

		if optFile.IsNone() {
			return res, nil
		}

		res = append(res, fid)

		if fid == virtualRoot {
			break
		}

		fid = optFile.Unwrap().Parent
	}

	slices.Reverse(res)
	return res, nil
}

func dialogDelete(wnd core.Window, presented *core.State[bool], delete drive.Delete, files []drive.FID) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		StrDialogDeleteTitle.Get(wnd),
		ui.Text(StrDialogDeleteDescX.Get(wnd, float64(len(files)), i18n.Int("amount", len(files)))),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Delete(func() {
			for _, file := range files {
				if err := delete(wnd.Subject(), file, drive.DeleteOptions{
					Recursive: true,
				}); err != nil {
					alert.ShowBannerError(wnd, err)
				}
			}
		}),
		alert.Large(),
	)
}

func dialogRename(wnd core.Window, presented *core.State[bool], stat drive.Stat, rename drive.Rename, optFid option.Opt[drive.FID]) core.View {
	if !presented.Get() || optFid.IsNone() {
		return nil
	}

	fid := optFid.Unwrap()

	optFile, err := stat(wnd.Subject(), fid)
	if err != nil {
		alert.ShowBannerError(wnd, err)
		return nil
	}

	if optFile.IsNone() {
		alert.ShowBannerError(wnd, fmt.Errorf("file not found: %s: %w", fid, os.ErrNotExist))
		return nil
	}

	file := optFile.Unwrap()
	newName := core.AutoState[string](wnd).Init(func() string {
		return file.Name()
	})

	errState := core.AutoState[string](wnd)

	return alert.Dialog(
		StrDialogRenameTitle.Get(wnd),
		ui.TextField(rstring.LabelName.Get(wnd), newName.Get()).InputValue(newName).SupportingText(StrDialogRenameDesc.Get(wnd)).ErrorText(errState.Get()),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Larger(),
		alert.Save(func() (close bool) {
			errState.Set("")
			if err := rename(wnd.Subject(), fid, newName.Get()); err != nil {
				if errors.Is(err, os.ErrInvalid) {
					errState.Set(StrDialogRenameInvalidName.Get(wnd))
					return
				}

				alert.ShowBannerError(wnd, err)
				return
			}

			return true
		}),
	)
}
