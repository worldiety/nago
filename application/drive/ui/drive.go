// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidrive

import (
	"fmt"
	"log/slog"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
	"golang.org/x/text/language"
)

var (
	StrName                 = i18n.MustString("nago.drive.table.filename", i18n.Values{language.English: "Name", language.German: "Name"})
	StrCreateDirectory      = i18n.MustString("nago.drive.create_directory", i18n.Values{language.English: "Create folder", language.German: "Ordner erstellen"})
	StrUploadFile           = i18n.MustString("nago.drive.upload_file", i18n.Values{language.English: "Upload File", language.German: "Datei hochladen"})
	StrDialogCreateDirTitle = i18n.MustString("nago.drive.dialog.create_directory_title", i18n.Values{language.English: "Create a folder", language.German: "Einen Ordner erstellen"})
	StrDialogEnterDirName   = i18n.MustString("nago.drive.dialog.enter_directory_name", i18n.Values{language.English: "Enter your folder name", language.German: "Name des Ordners eingeben"})
)

type TDrive struct {
	breadcrumbs        []drive.FID
	frame              ui.Frame
	onNavigate         func(drive.File)
	current            drive.FID
	stat               func(id drive.FID) (option.Opt[drive.File], error)
	canCreateFile      func(parentDir drive.File) bool
	canCreateDirectory func(parentDir drive.File) bool
	mkdir              func(parentDir drive.File, name string) error
}

func Drive() TDrive {
	return TDrive{}
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

func (c TDrive) Current(fid drive.FID) TDrive {
	c.current = fid
	return c
}

func (c TDrive) Breadcrumbs(breadcrumbs ...drive.FID) TDrive {
	c.breadcrumbs = breadcrumbs
	return c
}

func (c TDrive) autoInit(ctx core.RenderContext) TDrive {
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

	if len(c.breadcrumbs) == 0 && c.current != "" {
		c.breadcrumbs = []drive.FID{c.current}
	}

	return c
}

func (c TDrive) Render(ctx core.RenderContext) core.RenderNode {
	c = c.autoInit(ctx)
	curDir := c.loadFile(c.current).UnwrapOr(drive.File{})
	canCreateDir := c.canCreateDirectory(curDir)
	canCreateFile := c.canCreateFile(curDir)

	wnd := ctx.Window()
	createDirPresented := core.StateOf[bool](wnd, string(curDir.ID)+"-mkdir-presented")

	return ui.VStack(
		dialogCreateDirectory(ctx.Window(), createDirPresented, func(name string) error {
			return c.mkdir(curDir, name)
		}),
		viewBreadcrumbs(c.loadFiles(c.breadcrumbs...), c.onNavigate),
		ui.HStack(
			ui.If(canCreateDir, ui.SecondaryButton(func() {
				createDirPresented.Set(true)
			}).Title(StrCreateDirectory.Get(ctx.Window()))),
			ui.If(canCreateFile, ui.SecondaryButton(func() {

			}).Title(StrUploadFile.Get(ctx.Window()))),
		).FullWidth().
			Alignment(ui.Trailing).
			Gap(ui.L8),
		c.viewListing(ctx.Window(), curDir),
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

func (c TDrive) viewListing(wnd core.Window, file drive.File) core.View {
	if file.ID == "" {
		return nil
	}

	return dataview.FromData(wnd, dataview.Data[drive.File, drive.FID]{
		FindAll: func(yield func(drive.FID, error) bool) {
			for id := range file.Entries.All() {
				if !yield(id, nil) {
					return
				}
			}
		},
		FindByID: func(id drive.FID) (option.Opt[drive.File], error) {
			return c.stat(id)
		},
		Fields: []dataview.Field[drive.File]{
			{
				Name: StrName.Get(wnd),
				Map: func(obj drive.File) core.View {
					return ui.Text(obj.Filename)
				},
			},
		},
	})
}

func viewBreadcrumbs(breadcrumbs []drive.File, onNavigate func(drive.File)) core.View {
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
		tmp = append(tmp, ui.TertiaryButton(action).Title(file.Name()))
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
