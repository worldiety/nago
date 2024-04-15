package main

import (
	"fmt"
	"go.wdy.de/nago/presentation/protocol"
	"log/slog"
	"reflect"
	"strings"
)

var resolvedTypes = map[reflect.Type]*TSTypeDef{}

func init() {
	const prefix = "@/shared/protocol/"
	const genPrefix = prefix + "gen/"

	// oops, we have a reflection limitation here, see
	// https://github.com/golang/go/issues/54393
	// alternatively, we could implement a real AST transpiler
	// but for now, we keep this simple hardcoded reflection hack

	resolvedTypes = map[reflect.Type]*TSTypeDef{
		reflect.TypeOf([]string{}): {
			Name: "string[]",
		},
		reflect.TypeOf(""): {
			Name: "string",
		},
		reflect.TypeOf(protocol.Ptr(0)): {
			Name:    "Pointer",
			Package: prefix + "pointer",
		},
		reflect.TypeOf(protocol.ComponentFactoryId("")): {
			Name:    "ComponentFactoryId",
			Package: prefix + "componentFactoryId",
		},
		reflect.TypeOf(protocol.ComponentType("")): {
			Name:    "ComponentType",
			Package: prefix + "componentType",
		},
		reflect.TypeOf(protocol.Property[string]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "string"},
			},
		},
		reflect.TypeOf(protocol.Property[protocol.Component]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "Component", Package: genPrefix + "component"},
			},
		},
		reflect.TypeOf(protocol.Property[protocol.Button]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "Button", Package: genPrefix + "button"},
			},
		},
		reflect.TypeOf(protocol.Property[[]protocol.Component]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "Component", Package: genPrefix + "component"},
					},
				},
			},
		},
		reflect.TypeOf(protocol.Property[[]protocol.Button]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "Button", Package: genPrefix + "button"},
					},
				},
			},
		},

		reflect.TypeOf(protocol.Property[[]protocol.TableCell]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "TableCell", Package: genPrefix + "tableCell"},
					},
				},
			},
		},
		reflect.TypeOf(protocol.Property[[]protocol.TableRow]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "TableRow", Package: genPrefix + "tableRow"},
					},
				},
			},
		},
		reflect.TypeOf(protocol.Property[protocol.Intent]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "intent", Name: "Intent"},
			},
		},
		reflect.TypeOf(protocol.Property[protocol.Ptr]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "pointer", Name: "Pointer"},
			},
		},
		reflect.TypeOf(protocol.Property[bool]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "boolean"},
			},
		},
		reflect.TypeOf(protocol.Property[protocol.RIDSVG]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "ridsvg", Name: "RIDSVG"},
			},
		},

		reflect.TypeOf(protocol.Property[protocol.SVG]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "svg", Name: "SVG"},
			},
		},
		reflect.TypeOf(protocol.RequestId(0)): {
			Name:    "RequestId",
			Package: prefix + "requestId",
		},
		reflect.TypeOf((*protocol.Component)(nil)).Elem(): {
			Name:    "Component",
			Package: genPrefix + "component",
		},
		reflect.TypeOf([]protocol.Event{}): {
			Name: "[]",
			TypeParams: []*TSTypeDef{
				{Package: genPrefix + "event", Name: "Event"},
			},
		},
		reflect.TypeOf(map[string]string{}): {
			Name: "Map",
			TypeParams: []*TSTypeDef{
				{Name: "string"},
				{Name: "string"},
			},
		},

		reflect.TypeOf(protocol.Themes{}): {
			Name:    "Themes",
			Package: genPrefix + "themes",
		},
		reflect.TypeOf(protocol.Resources{}): {
			Name:    "Resources",
			Package: genPrefix + "resources",
		},
		reflect.TypeOf(protocol.Theme{}): {
			Name:    "Theme",
			Package: genPrefix + "theme",
		},
		reflect.TypeOf(protocol.ColorScheme("")): {
			Name:    "ColorScheme",
			Package: prefix + "colorScheme",
		},
		reflect.TypeOf(protocol.Color{}): {
			Name:    "Color",
			Package: prefix + "color",
		},
		reflect.TypeOf(protocol.Colors{}): {
			Name:    "Colors",
			Package: genPrefix + "colors",
		},
		reflect.TypeOf(map[protocol.RIDSVG]protocol.SVG{}): {
			Name: "Map",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "ridsvg", Name: "RIDSVG"},
				{Package: prefix + "svg", Name: "SVG"},
			},
		},
	}

	// also use the components "enumeration"
	for _, component := range protocol.Components {
		name := simpleName(component)
		resolvedTypes[component] = &TSTypeDef{
			Name:    name,
			Package: genPrefix + toLowerFirstChar(name),
		}
	}

	// also use the events "enumeration"
	for _, event := range protocol.Events {
		name := simpleName(event)
		resolvedTypes[event] = &TSTypeDef{
			Name:    name,
			Package: genPrefix + toLowerFirstChar(name),
		}
	}
}

func resolveType(t reflect.Type) *TSTypeDef {
	def, ok := resolvedTypes[t]
	if ok {
		return def
	}

	slog.Error(fmt.Sprintf("go type to typescript resolving not found: %v", t))

	return &TSTypeDef{
		Package: "?",
		Name:    "?",
	}
}

func simpleName(t reflect.Type) string {
	name := t.Name()
	idx := strings.LastIndex(name, ".")
	if idx > 0 {
		name = name[idx+1:]
	}

	return name
}
