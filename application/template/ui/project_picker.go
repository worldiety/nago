package uitemplate

import (
	"fmt"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"slices"
)

func ProjectPickerPage(wnd core.Window, pages Pages, findAll template.FindAll, deleteByID template.Delete) core.View {
	var filter []template.Tag
	if tag := wnd.Values()["tag"]; tag != "" {
		filter = append(filter, template.Tag(tag))
	}

	var projects []template.Project
	for project, err := range findAll(wnd.Subject(), filter) {
		if err != nil {
			return alert.BannerError(err)
		}

		projects = append(projects, project)
	}

	return ui.VStack(
		ui.H1("Vorlagen & Templates"),
		ui.HStack(
			ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo(pages.NewProject, core.Values{"tag": wnd.Values()["tag"]})
			}).PreIcon(heroSolid.Plus).Title("Neues Projekt erstellen"),
		).FullWidth().Alignment(ui.Trailing),
		projectList(wnd, pages, projects, deleteByID),
	).
		Gap(ui.L8).
		FullWidth().
		Alignment(ui.Leading)
}

func projectList(wnd core.Window, pages Pages, projects []template.Project, deleteByID template.Delete) core.View {
	if len(projects) == 0 {
		return ui.Text("Es gibt noch keine Projekte.")
	}

	return list.List(
		ui.Each(slices.Values(projects), func(t template.Project) core.View {
			deletePrjPresented := core.StateOf[bool](wnd, fmt.Sprintf("delete-%s", t.ID))
			return list.Entry().
				Leading(templateLogo(t)).
				Headline(t.Name).
				SupportingText(t.Description).
				Trailing(
					ui.HStack(
						alert.Dialog("Löschen", ui.Text("Soll das Projekt "+t.Name+" unwiderruflich gelöscht werden?"), deletePrjPresented, alert.Cancel(nil), alert.Delete(func() {
							if err := deleteByID(wnd.Subject(), t.ID); err != nil {
								alert.ShowBannerError(wnd, err)
							}
						})),
						ui.SecondaryButton(func() {
							deletePrjPresented.Set(true)
						}).PreIcon(flowbiteOutline.TrashBin).AccessibilityLabel("Löschen"),
						ui.SecondaryButton(func() {
							wnd.Navigation().ForwardTo(pages.Editor, core.Values{"project": string(t.ID)})
						}).PreIcon(heroSolid.Pencil).AccessibilityLabel("bearbeiten"),
					).Gap(ui.L8),
				)
		})...,
	).Caption(ui.Text("Projekte")).Frame(ui.Frame{}.FullWidth())
}

func templateLogo(t template.Project) core.View {
	if t.Logo != "" {
		return ui.Image().URI(t.Logo).Frame(ui.Frame{}.Size(ui.L24, ui.L24))
	}

	switch t.Type {
	case template.AsciidocPDF, template.LatexPDF, template.TypstPDF:
		return ui.ImageIcon(heroSolid.DocumentText)
	default:
		return ui.ImageIcon(heroSolid.Square3Stack3d)
	}
}
