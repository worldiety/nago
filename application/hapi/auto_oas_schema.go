// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"go.wdy.de/nago/pkg/oas/v31"
	"log/slog"
	"mime/multipart"
	"net/url"
	"reflect"
	"strings"
	"time"
)

func schemaOf[T any](doc *oas.OpenAPI) *oas.Schema {
	return schemaOfT(doc, reflect.TypeFor[T]())
}

func schemaOfT(doc *oas.OpenAPI, t reflect.Type) *oas.Schema {

	switch t.Kind() {
	case reflect.Ptr:
		tmp := &oas.Schema{
			GoPkgName: t.PkgPath(),
			GoName:    t.Name(),
		}

		if s := doc.ComponentsSchemas()[tmp.RefPlainName()]; s != nil {
			return &oas.Schema{
				Ref: tmp.RefName(),
			}
		}

		// this may be either a normal hierarchy, but at worst, we have a recursion of the same type
		res := schemaOfT(doc, t.Elem())
		if res == nil {
			return nil
		}

		return &oas.Schema{
			Ref: res.RefName(),
		}
	case reflect.Struct:
		switch t {
		case reflect.TypeFor[time.Time]():
			return &oas.Schema{
				Type:    "string",
				Format:  "date-time",
				Example: time.RFC3339,
			}
		case reflect.TypeFor[url.URL]():
			return &oas.Schema{
				Type:   "string",
				Format: "uri",
			}
		}

		tmp := &oas.Schema{
			GoPkgName: t.PkgPath(),
			GoName:    t.Name(),
		}

		if s := doc.ComponentsSchemas()[tmp.RefPlainName()]; s != nil {
			return &oas.Schema{
				Ref: tmp.RefName(),
			}
		}

		// ensure a valid forward declaration to break recursion
		doc.ComponentsSchemas()[tmp.RefPlainName()] = tmp

		res := schemaOfStruct(doc, t)
		if res.Ref == "" {
			doc.ComponentsSchemas()[res.RefPlainName()] = res
		}

		return res
	case reflect.String:
		return &oas.Schema{
			Type: "string",
		}
	case reflect.Int, reflect.Int64:
		return &oas.Schema{
			Type:   "number",
			Format: "int64",
		}
	case reflect.Int32, reflect.Uint32, reflect.Uint8, reflect.Uint16, reflect.Int8, reflect.Int16:
		return &oas.Schema{
			Type:   "number",
			Format: "int32",
		}
	case reflect.Bool:
		return &oas.Schema{
			Type:   "boolean",
			Format: "boolean",
		}
	case reflect.Map:
		switch t {
		case reflect.TypeFor[map[string]string]():
			return &oas.Schema{
				Type: "object",
				AdditionalProperties: &oas.Schema{
					Type: "string",
				},
			}
		}
	case reflect.Slice:
		switch t {
		case reflect.TypeFor[[]multipart.File]():
			//type: array
			//            items:
			//              type: string
			//              format: binary
			return &oas.Schema{
				Type: "array",
				Items: &oas.Schema{
					Type:   "string",
					Format: "binary",
				},
			}
		default:
			foreignRef := schemaOfT(doc, t.Elem())
			if foreignRef != nil && foreignRef.GoName != "" {
				return &oas.Schema{
					Type: "array",
					Items: &oas.Schema{
						Ref: foreignRef.RefName(),
					},
				}
			}

			return &oas.Schema{
				Type:  "array",
				Items: foreignRef,
			}
		}
	}

	slog.Error("cannot generate schema type", "type", t)
	return nil
}

func schemaOfStruct(doc *oas.OpenAPI, t reflect.Type) *oas.Schema {
	res := &oas.Schema{
		GoPkgName:  t.PkgPath(),
		GoName:     t.Name(),
		Properties: map[string]*oas.Schema{},
		Type:       "object",
	}

	for _, field := range reflect.VisibleFields(t) {
		if !field.IsExported() {
			continue
		}

		scheme := schemaOfT(doc, field.Type)
		if scheme == nil {
			slog.Error("failed to detect schema of field type", "type", field.Type, "field", field.Name)
			continue
		}

		scheme.Description = description(field)
		scheme.Required = append(scheme.Required, required(field)...)
		if ex, ok := example(field); ok {
			scheme.Example = ex
		}
		res.Properties[name(field)] = scheme
	}

	doc.ComponentsSchemas()[res.RefPlainName()] = res

	return &oas.Schema{
		Ref: res.RefName(),
	}
}

func name(f reflect.StructField) string {
	if v, ok := f.Tag.Lookup("json"); ok {
		return strings.Split(v, ",")[0]
	}

	return f.Name
}

func required(f reflect.StructField) []string {
	if v := f.Tag.Get("required"); v == "true" {
		return []string{name(f)}
	}

	return nil
}

func description(f reflect.StructField) string {
	if v, ok := f.Tag.Lookup("doc"); ok {
		return v
	}

	return f.Tag.Get("supportingText")
}

func example(f reflect.StructField) (string, bool) {
	if v, ok := f.Tag.Lookup("example"); ok {
		return v, true
	}

	return "", false
}
