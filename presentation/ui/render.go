package ui

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
)

func marshalComponent(rs *renderState, c LiveComponent) jsonComponent {
	if c == nil {
		return nil
	}
	compo := newJsonComponent(rs, c)
	//	x, _ := json.Marshal(compo)
	//	fmt.Println(string(x))
	return compo
}

type jsonComponent = map[string]any

func newJsonComponent(rs *renderState, c LiveComponent) jsonComponent {
	if c == nil {
		return nil
	}

	if rs.visited(c.ID()) {
		panic(fmt.Errorf("found a component cycle for ID %T@%d: component tree must be cycle free", c, c.ID()))
	}

	rs.visit(c.ID(), c)

	obj := jsonComponent{}
	obj["id"] = int64(c.ID())
	obj["type"] = c.Type()
	c.Properties().Each(func(idx int, v Property) {
		var value any
		switch t := v.value().(type) {
		case LiveComponent:
			value = newJsonComponent(rs, t)
			rs.addProp(v)
		case []LiveComponent:
			var tmp []jsonComponent
			for _, component := range t {
				tmp = append(tmp, newJsonComponent(rs, component))
			}
			value = tmp
			rs.addProp(v)
		case slice.Slice[LiveComponent]:
			var tmp []jsonComponent
			t.Each(func(idx int, component LiveComponent) {
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
			value = v.value()
			rs.addProp(v)
		}

		obj[v.Name()] = jsonProperty{
			ID:    int64(v.ID()),
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

type renderState struct {
	funcs    map[CID]*Func
	props    map[CID]Property
	elements map[CID]LiveComponent
	uploads  map[UploadToken]*FileField
}

func newRenderState() *renderState {
	return &renderState{
		funcs:    make(map[CID]*Func),
		props:    make(map[CID]Property),
		elements: make(map[CID]LiveComponent),
		uploads:  make(map[UploadToken]*FileField),
	}
}

func (r *renderState) Clear() {
	clear(r.funcs)
	clear(r.props)
	clear(r.elements)
}

func (r *renderState) visit(id CID, t LiveComponent) {
	r.elements[id] = t
	if fup, ok := t.(*FileField); ok {
		r.uploads[fup.UploadToken()] = fup
	}
}

func (r *renderState) visited(id CID) bool {
	_, ok := r.elements[id]
	return ok
}

func (r *renderState) addFunc(f *Func) {
	r.funcs[f.ID()] = f
}

func (r *renderState) addProp(p Property) {
	r.props[p.ID()] = p
}
