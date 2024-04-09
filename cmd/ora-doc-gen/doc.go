package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Doc struct {
	out *strings.Builder
}

func NewDoc() *Doc {
	return &Doc{out: &strings.Builder{}}
}

func (d *Doc) Printf(format string, args ...any) {
	d.out.WriteString(fmt.Sprintf(format, args...))
}

func (d *Doc) PrintJSON(caption string, t any) {
	d.Printf("\n```json\n")
	buf, err := json.MarshalIndent(t, " ", "  ")
	if err != nil {
		panic(err)
	}

	d.out.Write(buf)
	d.Printf("\n```\n")
	d.Printf("_%s_\n\n", caption)
}

func (d *Doc) PrintSpec(caption string, t any) {
	rType := reflect.TypeOf(t)
	d.Printf("| field | type | description |\n")
	d.Printf("|--|--|--|\n")
	for i := range rType.NumField() {
		field := rType.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Tag.Get("json")
		value := field.Tag.Get("value")
		desc := field.Tag.Get("description")

		typName := prettyPrint(field.Type)
		if value != "" {
			typName += " (const `" + value + "`)"
		}

		d.Printf("|%s|%s|%s|\n", name, typName, desc)
	}

	d.Printf("\n\n")
	d.Printf("_%s_\n\n", caption)
}

func (d *Doc) PrintTypescriptIface(caption string, t any) {
	rType := reflect.TypeOf(t)
	d.Printf("\n```typescript\n")
	d.Printf("export interface %s {\n", rType.Name())
	for i := range rType.NumField() {
		field := rType.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Tag.Get("json")
		value := field.Tag.Get("value")
		desc := field.Tag.Get("description")

		if name == "type" {
			d.Printf("  type: '%s',\n", value)
			continue
		}

		if name == "id" {
			d.Printf("  id: number,\n")
			continue
		}

		jsType := prettyPrintTypescript(field.Type)
		origGeneric := prettyPrint(field.Type)

		if desc != "" {
			d.Printf("  /** %s (original type %s) */\n", desc, origGeneric)
		}
		d.Printf("  %s: %s,\n", name, jsType)
	}

	d.Printf("}\n\n")
	d.Printf("\n```\n")
	d.Printf("_%s_\n\n", caption)
}

var rmStuffPkg = regexp.MustCompile(`protocol\.|go\.wdy\.de/nago/presentation/`)
var genericPart = regexp.MustCompile(`\[.+\]`)

func prettyPrint(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Slice:
		return "[]" + prettyPrint(t.Elem())
	case reflect.Ptr:
		return "*" + prettyPrint(t.Elem())
	case reflect.Map:
		return "map[" + prettyPrint(t.Key()) + "]" + prettyPrint(t.Elem())
	case reflect.Struct:
		fqn := t.Name()
		if strings.HasPrefix(fqn, "Property[") {
			typeParam := nameOf(fqn[len("Property[") : len(fqn)-1])
			return fmt.Sprintf("Property[%s]", typeParam)
		}

		return nameOf(fqn)
	default:
		return nameOf(t.Name())
	}
}

func prettyPrintTypescript(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Slice:
		return prettyPrintTypescript(t.Elem()) + "[]"
	case reflect.Ptr:
		return prettyPrintTypescript(t.Elem())
	case reflect.Map:
		return "map<" + prettyPrintTypescript(t.Key()) + "," + prettyPrintTypescript(t.Elem()) + ">"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Struct:
		fqn := t.Name()
		if strings.HasPrefix(fqn, "Property[") {
			typeParam := nameOf(fqn[len("Property[") : len(fqn)-1])
			switch typeParam {
			case "SVGSrc":
				typeParam = "string"
			case "Color":
				typeParam = "string"
			case "Ptr":
				typeParam = "number"
			case "bool":
				typeParam = "boolean"
			case "RequestId":
				typeParam = "number"
			}

			return fmt.Sprintf("Property<%s>", typeParam)
		}
		n := nameOf(fqn)
		switch n {
		case "Color":
			return "string"
		}

		return n
	default:
		return nameOf(t.Name())
	}
}

func nameOf(s string) string {
	idx := strings.LastIndex(s, ".")
	if idx < 0 {
		return s
	}

	return s[idx+1:]

}
