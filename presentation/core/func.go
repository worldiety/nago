package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/protocol"
)

// Func is a remote addressable function holder.
type Func struct {
	name  string
	f     func()
	id    protocol.Ptr
	dirty bool
}

func NewFunc(name string) *Func {
	return &Func{
		name: name,
		id:   NextPtr(),
	}
}

func (lf *Func) Unwrap() any {
	return lf // ??? TODO this may cause a visitor endless recursion
}

func (lf *Func) setValue(v string) error {
	return lf.Parse(v)
}

func (lf *Func) Parse(v string) error {
	return fmt.Errorf("cannot set function pointer by string: %s", v)
}

func (lf *Func) Set(f func()) {
	if f == nil {
		lf.Clear()
		return
	}

	lf.f = f
	lf.dirty = true
}

func (lf *Func) SetDirty(b bool) {
	lf.dirty = b
}

func (lf *Func) Dirty() bool {
	return lf.dirty
}

func (lf *Func) ID() protocol.Ptr {
	return lf.id
}

func (lf *Func) Invoke() {
	lf.f()
}

func (lf *Func) Name() string {
	return lf.name
}

func (lf *Func) Clear() {
	lf.id = 0
	lf.f = nil
}

func (lf *Func) Nil() bool {
	return lf == nil || lf.id.Nil() || lf.f == nil
}

func (lf *Func) Iter(yield func(*Func) bool) {
	yield(lf)
}
