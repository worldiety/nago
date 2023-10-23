package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

type EnumModel struct {
	Cardinality int
	TypeParams  []TypeParam
}

type TypeParam struct {
	Ordinal int
}

//go:embed enums.gotpl
var enumTplText string
var enumTpl = template.Must(template.New("enums.gotpl").Parse(enumTplText))

func emitEnums(dir string, from, to int) error {
	for i := from; i <= to; i++ {
		var model EnumModel
		model.Cardinality = i
		for tp := 1; tp <= i; tp++ {
			model.TypeParams = append(model.TypeParams, TypeParam{Ordinal: tp})
		}

		var buf bytes.Buffer
		if err := enumTpl.Execute(&buf, model); err != nil {
			return err
		}

		fbuf, err := format.Source(buf.Bytes())
		if err != nil {
			return err
		}

		fname := filepath.Join(dir, fmt.Sprintf("e%d.go", i))
		if err := os.WriteFile(fname, fbuf, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
