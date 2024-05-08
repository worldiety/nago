package main

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
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
		reflect.TypeOf(ora.Ptr(0)): {
			Name:    "Pointer",
			Package: prefix + "pointer",
		},
		reflect.TypeOf(ora.ComponentFactoryId("")): {
			Name:    "ComponentFactoryId",
			Package: prefix + "componentFactoryId",
		},
		reflect.TypeOf(ora.ComponentType("")): {
			Name:    "ComponentType",
			Package: prefix + "componentType",
		},
		reflect.TypeOf(ora.Property[string]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "string"},
			},
		},
		reflect.TypeOf(ora.Property[int]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "number"},
			},
		},
		reflect.TypeOf(ora.Property[float64]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "number"},
			},
		},
		reflect.TypeOf(ora.Property[int64]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "number"},
			},
		},
		reflect.TypeOf(ora.Property[ora.Component]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "Component", Package: genPrefix + "component"},
			},
		},
		reflect.TypeOf(ora.Property[ora.Button]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "Button", Package: genPrefix + "button"},
			},
		},
		reflect.TypeOf(ora.Property[[]ora.Component]{}): {
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
		reflect.TypeOf(ora.Property[[]ora.GridCell]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "GridCell", Package: genPrefix + "gridCell"},
					},
				},
			},
		},
		reflect.TypeOf(ora.Property[[]int64]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "number"},
					},
				},
			},
		},
		reflect.TypeOf(ora.Property[[]ora.BreadcrumbItem]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "BreadcrumbItem", Package: genPrefix + "breadcrumbItem"},
					},
				},
			},
		},
		reflect.TypeOf(ora.Property[[]ora.DropdownItem]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "DropdownItem", Package: genPrefix + "dropdownItem"},
					},
				},
			},
		},
		reflect.TypeOf(ora.Property[[]ora.StepInfo]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{
					Name: "[]",
					TypeParams: []*TSTypeDef{
						{Name: "StepInfo", Package: genPrefix + "stepInfo"},
					},
				},
			},
		},
		reflect.TypeOf(ora.Property[[]ora.Button]{}): {
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

		reflect.TypeOf(ora.Property[[]ora.TableCell]{}): {
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
		reflect.TypeOf(ora.Property[[]ora.TableRow]{}): {
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
		reflect.TypeOf(ora.Property[ora.Intent]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "intent", Name: "Intent"},
			},
		},
		reflect.TypeOf(ora.Property[ora.Ptr]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "pointer", Name: "Pointer"},
			},
		},
		reflect.TypeOf(ora.Property[bool]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Name: "boolean"},
			},
		},
		reflect.TypeOf(ora.Property[ora.RIDSVG]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "ridsvg", Name: "RIDSVG"},
			},
		},

		reflect.TypeOf(ora.Property[ora.SVG]{}): {
			Name:    "Property",
			Package: prefix + "property",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "svg", Name: "SVG"},
			},
		},
		reflect.TypeOf(ora.RequestId(0)): {
			Name:    "RequestId",
			Package: prefix + "requestId",
		},
		reflect.TypeOf((*ora.Component)(nil)).Elem(): {
			Name:    "Component",
			Package: genPrefix + "component",
		},
		reflect.TypeOf([]ora.Event{}): {
			Name: "[]",
			TypeParams: []*TSTypeDef{
				{Package: genPrefix + "event", Name: "Event"},
			},
		},
		reflect.TypeOf(map[string]string{}): {
			Name: "Record",
			TypeParams: []*TSTypeDef{
				{Name: "string"},
				{Name: "string"},
			},
		},

		reflect.TypeOf(ora.Themes{}): {
			Name:    "Themes",
			Package: genPrefix + "themes",
		},
		reflect.TypeOf(ora.Resources{}): {
			Name:    "Resources",
			Package: genPrefix + "resources",
		},
		reflect.TypeOf(ora.Theme{}): {
			Name:    "Theme",
			Package: genPrefix + "theme",
		},
		reflect.TypeOf(ora.ColorScheme("")): {
			Name:    "ColorScheme",
			Package: prefix + "colorScheme",
		},
		reflect.TypeOf(ora.Color{}): {
			Name:    "Color",
			Package: prefix + "color",
		},
		reflect.TypeOf(ora.Colors{}): {
			Name:    "Colors",
			Package: genPrefix + "colors",
		},
		reflect.TypeOf(map[ora.RIDSVG]ora.SVG{}): {
			Name: "Map",
			TypeParams: []*TSTypeDef{
				{Package: prefix + "ridsvg", Name: "RIDSVG"},
				{Package: prefix + "svg", Name: "SVG"},
			},
		},
	}

	// also use the components "enumeration"
	for _, component := range ora.Components {
		name := simpleName(component)
		resolvedTypes[component] = &TSTypeDef{
			Name:    name,
			Package: genPrefix + toLowerFirstChar(name),
		}
	}

	// also use the events "enumeration"
	for _, event := range ora.Events {
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
