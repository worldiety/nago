package ui

import "go.wdy.de/nago/presentation/core"

func marshalComponent(rs *core.RenderState, c core.Component) jsonComponent {
	if c == nil {
		return nil
	}
	compo := newJsonComponent(rs, c)
	//	x, _ := json.Marshal(compo)
	//	fmt.Println(string(x))
	return compo
}

type jsonComponent = map[string]any

func newJsonComponent(rs *core.RenderState, c core.Component) jsonComponent {
	if c == nil {
		return nil
	}
	/*
		if rs.visited(c.ID()) {
			panic(fmt.Errorf("found a component cycle for ID %T@%d: component tree must be cycle free", c, c.ID()))
		}

		rs.visit(c.ID(), c)

		obj := jsonComponent{}
		obj["id"] = int64(c.ID())
		obj["type"] = c.Type()
		c.Properties().Each(func(idx int, v Property) {
			var value any
			switch t := v.Unwrap().(type) {
			case core.Component:
				value = newJsonComponent(rs, t)
				rs.addProp(v)
			case []core.Component:
				var tmp []jsonComponent
				for _, component := range t {
					tmp = append(tmp, newJsonComponent(rs, component))
				}
				value = tmp
				rs.addProp(v)
			case slice.Slice[core.Component]:
				var tmp []jsonComponent
				t.Each(func(idx int, component core.Component) {
					tmp = append(tmp, newJsonComponent(rs, component))
				})
				value = tmp
				rs.addProp(v)
			case *Func:
				value = int64(t.ID())
				if t.Nil() {
					value = 0
				}
				rs.addFunc(t)
			default:
				value = v.Unwrap()
				rs.addProp(v)
			}

			obj[v.Name()] = jsonProperty{
				ID:    int64(v.ID()),
				Name:  v.Name(),
				Value: value,
			}
		})

		return obj
	*/

	return nil
}

type jsonProperty struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}
