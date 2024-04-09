package main

import (
	"embed"
	"text/template"
)

//go:embed *.tstpl
var tpls embed.FS

var parsedTemplates = template.Must(template.ParseFS(tpls, "*.tstpl"))

func importInto(t *TSTypeDef, dst *[]Import) {
	found := false
	for _, imp := range *dst {
		if imp.Type == t.Name && imp.From == t.Package {
			found = true
			break
		}
	}

	if !found && t.Package != "" {
		*dst = append(*dst, Import{
			Type: t.Name,
			From: t.Package,
		})
	}

	for _, param := range t.TypeParams {
		importInto(param, dst)
	}
}
