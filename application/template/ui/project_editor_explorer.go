package uitemplate

import (
	"fmt"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
	"path"
	"strings"
)

func viewProjectExplorer(wnd core.Window, prj template.Project, selectedFile *core.State[template.File]) ui.DecoredView {
	return ui.VStack(
		// toolbar
		ui.HStack(
			ui.Menu(ui.TertiaryButton(nil).PreIcon(flowbiteOutline.DotsVertical),
				ui.MenuGroup(
					ui.MenuItem(func() {

					}, ui.Text("Neue Datei")),
				),
			),
		).FullWidth().
			Alignment(ui.Trailing),
		ui.HLine().Padding(ui.Padding{}),
		// tree view
		ui.ScrollView(
			ui.VStack(
				ui.ForEach(prj.Files, func(t template.File) core.View {
					return fileEntry(selectedFile, t)
				})...,
			).Alignment(ui.Leading).FullWidth(),
		).Frame(ui.Frame{}.FullWidth()).
			Axis(ui.ScrollViewAxisVertical),
	).Alignment(ui.TopLeading).Frame(ui.Frame{Width: ui.L560})
}

func iconByFileName(name string) core.SVG {
	switch strings.ToLower(path.Ext(name)) {
	case ".html", ".gohtml":
		return flowbiteSolid.Html
	case ".jpg", ".svg", ".jpeg", ".png", ".gif", ".bmp":
		return flowbiteSolid.Image
	case ".csv":
		return flowbiteSolid.FileCsv
	case ".css":
		return flowbiteSolid.Css
	case ".go", ".tex", ".typ":
		return flowbiteOutline.FileCode
	case ".pdf":
		return flowbiteSolid.FilePdf
	default:
		return flowbiteSolid.File
	}
}

func fileEntry(selectedFile *core.State[template.File], file template.File) core.View {
	if strings.Contains(file.Filename, "/") {
		segments := strings.Split(file.Filename, "/")
		var tree []core.View
		for idx, segment := range segments {
			if idx == len(segments)-1 {
				// last one
				tree = append(tree, leafFileEntry(selectedFile, file, path.Base(file.Filename), idx))
			} else {
				tree = append(tree, nodeFileEntry(segment, idx))
			}
		}

		return ui.VStack(tree...).Alignment(ui.TopLeading).FullWidth()
	}

	return leafFileEntry(selectedFile, file, file.Filename, 0)
}

const baseRemIndent = 1

func nodeFileEntry(name string, indent int) core.View {
	return ui.HStack(
		ui.FixedSpacer(ui.Length(fmt.Sprintf("%0.2frem", baseRemIndent*float64(indent))), ""),
		ui.ImageIcon(flowbiteOutline.ChevronDown),
		ui.ImageIcon(flowbiteOutline.Folder),
		ui.Text(name),
	).Gap(ui.L8).
		Alignment(ui.Leading).
		FullWidth().
		HoveredBackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L4)).
		Padding(ui.Padding{}.All(ui.L4))
}

func leafFileEntry(selectedFile *core.State[template.File], file template.File, name string, indent int) core.View {
	moreIndent := 0.0
	if indent > 0 {
		moreIndent = baseRemIndent * 3
	}

	var selectedBgColor ui.Color
	hoveredBgColor := ui.ColorCardBody
	if selectedFile.Get().Filename == file.Filename {
		selectedBgColor = ui.ColorCardFooter
		hoveredBgColor = ui.ColorInteractive
	}

	return ui.HStack(
		ui.FixedSpacer(ui.Length(fmt.Sprintf("%0.2frem", moreIndent+baseRemIndent*float64(indent))), ""),
		ui.ImageIcon(iconByFileName(name)),
		ui.Text(name),
	).Gap(ui.L8).
		Action(func() {
			selectedFile.Set(file)
			selectedFile.Notify()
		}).
		Alignment(ui.Leading).
		FullWidth().
		HoveredBackgroundColor(hoveredBgColor).
		BackgroundColor(selectedBgColor).
		Border(ui.Border{}.Radius(ui.L4)).
		Padding(ui.Padding{}.All(ui.L4))
}
