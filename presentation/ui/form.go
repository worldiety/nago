package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TForm struct {
	children     []core.View
	id           string
	action       func()
	autocomplete bool
}

func Form(children ...core.View) TForm {
	return TForm{
		children: children,
	}
}

func (c TForm) ID(id string) TForm {
	c.id = id
	return c
}

func (c TForm) Action(action func()) TForm {
	c.action = action
	return c
}

func (c TForm) Autocomplete(b bool) TForm {
	c.autocomplete = b
	return c
}

func (c TForm) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Form{
		Children:     renderComponents(ctx, c.children),
		Action:       ctx.MountCallback(c.action),
		Id:           proto.Str(c.id),
		Autocomplete: proto.Bool(c.autocomplete),
	}
}
