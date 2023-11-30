package ui

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
)

func marshalComponent(c LiveComponent) jsonComponent {
	if c == nil {
		return nil
	}
	tmp := map[CID]bool{}
	compo := newJsonComponent(tmp, c)
	//	x, _ := json.Marshal(compo)
	//	fmt.Println(string(x))
	return compo
}

type jsonComponent = map[string]any

func newJsonComponent(visited map[CID]bool, c LiveComponent) jsonComponent {
	if c == nil {
		return nil
	}

	if visited[c.ID()] {
		panic(fmt.Errorf("found a component cycle for ID %T@%d: component tree must be cycle free", c, c.ID()))
	}

	visited[c.ID()] = true

	obj := jsonComponent{}
	obj["id"] = int64(c.ID())
	obj["type"] = c.Type()
	c.Properties().Each(func(idx int, v Property) {
		ignoreProp := false
		var value any
		switch t := v.Value().(type) {
		case LiveComponent:
			value = newJsonComponent(visited, t)
		case slice.Slice[LiveComponent]:
			var tmp []jsonComponent
			t.Each(func(idx int, component LiveComponent) {
				tmp = append(tmp, newJsonComponent(visited, component))
			})
			value = tmp
		case *Func:
			value = int64(t.ID())
			if t.Nil() {
				ignoreProp = true
			}
		default:
			value = v.Value()
		}

		if !ignoreProp {
			obj[v.Name()] = jsonProperty{
				ID:    int64(v.ID()),
				Type:  propertyTypeName(v),
				Name:  v.Name(),
				Value: value,
			}
		}
	})

	return obj
}

type jsonProperty struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func propertyTypeName(p Property) string {
	switch p.Value().(type) {
	case string:
		return "string"
	case int64:
		return "int"
	case bool:
		return "bool"
	case float64:
		return "float"
	case SVGSrc:
		return "svg"
	case IntentColor:
		return "intentColor"
	case *Func:
		return "func"
	case slice.Slice[LiveComponent], slice.Slice[*Button]:
		return "componentList"
	case LiveComponent:
		return p.Value().(LiveComponent).Type()
	default:
		panic(fmt.Errorf("type not implemented: %T", p.Value()))
	}
}
