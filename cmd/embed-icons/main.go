package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed embed.tpl
var tplText string

var tpl = template.Must(template.New("embed.tpl").Parse(tplText))

func main() {
	if err := realMain(); err != nil {
		panic(err)
	}
}

func realMain() error {
	paths := []string{
		"presentation/icons/hero/solid",
		"presentation/icons/hero/outline",
		"presentation/icons/flowbite/solid",
		"presentation/icons/flowbite/outline",
		"presentation/icons/material/filled",
		"presentation/icons/material/outlined",
		"presentation/icons/material/round",
		"presentation/icons/material/sharp",
		"presentation/icons/material/two-tone",
	}

	for _, path := range paths {
		if err := deploy(path); err != nil {
			return err
		}
	}

	return nil
}

type TplModel struct {
	Vars []Var
}

type Var struct {
	Varname  string
	Filename string
}

func deploy(dir string) error {
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	fmt.Println(path)

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var res TplModel

	for _, file := range files {
		if file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".svg") {
			res.Vars = append(res.Vars, Var{
				Varname:  makeVarName(file.Name()),
				Filename: file.Name(),
			})
		}
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, res); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(buf.String())
		return err
	}

	return os.WriteFile(filepath.Join(path, "aaa.embed.gen.go"), formatted, 0644)
}

func makeVarName(s string) string {
	name := strings.TrimSuffix(s, ".svg")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.Title(name)
	name = strings.ReplaceAll(name, " ", "")
	if name[0] >= '0' && name[0] <= '9' {
		return "I" + name
	}
	return name
}
