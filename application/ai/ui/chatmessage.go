// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"io"
	"os"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/rest"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/markdown"
)

type MessageStyle int

const (
	MessageAgent MessageStyle = iota
	MessageHuman
)

type TChatMessage struct {
	style       MessageStyle
	text        string
	markdown    bool
	ico         core.SVG
	file        file.File
	fileAsImage bool
	prov        provider.Provider
	download    func()
}

func ChatMessage() TChatMessage {
	return TChatMessage{}
}

func (c TChatMessage) Text(text string) TChatMessage {
	c.text = text
	return c
}

func (c TChatMessage) Markdown(text string) TChatMessage {
	c.text = text
	c.markdown = true
	return c
}

func (c TChatMessage) Download(fn func()) TChatMessage {
	c.download = fn
	return c
}

func (c TChatMessage) File(f file.File) TChatMessage {
	c.file = f
	return c
}

func (c TChatMessage) Image(f file.File) TChatMessage {
	c.file = f
	c.fileAsImage = true
	return c
}

func (c TChatMessage) Icon(ico core.SVG) TChatMessage {
	c.ico = ico
	return c
}

func (c TChatMessage) Style(style MessageStyle) TChatMessage {
	c.style = style
	return c
}

func (c TChatMessage) Provider(prov provider.Provider) TChatMessage {
	c.prov = prov
	return c
}

func (c TChatMessage) Render(ctx core.RenderContext) core.RenderNode {
	var color ui.Color
	if c.style == MessageHuman {
		color = ui.M3
	}

	wnd := ctx.Window()

	// special image case
	if c.file.ID != "" && c.fileAsImage {
		frame := ui.Frame{Width: "62%", Height: ""} // golden ratio in chat but automatically adjust height with correct ratio
		if wnd.Info().SizeClass < core.SizeClassLarge {
			frame.Width = ui.Full
		}

		return ui.VStack(
			ui.Image().
				URI(rest.URL(c.prov.Identity(), c.file.ID)).
				ObjectFit(ui.FitFill).
				Frame(ui.Frame{Width: ui.Full}).
				Border(ui.Border{}.Radius(ui.L12)),

			ui.HStack(
				ui.SecondaryButton(func() {
					wnd.Navigation().Open(rest.URL(c.prov.Identity(), c.file.ID))
				}).PreIcon(icons.Download).AccessibilityLabel(rstring.ActionDownload.Get(wnd)),
			).Position(ui.Position{Type: ui.PositionAbsolute, Top: "0px", Right: "0px"}).
				Padding(ui.Padding{}.All(ui.L4)),
		).Position(ui.Position{Type: ui.PositionOffset}).
			Frame(frame).
			Render(ctx)
	}

	return ui.VStack(
		func() core.View {
			if c.file.ID != "" {
				return ui.HStack(
					ui.ImageIcon(icons.File),
					ui.LinkWithAction(c.file.FilenameWithExt(), func() {
						if c.prov != nil && c.prov.Files().IsSome() {
							optReader, err := c.prov.Files().Unwrap().Get(wnd.Subject(), c.file.ID)
							if err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}

							if optReader.IsNone() {
								alert.ShowBannerError(wnd, fmt.Errorf("file not found: %s: %w", c.file.ID, os.ErrNotExist))
								return
							}

							wnd.ExportFiles(core.ExportFilesOptions{
								Files: []core.File{
									core.NewReaderFile(func() (io.ReadCloser, error) {
										return optReader.Unwrap(), nil
									}).SetName(c.file.Name),
								},
							})
						}
					}).AccessibilityLabel(rstring.ActionDownload.Get(wnd)),
				).FullWidth().Gap(ui.L8)
			}

			if !c.markdown {
				ui.Text(c.text)
			}

			return markdown.Render(markdown.Options{Window: ctx.Window(), RichText: true, TrimParagraph: true}, []byte(c.text))
		}(),

		ui.If(len(c.ico) != 0, ui.ImageIcon(c.ico)),
		ui.If(c.download != nil, ui.HStack(
			ui.TertiaryButton(c.download).PreIcon(icons.Download).AccessibilityLabel(rstring.ActionDownload.Get(wnd)),
		).FullWidth()),
	).
		BackgroundColor(color).
		Border(ui.Border{}.Width(ui.L1).Color(ui.M5).Radius(ui.L24)).
		Padding(ui.Padding{}.All(ui.L12)).
		Render(ctx)
}
