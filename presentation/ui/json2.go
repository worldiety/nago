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
		var value any
		switch t := v.value().(type) {
		case LiveComponent:
			value = newJsonComponent(visited, t)
		case []LiveComponent:
			var tmp []jsonComponent
			for _, component := range t {
				tmp = append(tmp, newJsonComponent(visited, component))
			}
			value = tmp
		case slice.Slice[LiveComponent]:
			var tmp []jsonComponent
			t.Each(func(idx int, component LiveComponent) {
				tmp = append(tmp, newJsonComponent(visited, component))
			})
			value = tmp
		case *Func:
			value = int64(t.ID())
			if t.Nil() {
				value = 0
			}
		default:
			value = v.value()
		}

		obj[v.Name()] = jsonProperty{
			ID: int64(v.ID()),
			//Type:  propertyTypeName(v),
			Name:  v.Name(),
			Value: value,
		}
	})

	return obj
}

type jsonProperty struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}
