package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
)

type RenderState struct {
	funcs    map[ora.Ptr]*Func
	props    map[ora.Ptr]Property
	elements map[ora.Ptr]Component
	visited  map[Component]bool
	//uploads   map[ui.UploadToken]*ui.FileField
	//downloads map[ui.DownloadToken]func() (io.Reader, error)
}

func NewRenderState() *RenderState {
	return &RenderState{
		funcs:    make(map[ora.Ptr]*Func),
		props:    make(map[ora.Ptr]Property),
		elements: make(map[ora.Ptr]Component),
		//	uploads:   make(map[ui.UploadToken]*ui.FileField),
		//	downloads: make(map[ui.DownloadToken]func() (io.Reader, error)),
		visited: map[Component]bool{},
	}
}

func (r *RenderState) Clear() {
	clear(r.funcs)
	clear(r.props)
	clear(r.elements)
	//clear(r.downloads)
	clear(r.visited)
}

func (r *RenderState) Scan(c Component) {
	Visit(c)(func(c Component) bool {
		if c == nil {
			panic(fmt.Errorf("visitor received a nil component"))
		}

		if _, ok := r.elements[c.ID()]; ok {
			panic(fmt.Errorf("another component with the same id has already been visited"))
		}

		if _, ok := r.visited[c]; ok {
			panic(fmt.Errorf("component cycle found, graph must be cyclic-free"))
		}

		r.elements[c.ID()] = c
		r.visited[c] = true

		c.Properties(func(property Property) bool {

			r.props[property.ID()] = property

			property.AnyIter(func(a any) bool {

				if fn, ok := a.(*Func); ok {
					r.funcs[fn.ID()] = fn
					//fmt.Printf("registered func %v\n", fn.ID())
				}
				return true
			})

			return true
		})

		return true
	})
}
