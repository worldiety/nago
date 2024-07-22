package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TModal struct {
	content          core.View
	onDismissRequest func()
}

func Modal(content core.View) TModal {
	return TModal{content: content}
}

func (c TModal) Render(context core.RenderContext) ora.Component {
	return ora.Modal{
		Type:             ora.ModalT,
		Content:          c.content.Render(context),
		OnDismissRequest: context.MountCallback(c.onDismissRequest),
	}
}
