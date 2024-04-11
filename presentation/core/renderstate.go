package core

import (
	"go.wdy.de/nago/presentation/protocol"
)

type RenderState struct {
	funcs    map[protocol.Ptr]*Func
	props    map[protocol.Ptr]Property
	elements map[protocol.Ptr]Component
	//uploads   map[ui.UploadToken]*ui.FileField
	//downloads map[ui.DownloadToken]func() (io.Reader, error)
}

func NewRenderState() *RenderState {
	return &RenderState{
		funcs:    make(map[protocol.Ptr]*Func),
		props:    make(map[protocol.Ptr]Property),
		elements: make(map[protocol.Ptr]Component),
		//	uploads:   make(map[ui.UploadToken]*ui.FileField),
		//	downloads: make(map[ui.DownloadToken]func() (io.Reader, error)),
	}
}

func (r *RenderState) Clear() {
	clear(r.funcs)
	clear(r.props)
	clear(r.elements)
	//clear(r.downloads)
}

func (r *RenderState) Visit(id protocol.Ptr, t Component) {
	r.elements[id] = t
	// TODO this is causing a cycle we don't want
	/*if fup, ok := t.(*ui.FileField); ok {
		r.uploads[fup.UploadToken()] = fup
	}

	if ds, ok := t.(ui.DownloadSource); ok && ds.DownloadSource() != nil {
		r.downloads[ds.DownloadToken()] = ds.DownloadSource()
	}*/
}

func (r *RenderState) Visited(id protocol.Ptr) bool {
	_, ok := r.elements[id]
	return ok
}

func (r *RenderState) AddFunc(f *Func) {
	r.funcs[f.ID()] = f
}

func (r *RenderState) AddProp(p Property) {
	r.props[p.ID()] = p
}
