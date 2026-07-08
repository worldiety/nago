// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidrive

import (
	"log/slog"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
)

// dialogMove renders the "move to" folder picker dialog. The user navigates through the drive folder tree
// (only directories are shown) and confirms the destination with "Move here". The destination is the folder
// currently displayed inside the dialog, mirroring the SharePoint behavior.
func (c TDrive) dialogMove(wnd core.Window, uc drive.UseCases, presented *core.State[bool], fidsState *core.State[[]drive.FID]) core.View {
	if !presented.Get() {
		return nil
	}

	fids := fidsState.Get()
	if len(fids) == 0 {
		return nil
	}

	movedSet := make(map[drive.FID]struct{}, len(fids))
	for _, id := range fids {
		movedSet[id] = struct{}{}
	}

	// determine the source parent (all selected items share the current directory) and the drive root as the
	// starting point of the picker.
	var sourceParent drive.FID
	if optFirst := c.loadFile(fids[0]); optFirst.IsSome() {
		sourceParent = optFirst.Unwrap().Parent
	}

	rootFID := c.pickerRoot(sourceParent)

	// the folder currently shown inside the dialog; starts at the drive root.
	navState := core.StateOf[drive.FID](wnd, "mv-nav-"+string(fids[0])).Init(func() drive.FID {
		return rootFID
	})

	curNav := navState.Get()
	if curNav == "" {
		curNav = rootFID
		navState.Set(rootFID)
	}

	optNavDir := c.loadFile(curNav)
	navDir := optNavDir.UnwrapOr(drive.File{})

	// breadcrumb path from the root down to the currently shown folder.
	breadcrumbs, err := c.calculateBreadcrumbs(rootFID, curNav)
	if err != nil {
		slog.Error("move dialog: cannot compute breadcrumbs", "err", err)
	}

	// only directories are selectable targets; exclude the moved items themselves so the user cannot descend
	// into a subtree that is being moved.
	childDirs := func(yield func(drive.FID, error) bool) {
		for id := range navDir.Entries.All() {
			if _, moved := movedSet[id]; moved {
				continue
			}
			optChild := c.loadFile(id)
			if optChild.IsNone() || !optChild.Unwrap().IsDir() {
				continue
			}
			if !yield(id, nil) {
				return
			}
		}
	}

	columns := []dataview.Field[drive.File]{
		{
			Name: "",
			Map: func(obj drive.File) core.View {
				return ui.ImageIcon(icons.Folder)
			},
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
			ID:   "items",
			Name: rstring.LabelXItems.Get(wnd, 0, i18n.Int("x", 0)),
			Map: func(obj drive.File) core.View {
				return ui.Text(rstring.LabelXItems.Get(wnd, float64(obj.Entries.Len()), i18n.Int("x", obj.Entries.Len())))
			},
			Visible: dataview.MinSizeMedium(),
		},
	}

	listing := dataview.FromData(
		wnd,
		dataview.Data[drive.File, drive.FID]{
			FindAll: childDirs,
			FindByID: func(id drive.FID) (option.Opt[drive.File], error) {
				return uc.Stat(wnd.Subject(), id)
			},
			Fields: columns,
			ID:     "mv-picker-" + string(fids[0]),
		},
	).
		Action(func(e drive.File) {
			if e.IsDir() {
				navState.Set(e.ID)
			}
		}).
		Search(true).
		Style(dataview.List).
		Selection(false)

	// the destination is valid unless it is the source parent (no-op) or one of the moved items itself.
	// The use case additionally rejects moving a directory into its own descendant.
	_, targetIsMoved := movedSet[curNav]
	canMoveHere := curNav != "" && curNav != sourceParent && !targetIsMoved

	body := ui.VStack(
		ui.If(len(breadcrumbs) > 1, ui.HStack(c.viewBreadcrumbs(wnd, c.loadFiles(breadcrumbs...), func(f drive.File) {
			navState.Set(f.ID)
		})).Alignment(ui.Leading).FullWidth()),
		listing,
		ui.If(navDir.ID != "" && navDir.Entries.Len() == 0, ui.Text(StrDialogMoveEmptyDir.Get(wnd)).Color(ui.ColorText).Padding(ui.Padding{}.Vertical(ui.L16))),
	).Gap(ui.L8).FullWidth()

	return alert.Dialog(
		StrDialogMoveTitle.Get(wnd),
		body,
		presented,
		alert.Larger(),
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				target := navState.Get()
				var moved int
				for _, fid := range fids {
					if err := uc.Move(wnd.Subject(), fid, target); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
					moved++
				}

				if moved > 0 {
					alert.ShowBannerMessage(wnd, alert.Message{
						Title:   StrDialogMoveTitle.Get(wnd),
						Message: StrDialogMoveSuccessX.Get(wnd, float64(moved), i18n.Int("amount", moved)),
						Intent:  alert.IntentOk,
					})
				}

				// refresh the current listing
				c.current.Notify()
				close(true)
			}).
				Title(StrDialogMoveConfirm.Get(wnd)).
				Enabled(canMoveHere)
		}),
	)
}

// pickerRoot returns the topmost ancestor to start the move picker from. If a virtual root is configured on the
// drive it is used, otherwise the natural root is discovered by walking up the parent chain from the given fid.
func (c TDrive) pickerRoot(fid drive.FID) drive.FID {
	if c.root != "" {
		return c.root
	}

	visited := map[drive.FID]struct{}{}
	cur := fid
	for cur != "" {
		if _, ok := visited[cur]; ok {
			break // guard against cycles
		}
		visited[cur] = struct{}{}

		optFile := c.loadFile(cur)
		if optFile.IsNone() {
			break
		}

		parent := optFile.Unwrap().Parent
		if parent == "" {
			return cur
		}
		cur = parent
	}

	return fid
}
