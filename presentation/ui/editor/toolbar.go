// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package editor

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"iter"
	"slices"
	"strings"
)

type TVToolWindow struct {
	name     string
	icon     core.SVG
	top      ui.DecoredView
	content  ui.DecoredView
	bottom   ui.DecoredView
	visible  bool
	position ui.Position
}

func ToolWindow(icon core.SVG, name string) TVToolWindow {
	return TVToolWindow{
		visible: true,
		name:    name,
	}
}

func (c TVToolWindow) Visible(b bool) TVToolWindow {
	c.visible = b
	return c
}

func (c TVToolWindow) Top(v ui.DecoredView) TVToolWindow {
	c.top = v
	return c
}

func (c TVToolWindow) Bottom(v ui.DecoredView) TVToolWindow {
	c.bottom = v
	return c
}

func (c TVToolWindow) Content(v ui.DecoredView) TVToolWindow {
	c.content = v
	return c
}

func (c TVToolWindow) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		ui.IfFunc(c.top != nil, func() core.View {
			return c.top.Frame(ui.Frame{Width: toolbarWidth}).Border(ui.Border{BottomWidth: ui.L1, BottomColor: ui.M7})
		}),
		ui.IfFunc(c.content != nil, func() core.View {
			return ui.ScrollView(c.content.Frame(ui.Frame{Width: ui.Full, Height: ui.Full})).Frame(ui.Frame{Width: ui.Full, Height: ui.Full})
		}),

		ui.Spacer(),
		c.bottom,
	).Frame(ui.Frame{Width: ui.Full, Height: ui.Full}).Render(ctx)

}

type ToolWindowListConfig[T data.Aggregate[ID], ID data.IDType] struct {
	Name           string
	Icon           core.SVG
	ListIcon       func(T) core.SVG
	OnOptions      func(T)
	OnSelected     func(T)
	OnAddToContent func(T)
	List           iter.Seq2[T, error]
	Delete         func(subject auth.Subject, id ID) error
	CreateEmpty    func(subject auth.Subject) error
}

func ToolWindowList[T data.Aggregate[ID], ID data.IDType](wnd core.Window, cfg ToolWindowListConfig[T, ID]) TVToolWindow {
	var files ui.DecoredView

	deletePresented := core.AutoState[bool](wnd)
	selectedT := core.StateOf[T](wnd, "selected-"+cfg.Name).Observe(func(newValue T) {
		if cfg.OnSelected != nil {
			cfg.OnSelected(newValue)
		}
	})

	if cfg.List != nil {
		if cfg.ListIcon == nil {
			cfg.ListIcon = func(t T) core.SVG {
				return flowbiteOutline.File
			}
		}
		files = ui.VStack()
		docs, err := xslices.Collect2(cfg.List)
		if err != nil {
			files = ui.VStack(alert.BannerError(err))
		} else {
			slices.SortFunc(docs, func(a, b T) int {
				return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
			})

			var tmp []core.View
			for _, doc := range docs {
				var selectedColor ui.Color

				if doc.Identity() == selectedT.Get().Identity() {
					selectedColor = ui.M5
				}
				tmp = append(tmp, ui.HoverGroup(
					ui.HStack(ui.ImageIcon(cfg.ListIcon(doc)), ui.Text(fmt.Sprintf("%v", doc))).
						Alignment(ui.Leading).
						BackgroundColor(selectedColor).
						Frame(ui.Frame{Width: ui.Full, Height: ui.Full}),

					ui.HStack(
						ui.ImageIcon(cfg.ListIcon(doc)),
						ui.Text(fmt.Sprintf("%v", doc)),
						ui.Spacer(),

						ui.IfFunc(cfg.OnAddToContent != nil, func() core.View {
							return ui.TertiaryButton(func() {
								selectedT.Set(doc)
								selectedT.Notify()
								cfg.OnAddToContent(doc)
							}).PreIcon(flowbiteOutline.ChevronRight).AccessibilityLabel(fmt.Sprintf("%s hinzufügen", cfg.Name))
						}),

						ui.IfFunc(cfg.Delete != nil, func() core.View {
							return ui.TertiaryButton(func() {
								selectedT.Set(doc)
								selectedT.Notify()
								deletePresented.Set(true)
							}).PreIcon(flowbiteOutline.TrashBin).AccessibilityLabel(fmt.Sprintf("%s erstellen", cfg.Name))
						}),

						ui.TertiaryButton(func() {
							selectedT.Set(doc)
							selectedT.Notify()
							if cfg.OnOptions != nil {
								cfg.OnOptions(doc)
							}
						}).PreIcon(flowbiteOutline.DotsVertical).
							AccessibilityLabel("Optionen").
							Visible(cfg.OnOptions != nil),
					).
						Alignment(ui.Leading).
						Action(func() {
							selectedT.Set(doc)
							selectedT.Notify()
						}).
						BackgroundColor(ui.M4).
						Frame(ui.Frame{Width: ui.Full, Height: ui.Full}),
				).Frame(ui.Frame{Width: ui.Full, Height: ui.L32}))
			}

			files = ui.VStack(tmp...)
		}
	}

	return ToolWindow(cfg.Icon, cfg.Name).
		Top(ui.HStack(
			alert.Dialog(
				"Element löschen",
				ui.Text("Soll der Eintrag gelöscht werden?"),
				deletePresented,
				alert.Cancel(nil),
				alert.Delete(func() {
					if err := cfg.Delete(wnd.Subject(), selectedT.Get().Identity()); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				}),
			),
			ui.Text(cfg.Name).Padding(ui.Padding{Left: ui.L4}),
			ui.Spacer(),
			ui.IfFunc(cfg.CreateEmpty != nil, func() core.View {
				return ui.TertiaryButton(func() {
					if err := cfg.CreateEmpty(wnd.Subject()); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				}).PreIcon(flowbiteOutline.Plus).AccessibilityLabel(fmt.Sprintf("%s erstellen", cfg.Name))
			}),
		).FullWidth()).
		Content(files)
}
