package uitemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/picker"
	"golang.org/x/text/language"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

func viewProjectExecute(wnd core.Window, prj template.Project, uc template.UseCases, runCfgState *core.State[template.RunConfiguration], presented *core.State[bool], console *core.State[string]) core.View {
	if !presented.Get() {
		return nil
	}
	modelErrState := core.AutoState[string](wnd)
	langState := core.AutoState[string](wnd)
	templateNameState := core.AutoState[string](wnd)
	presentedAddRunConfiguration := core.AutoState[bool](wnd)

	modelState := core.AutoState[string](wnd).Observe(func(newValue string) {
		if newValue == "" {
			modelErrState.Set("")
			return
		}

		var tmp any
		if err := json.Unmarshal([]byte(newValue), &tmp); err != nil {
			modelErrState.Set("Das JSON Modell ist ungültig: " + err.Error())
		} else {
			modelErrState.Set("")
		}

	})

	runCfgState.Observe(func(cfg template.RunConfiguration) {
		modelErrState.Set("")
		langState.Set(cfg.Language)
		templateNameState.Set(cfg.Template)
		modelState.Set(cfg.Model)
	})

	var content core.View
	switch prj.Type {
	case template.TreeTemplatePlain, template.TreeTemplateHTML, template.LatexPDF, template.TypstPDF:
		content = executeTreeTemplateView(wnd, prj, uc, runCfgState, templateNameState, langState, modelState, modelErrState, presentedAddRunConfiguration)
	}

	return ui.Modal(ui.VStack(
		ui.Dialog(
			content,
		).ModalPadding(ui.Padding{}.All(ui.L4)).
			DisableBoxLayout(true).
			Title(ui.Text("Ausführen")).
			TitleX(ui.TertiaryButton(func() {
				presented.Set(false)
			}).PreIcon(flowbiteOutline.Close)).
			Footer(ui.HStack(
				ui.SecondaryButton(func() {
					presentedAddRunConfiguration.Set(true)
				}).Title("Konfiguration speichern"),
				ui.Spacer(),
				ui.PrimaryButton(func() {
					if modelErrState.Get() != "" {

						return
					}

					var langTag language.Tag
					if t, err := language.Parse(langState.Get()); err == nil {
						langTag = t
					}

					var obj any
					if modelState.Get() != "" {
						if err := json.Unmarshal([]byte(modelState.Get()), &obj); err != nil {
							console.Set(err.Error())

							return
						}
					}

					reader, err := uc.Execute(wnd.Subject(), prj.ID, template.ExecOptions{
						Context:      context.Background(),
						Language:     langTag,
						TemplateName: templateNameState.Get(),
						Model:        obj,
					})

					if err != nil {
						console.Set(err.Error())

						return
					}

					buf, err := io.ReadAll(reader)
					if err != nil {
						console.Set(err.Error())

						return
					}

					if utf8.Valid(buf) {
						console.Set(string(buf))
					} else {
						console.Set(fmt.Sprintf("created %d output bytes", len(buf)))
					}

				}).Title("Ausführen"),
			).FullWidth()),
	).Position(ui.Position{
		Type:  ui.PositionFixed,
		Top:   "9rem",
		Right: "0px",
	}).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{Width: "29rem", Height: ui.L880}))

}

func executeTreeTemplateView(
	wnd core.Window,
	prj template.Project,
	uc template.UseCases,
	runCfgState *core.State[template.RunConfiguration],
	templateNameState *core.State[string],
	langState *core.State[string],
	modelState *core.State[string],
	modelErrState *core.State[string],
	presentedAddRunConfiguration *core.State[bool],
) core.View {

	newRunConfigurationName := core.AutoState[string](wnd)
	return ui.VStack(
		alert.Dialog(
			"Konfiguration hinzufügen",
			ui.TextField("Name", newRunConfigurationName.Get()).InputValue(newRunConfigurationName),
			presentedAddRunConfiguration,
			alert.Cancel(nil),
			alert.Save(func() (close bool) {
				cfg := runCfgState.Get()
				cfg.Name = newRunConfigurationName.Get()
				cfg.Language = langState.Get()
				cfg.Template = templateNameState.Get()
				cfg.Model = modelState.Get()

				if err := uc.AddRunConfiguration(wnd.Subject(), prj.ID, cfg); err != nil {
					alert.ShowBannerError(wnd, err)
					return false
				}

				return true
			}),
		),
		configurationPicker(wnd, uc, prj, runCfgState),
		ui.HLine(),
		ui.TextField("Sprache", langState.Get()).
			SupportingText("Leer lassen für undefined. Ansonsten BCP47 Code, wie z.B. de oder en_US").
			FullWidth().
			InputValue(langState),
		ui.TextField("Template", templateNameState.Get()).
			SupportingText("Ein Templatename, wie er in der Templatesprache mittels {{define \"myname\"}} definiert wurde.").
			FullWidth().
			InputValue(templateNameState),
		ui.Text("Modell"),
		ui.CodeEditor(modelState.Get()).
			Frame(ui.Frame{Height: ui.L160}).
			FullWidth().
			Language("json").
			InputValue(modelState),
		ui.IfElse(modelErrState.Get() == "",
			ui.Text("JSON Eingabe, mit der das Template ausgeführt werden soll. Erforderlich, wenn Variablen interpoliert werden müssen.").
				Font(ui.Small),
			ui.Text(modelErrState.Get()).Font(ui.Small).Color(ui.ColorError),
		),
	).Gap(ui.L8).FullWidth().Alignment(ui.Leading)
}

func configurationPicker(wnd core.Window, uc template.UseCases, prj template.Project, runConfigurationSelected *core.State[template.RunConfiguration]) core.View {
	var groups []ui.TMenuGroup

	groups = append(groups, ui.MenuGroup(
		ui.MenuItem(func() {
			runConfigurationSelected.Set(template.RunConfiguration{})
			runConfigurationSelected.Notify()
		}, ui.Text("leere Konfiguration")),
	))

	slices.SortFunc(prj.RunConfigurations, func(a, b template.RunConfiguration) int {
		return strings.Compare(a.Name, b.Name)
	})

	invalidate := core.AutoState[int](wnd)

	if len(prj.RunConfigurations) > 0 {
		var items []ui.TMenuItem

		for _, configuration := range prj.RunConfigurations {
			items = append(items, ui.MenuItem(func() {
				runConfigurationSelected.Set(configuration)
				runConfigurationSelected.Notify()

			}, ui.HStack(
				ui.Text(configuration.Name).TextAlignment(ui.TextAlignStart),
				ui.Spacer(),
				ui.TertiaryButton(func() {
					if err := uc.RemoveRunConfiguration(wnd.Subject(), prj.ID, configuration.ID); err != nil {
						alert.ShowBannerError(wnd, err)
					}

					invalidate.Set(invalidate.Get() + 1)
				}).PreIcon(flowbiteOutline.TrashBin),
			).FullWidth()))
		}

		groups = append(groups, ui.MenuGroup(items...))
	}

	return ui.Menu(
		picker.Button(nil).Content(ui.Text("Konfiguration wählen")).Frame(ui.Frame{}.FullWidth()),
		groups...,
	).Frame(ui.Frame{}.FullWidth())
}
