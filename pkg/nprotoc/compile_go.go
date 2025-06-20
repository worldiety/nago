// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"reflect"
	"strings"
	"text/template"
	"unicode"
)

var (
	//go:embed go.tpl
	goTpl string

	//go:embed writer.go
	srcGoWriter string
)

type GoModel struct {
	PackageName  string
	BinaryWriter string
	Marshals     []string
}

func (c *Compiler) GenerateGo() ([]byte, error) {
	tpl, err := template.New("go.tpl").Parse(goTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	prj := c.Project()

	if err := c.emitGoDecl(); err != nil {
		return nil, err
	}

	if err := c.emitMarshal(); err != nil {
		return nil, err
	}

	if err := c.emitUnmarshal(); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, GoModel{
		PackageName:  prj.Go.Package,
		BinaryWriter: trim(srcGoWriter),
		Marshals:     c.marshals,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	buf.Write(c.buf.Bytes())

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return linify(buf.Bytes()), fmt.Errorf("failed to format source: %w", err)
	}

	return src, nil
}

func (c *Compiler) emitGoDecl() error {
	for typename, declaration := range c.sortedDecl() {
		switch decl := declaration.(type) {
		case Enum:
			c.emitGoEnum(typename, decl)
		case Uint:
			c.emitGoUint(typename, decl)
		case Int:
			c.emitGoInt(typename, decl)

		case Record:
			if err := c.emitGoRecord(typename, decl); err != nil {
				return fmt.Errorf("cannot emit record %s: %w", typename, err)
			}
		case String:
			if err := c.goEmitString(typename, decl); err != nil {
				return fmt.Errorf("cannot emit string %s: %w", typename, err)
			}
		case Array:
			if err := c.goEmitArray(typename, decl); err != nil {
				return fmt.Errorf("cannot emit array %s: %w", typename, err)
			}
		case Bool:
			if err := c.goEmitBool(typename, decl); err != nil {
				return fmt.Errorf("cannot emit boolean %s: %w", typename, err)
			}
		case Map:
			if err := c.goEmitMap(typename, decl); err != nil {
				return fmt.Errorf("cannot emit map %s: %w", typename, err)
			}
		case Float64:
			if err := c.goEmitFloat64(typename, decl); err != nil {
				return fmt.Errorf("cannot emit float64 %s: %w", typename, err)
			}
		default:
		}
	}

	return nil
}

func (c *Compiler) Project() Project {
	d, ok := c.declr["_project"]
	if !ok {
		return Project{}
	}

	return d.(Project)
}

func (c *Compiler) makeGoDoc(s string) string {
	if s == "" {
		return ""
	}

	var buf bytes.Buffer
	for _, line := range strings.Split(s, "\n") {
		buf.WriteString(fmt.Sprintf("// %s\n", line))
	}

	return buf.String()
}

func (c *Compiler) goZeroLiteral(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%s(false)", t.Name())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%s(0)", t.Name())
	case reflect.Interface, reflect.Ptr:
		return "nil"
	default:
		return fmt.Sprintf("(%s{})", t.Name())
	}
}

func goFieldName(str string) string {
	if str == "" {
		return ""
	}

	var buf strings.Builder
	for i, r := range str {
		if i == 0 {
			buf.WriteRune(unicode.ToUpper(r))
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}
