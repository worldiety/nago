package ui

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed *.gohtml
var tplFs embed.FS
var tpl *template.Template

func init() {
	t := template.Must(template.ParseFS(tplFs, "*.gohtml"))
	tpl = t
}

func render(name string, model any, w http.ResponseWriter) {
	err := tpl.ExecuteTemplate(w, name, model)
	if err != nil {
		slog.Default().Error("failed to exec template", slog.Any("err", err))
		return
	}
}
