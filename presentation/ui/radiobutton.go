package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Radiobutton struct {
	id         ora.Ptr
	selected   Bool
	onClicked  *Func
	disabled   Bool
	visible    Bool
	properties []core.Property
}

func NewRadiobutton(with func(rab *Radiobutton)) *Radiobutton {
	r := &Radiobutton{
		id:        nextPtr(),
		selected:  NewShared[bool]("selected"),
		onClicked: NewFunc("action"),
		disabled:  NewShared[bool]("disabled"),
		visible:   NewShared[bool]("visible"),
	}

	r.properties = []core.Property{r.selected, r.onClicked, r.disabled, r.visible}
	r.visible.Set(true)
	if with != nil {
		with(r)
	}
	return r
}

func (r *Radiobutton) ID() ora.Ptr {
	return r.id
}

func (r *Radiobutton) Properties(yield func(core.Property) bool) {
	for _, property := range r.properties {
		if !yield(property) {
			return
		}
	}
}

func (r *Radiobutton) Render() ora.Component {
	return r.renderRadiobutton()
}

func (r *Radiobutton) Selected() Bool { return r.selected }

func (r *Radiobutton) OnClicked() *Func {
	return r.onClicked
}

func (r *Radiobutton) Disabled() Bool {
	return r.disabled
}

func (r *Radiobutton) Visible() Bool {
	return r.visible
}

func (r *Radiobutton) renderRadiobutton() ora.Radiobutton {
	return ora.Radiobutton{
		Ptr:       r.id,
		Type:      ora.RadiobuttonT,
		Disabled:  r.disabled.render(),
		Selected:  r.selected.render(),
		OnClicked: renderFunc(r.onClicked),
	}
}

func (r *Radiobutton) UpdateRadioButtons(radiobuttons []*Radiobutton, selectedButton *Radiobutton) {
	for _, v := range radiobuttons {
		if v != selectedButton {
			v.Selected().Set(false)
		}
	}
}
