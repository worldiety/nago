package uitemplate

import (
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"io"
	"os"
	"path"
	"strings"
)

func viewProjectSource(wnd core.Window, prj template.Project, selectedFile *core.State[template.File], uc template.UseCases, save func(str string)) core.View {
	const css = "calc(100dvh - 27rem)" // TODO i don't understand the nested css rules here

	if selectedFile.Get().Blob == "" {
		return ui.VStack().Frame(ui.Frame{Height: css, MinHeight: css}.FullWidth())
	}

	optReader, err := uc.LoadProjectBlob(wnd.Subject(), prj.ID, selectedFile.Get().Blob)
	if err != nil {
		return alert.BannerError(err)
	}

	if optReader.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return alert.BannerError(err)
	}

	editState := core.AutoState[string](wnd).Observe(func(newValue string) {
		if string(buf) != newValue {
			save(newValue)
		}
	})

	return ui.ScrollView(
		ui.CodeEditor(string(buf)).
			InputValue(editState).
			Frame(ui.Frame{Width: "100%", Height: css}).
			Language(estimateEditorType(selectedFile.Get())),
	).Frame(ui.Frame{Height: css, MinHeight: css}.FullWidth()).
		Axis(ui.ScrollViewAxisVertical)
}

func estimateEditorType(file template.File) string {
	switch strings.ToLower(path.Ext(file.Filename)) {
	case ".html", ".htm", ".tmpl", ".gohtml":
		return "html"
	case ".go":
		return "go"
	case ".css":
		return "css"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	default:
		return ""
	}
}
