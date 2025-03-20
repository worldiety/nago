package uitemplate

import (
	"bytes"
	"fmt"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"path"
	"strings"
)

func viewProjectExplorer(wnd core.Window, prj template.Project, uc template.UseCases, selectedFile *core.State[template.File]) ui.DecoredView {
	presentedNewFile := core.AutoState[bool](wnd)
	presentedRenameFile := core.AutoState[bool](wnd)
	presentedDeleteFile := core.AutoState[bool](wnd)
	fileName := core.AutoState[string](wnd)

	var fileMenuOptions []ui.TMenuItem
	if selectedFile.Get().Blob != "" {
		fileMenuOptions = append(fileMenuOptions, ui.MenuItem(func() {
			presentedRenameFile.Set(true)
			fileName.Set(selectedFile.Get().Filename)
		}, ui.Text(selectedFile.Get().Filename+" umbenennen").TextAlignment(ui.TextAlignStart)))
		fileMenuOptions = append(fileMenuOptions, ui.MenuItem(func() {
			presentedDeleteFile.Set(true)
		}, ui.Text(selectedFile.Get().Filename+" löschen").TextAlignment(ui.TextAlignStart)))
	}

	return ui.VStack(
		// create file
		ui.IfFunc(presentedNewFile.Get(), func() core.View {
			return alert.Dialog(
				"Neue Datei",
				ui.TextField("Neuer Dateiname", fileName.Get()).InputValue(fileName),
				presentedNewFile,
				alert.Cancel(nil),
				alert.Save(func() (close bool) {
					if err := uc.CreateProjectBlob(wnd.Subject(), prj.ID, fileName.Get(), bytes.NewBuffer(nil)); err != nil {
						alert.ShowBannerError(wnd, err)
						return false
					}

					return true
				}),
			)
		}),

		// rename file
		ui.IfFunc(presentedRenameFile.Get(), func() core.View {
			return alert.Dialog(
				"Datei umbenennen",
				ui.TextField("Neuer Dateiname", fileName.Get()).InputValue(fileName).SupportingText("Pfade per / trennen. Lokalisierte Dateien liegen in locales/<lang>/<filename>"),
				presentedRenameFile,
				alert.Cancel(nil),
				alert.Save(func() (close bool) {
					if err := uc.RenameProjectBlob(wnd.Subject(), prj.ID, selectedFile.Get().Filename, fileName.Get()); err != nil {
						alert.ShowBannerError(wnd, err)
						return false
					}

					f := selectedFile.Get()
					f.Filename = fileName.Get()
					selectedFile.Set(f)
					return true
				}),
			)
		}),

		// delete file
		ui.IfFunc(presentedDeleteFile.Get(), func() core.View {
			return alert.Dialog(
				"Datei löschen",
				ui.Text(fmt.Sprintf("Soll die Datei '%s' wirklich gelöscht werden?", selectedFile.Get().Filename)),
				presentedDeleteFile,
				alert.Cancel(nil),
				alert.Delete(func() {
					if err := uc.DeleteProjectBlob(wnd.Subject(), prj.ID, selectedFile.Get().Filename); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					selectedFile.Set(template.File{})
				}),
			)
		}),

		// toolbar
		ui.HStack(
			ui.Menu(ui.TertiaryButton(nil).PreIcon(flowbiteOutline.DotsVertical),
				ui.MenuGroup(
					ui.MenuItem(func() {
						fileName.Set("")
						presentedNewFile.Set(true)
					}, ui.Text("Neue Datei")),
				),
				ui.MenuGroup(fileMenuOptions...),
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
