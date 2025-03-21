package uitemplate

import (
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/list"
	"slices"
	"strings"
)

const (
	page1 = iota
)

func NewProjectPage(wnd core.Window, pages Pages, create template.Create) core.View {
	state := core.AutoState[template.Project](wnd)
	return ui.VStack(
		ui.H1("Projekt"),
		ui.VStack(

			form.MultiSteps(
				form.Step(newProjektPage1(state)).Headline("Bezeichnung"),
				form.Step(newProjektPage2(state)).Headline("Projekt-Typ"),
				form.Step(newProjektPage3(state)).Headline("Projekt-Tags"),
			).CanShow(func(currentIdx int, wantedIndex int) bool {
				switch currentIdx {
				case page1:
					return state.Get().Name != ""
				}

				return true
			}).
				ButtonDone(ui.PrimaryButton(func() {
					if _, err := create(wnd.Subject(), state.Get()); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					vals := core.Values{}
					if tag, ok := wnd.Values()["tag"]; ok {
						vals["tag"] = tag
					}
					wnd.Navigation().BackwardTo(pages.Projects, vals)
				}).Title("Speichern")).
				Frame(ui.Frame{}.FullWidth()),
		).FullWidth(),
	).Alignment(ui.Leading).Frame(ui.Frame{MaxWidth: ui.L560}.FullWidth())
}

func newProjektPage3(state *core.State[template.Project]) core.View {
	tagsState := core.DerivedState[string](state, "tags")

	state.Observe(func(newValue template.Project) {
		var tmp strings.Builder
		for _, tag := range state.Get().Tags {
			tmp.WriteString(string(tag))
			tmp.WriteString("\n")
		}

		tagsState.Set(tmp.String())
	})
	tagsState.Observe(func(newValue string) {
		prj := state.Get()
		prj.Tags = nil
		for _, s := range strings.Split(newValue, "\n") {
			if len(strings.TrimSpace(s)) == 0 {
				continue
			}
			prj.Tags = append(prj.Tags, template.Tag(s))
		}

		state.Set(prj)
		state.Notify()
	})

	return ui.VStack(
		ui.TextField("Tags", tagsState.Get()).
			InputValue(tagsState).
			SupportingText("Ein Tag pro Zeile.").
			Lines(5).
			FullWidth(),
	).Gap(ui.L16).FullWidth()
}

func newProjektPage2(state *core.State[template.Project]) core.View {
	type execType struct {
		headline   string
		supporting string
		typ        template.ExecType
		tags       []template.Tag
	}

	entries := []execType{
		{
			headline:   "Generisch",
			supporting: "Ein generisches Projekt ohne Text- oder HTML-Templateausführung. Eine Ausführung wird durch einen individuellen Prozessschritt ermöglicht.",
			typ:        template.Unprocessed,
		},
		{
			headline:   "Text zu Text",
			supporting: "Eine reine Plain-Text-Vorlage.",
			typ:        template.TreeTemplatePlain,
			tags:       []template.Tag{template.TagText},
		},

		{
			headline:   "HTML zu HTML",
			supporting: "Eine Vorlage mit HTML-Dateien als HTML-Vorlage.",
			typ:        template.TreeTemplateHTML,
			tags:       []template.Tag{template.TagHTML},
		},

		{
			headline:   "Typst zu PDF",
			supporting: "Eine Vorlage, die ein Typst-Projekt als Text-Template ausführt und ein PDF erzeugt.",
			typ:        template.TypstPDF,
			tags:       []template.Tag{template.TagPDF},
		},
		{
			headline:   "Latex zu PDF",
			supporting: "Eine Vorlage, die ein Latex-Projekt als Text-Template ausführt und ein PDF erzeugt.",
			typ:        template.LatexPDF,
			tags:       []template.Tag{template.TagPDF},
		},
		{
			headline:   "AsciiDoc zu PDF",
			supporting: "Eine Vorlage, die ein AsciiDoc-Projekt als Text-Template ausführt und ein PDF erzeugt.",
			typ:        template.AsciidocPDF,
			tags:       []template.Tag{template.TagPDF},
		},
	}

	return ui.VStack(
		list.List(
			ui.ForEach(entries, func(t execType) core.View {
				var selectedView core.View
				if state.Get().Type == t.typ {
					selectedView = ui.ImageIcon(heroSolid.Check)
				} else {
					// glitch: cannot set fixed width of immediate element inside hstack
					selectedView = ui.HStack(ui.VStack().Frame(ui.Frame{}.Size(ui.L24, ui.L24)))
				}

				return list.Entry().
					Leading(selectedView).
					Headline(t.headline).
					SupportingText(t.supporting)
			})...,
		).OnEntryClicked(func(idx int) {
			prj := state.Get()
			prj.Type = entries[idx].typ
			prj.Tags = append(prj.Tags, entries[idx].tags...)
			slices.Sort(prj.Tags)
			prj.Tags = slices.Compact(prj.Tags)
			state.Set(prj)
			state.Notify()
		}),
	)
}

func newProjektPage1(state *core.State[template.Project]) core.View {
	nameState := core.DerivedState[string](state, "name").Init(func() string {
		return state.Get().Name
	})
	nameState.Observe(func(newValue string) {
		prj := state.Get()
		prj.Name = newValue
		state.Set(prj)
		state.Notify()
	})

	idState := core.DerivedState[string](state, "id").Init(func() string {
		return string(state.Get().ID)
	})
	idState.Observe(func(newValue string) {
		prj := state.Get()
		prj.ID = template.ID(newValue)
		state.Set(prj)
		state.Notify()
	})

	return ui.VStack(
		ui.TextField("Name", nameState.Get()).
			InputValue(nameState).
			SupportingText("Ein menschenlesbarer Name. Sollte wenn möglich eindeutig sein.").
			FullWidth(),
		ui.TextField("ID", idState.Get()).
			InputValue(idState).
			SupportingText("Manche Systemkomponenten benötigen einen festen Bezeichner. Alternativ kann das Feld leer gelassen werden, um eine zufällige ID zu generieren.").
			FullWidth(),
	).Gap(ui.L16).FullWidth()
}
