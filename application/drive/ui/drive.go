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

	wnd := ctx.Window()

	uc, hasUC := core.FromContext[drive.UseCases](ctx.Window().Context(), "")
	if !hasUC {
		return alert.BannerError(fmt.Errorf("drive management not found")).Render(ctx)
	}

	return ui.VStack(

		ui.HStack(c.viewBreadcrumbs(wnd, c.loadFiles(c.breadcrumbs...), c.onNavigate)).Alignment(ui.Leading).FullWidth(),

		c.viewListing(
			ctx.Window(),
			uc,
			curDir,
		),
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

func (c TDrive) viewListing(wnd core.Window, uc drive.UseCases, curDir drive.File) core.View {

	displyName, ok := core.FromContext[user.DisplayName](wnd.Context(), "")
	if !ok {
		return alert.BannerError(fmt.Errorf("user.DisplayName not found"))
	}

	allIdents := func(yield func(drive.FID, error) bool) {
		file := c.loadFile(c.current.Get()).UnwrapOr(drive.File{})
		for id := range file.Entries.All() {
			if !yield(id, nil) {
				return
			}
		}
	}

	createDirPresented := core.StateOf[bool](wnd, string(curDir.ID)+"-mkdir-presented")
	renamePresented := core.StateOf[bool](wnd, string(curDir.ID)+"-rn-presented")
	selectedFid := core.StateOf[drive.FID](wnd, string(curDir.ID)+"-selectedfid")

	canCreateDir := c.canCreateDirectory(curDir)
	canCreateFile := c.canCreateFile(curDir)

	var columns []dataview.Field[drive.File]

	columns = append(columns, []dataview.Field[drive.File]{

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
			Visible: dataview.MinSizeMedium(),
		},
		{
			ID:   "fname",
			Name: StrName.Get(wnd),
			Map: func(obj drive.File) core.View {
				return ui.Text(obj.Filename)
			},
			Comparator: func(a, b drive.File) int {
				return xstrings.CompareIgnoreCase(a.Filename, b.Filename)
			},
		},

		{
			ID:   "lastmod",
			Name: rstring.LabelChanged.Get(wnd),
			Map: func(obj drive.File) core.View {
				return ui.Text(date.Format(wnd.Locale(), date.Date, obj.ModTime())).AccessibilityLabel(date.Format(wnd.Locale(), date.Time, obj.ModTime()))
			},
			Visible: dataview.MinSizeMedium(),
			Comparator: func(a, b drive.File) int {
				return a.ModTime().Compare(b.ModTime())
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
			Visible: dataview.MinSizeLarge(),
		},
		{
			ID:   "size",
			Name: StrFileSize.Get(wnd),
			Map: func(obj drive.File) core.View {
				if obj.IsDir() {
					return ui.Text(rstring.LabelXItems.Get(wnd, float64(obj.Entries.Len()), i18n.Int("x", obj.Entries.Len())))
				}

				return ui.Text(xstrings.FormatByteSize(wnd.Locale(), obj.Size(), 1))
			},
			Visible: dataview.MinSizeMedium(),
			Comparator: func(a, b drive.File) int {
				return int(a.Size() - b.Size())
			},
		},
	}...)

	dv := dataview.FromData(
		wnd,
		dataview.Data[drive.File, drive.FID]{
			FindAll: allIdents,
			FindByID: func(id drive.FID) (option.Opt[drive.File], error) {
				return uc.Stat(wnd.Subject(), id)
			},
			Fields: columns,
		},
	).ModelOptions(pager.ModelOptions{StatePrefix: string(curDir.ID)}).
		Action(func(e drive.File) {
			if e.IsDir() && c.actionDirectory != nil {
				c.actionDirectory(e)
				return
			}
		}).
		CreateOptions(
			dataview.CreateOption{
				Icon: icons.FolderPlus,
				Name: StrCreateDirectory.Get(wnd),
				Action: func() error {
					createDirPresented.Set(true)
					return nil
				},
				Visible: func() bool {
					return canCreateDir
				},
			},

			dataview.CreateOption{
				Icon: icons.Upload,
				Name: StrUploadFile.Get(wnd),
				Action: func() error {
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
					return nil
				},
				Visible: func() bool {
					return canCreateFile
				},
			},
		).
		SelectOptions(
			dataview.NewSelectOptionDelete(wnd, func(selected []drive.FID) error {
				for _, file := range selected {
					if err := uc.Delete(wnd.Subject(), file, drive.DeleteOptions{
						Recursive: true,
					}); err != nil {
						return err
					}
				}

				return nil
			}),

			dataview.SelectOption[drive.FID]{
				Icon: icons.Download,
				Name: rstring.ActionDownload.Get(wnd),
				Action: func(selected []drive.FID) error {
					if len(selected) == 1 {
						optFile, err := uc.Get(wnd.Subject(), selected[0], "")
						if err != nil {
							return err
						}

						if optFile.IsNone() {
							return os.ErrNotExist
						}

						file := optFile.Unwrap()
						wnd.ExportFiles(core.ExportFilesOptions{
							ID:    string(selected[0]),
							Files: []core.File{file},
						})

						return nil
					}

					// zip export
					zip, err := uc.Zip(wnd.Subject(), selected)
					if err != nil {
						return err
					}

					wnd.ExportFiles(core.ExportFilesOptions{
						ID:    string(selected[0]),
						Files: []core.File{zip},
					})

					return nil
				},
			},

			dataview.SelectOption[drive.FID]{
				Icon: icons.Pen,
				Name: rstring.ActionRename.Get(wnd),
				Action: func(selected []drive.FID) error {
					selectedFid.Set(selected[0])
					renamePresented.Set(true)
					return nil
				},
				Visible: func(selected []drive.FID) bool {
					return len(selected) == 1
				},
			},
		).
		Search(true).
		Style(dataview.Table).
		Selection(true)

	return ui.VStack(
		dialogCreateDirectory(wnd, createDirPresented, func(name string) error {
			return c.mkdir(curDir, name)
		}),
		dialogRename(wnd, renamePresented, uc.Stat, uc.Rename, selectedFid),
		dv,
	).FullWidth()
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

func dialogRename(wnd core.Window, presented *core.State[bool], stat drive.Stat, rename drive.Rename, optFid *core.State[drive.FID]) core.View {
	if !presented.Get() || optFid.Get() == "" {
		return nil
	}

	fid := optFid.Get()

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
