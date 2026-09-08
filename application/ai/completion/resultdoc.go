// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// maxResultDocDepth bounds how deep a rendered result description descends into nested types. Beyond it the
// shape is summarised, because a model that needs four levels of nesting spelled out is better served by
// looking at the first actual result.
const maxResultDocDepth = 4

// renderResultDoc turns the return type of a tool into a compact description a model can read.
//
// It exists because there is nowhere else to put it: no provider accepts an output schema for a tool. The
// tool definition carries a name, a description and an input schema, and nothing more (see the Anthropic and
// OpenAI-compatible adapters). So the only way to tell a model what it will get back is to say so in the
// description text.
//
// That is also why this is opt-in via [Tool.WithResultDoc] rather than automatic: the text is sent on every
// single request, for every tool, for the whole conversation. Where field names already say what they are,
// the model learns the shape from the first actual result at no cost at all.
func renderResultDoc(t reflect.Type) string {
	body := describeType(t, "", 0)
	if body == "" {
		return ""
	}

	return "Returns " + body
}

// describeType renders one type. Structs and slices of structs are expanded; everything else is named.
// indent is prepended to every line after the first, so nesting stays readable.
func describeType(t reflect.Type, indent string, depth int) string {
	t = deref(t)
	if t == nil {
		return ""
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return "a string (base64 encoded bytes)."
		}

		elem := deref(t.Elem())
		if isExpandable(elem) && depth < maxResultDocDepth {
			return "a list of objects with:\n" + describeFields(elem, indent, depth+1)
		}

		return fmt.Sprintf("a list of %s.", scalarName(elem))

	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "a timestamp (RFC 3339)."
		}

		if depth >= maxResultDocDepth {
			return "an object."
		}

		fields := describeFields(t, indent, depth+1)
		if fields == "" {
			return "an object."
		}

		return "an object with:\n" + fields

	case reflect.Map:
		return fmt.Sprintf("an object whose values are %s.", scalarName(deref(t.Elem())))

	default:
		return fmt.Sprintf("%s.", scalarName(t))
	}
}

// describeFields renders the exported fields of a struct as a bullet list, in declaration order.
//
// Declaration order rather than sorted: it is stable across builds - which matters, because an unstable
// description would defeat provider-side prompt caching - and it is the order a reader of the Go type sees,
// which is usually the order that makes sense.
func describeFields(t reflect.Type, indent string, depth int) string {
	var sb strings.Builder

	for i := range t.NumField() {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous { // unexported, non-embedded
			continue
		}

		name, omitempty, skip := parseJSONField(f)
		if skip {
			continue
		}

		// Embedded structs without an explicit name are promoted, exactly as the schema does it.
		if f.Anonymous && name == "" {
			if ft := deref(f.Type); ft.Kind() == reflect.Struct {
				if promoted := describeFields(ft, indent, depth); promoted != "" {
					sb.WriteString(promoted)
					sb.WriteString("\n")
				}
				continue
			}
		}

		if name == "" {
			name = f.Name
		}

		sb.WriteString(indent)
		sb.WriteString("- ")
		sb.WriteString(name)

		nested := deref(f.Type)
		isNested := isExpandable(nested) ||
			(nested.Kind() == reflect.Slice && isExpandable(deref(nested.Elem())))

		if isNested && depth < maxResultDocDepth {
			// Expand in place. The nested bullets are indented one level further so the structure is
			// visible without the model having to guess which field a bullet belongs to.
			if desc := fieldDescription(f); desc != "" {
				sb.WriteString(": ")
				sb.WriteString(desc)
				sb.WriteString(" — ")
			} else {
				sb.WriteString(" is ")
			}

			sb.WriteString(describeType(f.Type, indent+"  ", depth))
			sb.WriteString("\n")

			continue
		}

		sb.WriteString(" (")
		sb.WriteString(typeName(f.Type))
		if !omitempty && f.Type.Kind() != reflect.Pointer {
			// Only stated for results, where it means "always present" rather than "you must send it".
			sb.WriteString(", always present")
		}
		sb.WriteString(")")

		if desc := fieldDescription(f); desc != "" {
			sb.WriteString(": ")
			sb.WriteString(desc)
		}

		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// isExpandable reports whether a type is a struct worth spelling out rather than naming.
func isExpandable(t reflect.Type) bool {
	return t != nil && t.Kind() == reflect.Struct && t != reflect.TypeOf(time.Time{})
}

// deref removes pointer indirection.
func deref(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// typeName names a field type in the words a JSON schema uses.
func typeName(t reflect.Type) string {
	t = deref(t)
	if t == nil {
		return "value"
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		return "list of " + typeName(t.Elem())
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "timestamp, RFC 3339"
		}
		return "object"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Map, reflect.Interface:
		return "object"
	default:
		return "value"
	}
}

// scalarName names a type in the words a JSON schema uses, with an article.
func scalarName(t reflect.Type) string {
	if t == nil {
		return "a value"
	}

	switch t.Kind() {
	case reflect.Bool:
		return "a boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "an integer"
	case reflect.Float32, reflect.Float64:
		return "a number"
	case reflect.String:
		return "a string"
	case reflect.Map, reflect.Struct, reflect.Interface:
		return "an object"
	default:
		return "a value"
	}
}
