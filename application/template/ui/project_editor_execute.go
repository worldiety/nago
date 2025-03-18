package uitemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
	"io"
	"unicode/utf8"
)

func viewProjectExecute(wnd core.Window, prj template.Project, uc template.UseCases, presented *core.State[bool], console *core.State[string]) core.View {
	if !presented.Get() {
		return nil
	}

	modelErrState := core.AutoState[string](wnd)
	langState := core.AutoState[string](wnd)
	templateNameState := core.AutoState[string](wnd)

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

	var content core.View
	switch prj.Type {
	case template.TreeTemplatePlain, template.TreeTemplateHTML:
		content = executeTreeTemplateView(wnd, prj, uc, templateNameState, langState, modelState, modelErrState)
	}

	return alert.Dialog("Als Vorschau ausführen",
		content,
		presented,
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				if modelErrState.Get() != "" {
					close(false)
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
						close(false)
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
					close(false)
					return
				}

				buf, err := io.ReadAll(reader)
				if err != nil {
					console.Set(err.Error())
					close(false)
					return
				}

				if utf8.Valid(buf) {
					console.Set(string(buf))
				} else {
					console.Set(fmt.Sprintf("created %d output bytes", len(buf)))
				}

				presented.Set(false)
				close(true)
			}).Title("Run")
		}),
		alert.Closeable(),
		alert.MinWidth(ui.L560),
	)
}

func executeTreeTemplateView(
	wnd core.Window,
	prj template.Project,
	uc template.UseCases,
	templateNameState *core.State[string],
	langState *core.State[string],
	modelState *core.State[string],
	modelErrState *core.State[string],
) core.View {

	return ui.VStack(
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
