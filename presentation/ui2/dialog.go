package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TDialog struct {
	uri    ora.URI
	dlg    ora.VStack
	body   core.View
	footer core.View
	title  core.View
}

func Dialog(body core.View) TDialog {
	return TDialog{
		body: body,
	}
}

func (c TDialog) Title(title core.View) TDialog {
	c.title = title
	return c
}

func (c TDialog) Footer(footer core.View) TDialog {
	c.footer = footer
	return c
}

func (c TDialog) Render(ctx core.RenderContext) ora.Component {
	colors := core.ColorSet[ora.Colors](ctx.Window())
	dlg := Box(BoxLayout{Center: VStack(
		If(c.title != nil, HStack(c.title).Alignment(ora.Leading).BackgroundColor(colors.M4).Frame(ora.Frame{}.FullWidth()).Padding(ora.Padding{Left: ora.L20, Top: ora.L12, Bottom: ora.L12})),
		VStack(
			c.body,
			If(c.footer != nil, HLine()),
			HStack(c.footer).Alignment(ora.Trailing).Frame(ora.Frame{}.FullWidth()),
		).
			Frame(ora.Frame{MaxWidth: ora.L400}.FullWidth()).
			Padding(ora.Padding{Left: ora.L20, Top: ora.L16, Right: ora.L20, Bottom: ora.L20}),
	).
		BackgroundColor(colors.M1).
		Border(ora.Border{}.Radius(ora.L20).Elevate(4)).
		Frame(ora.Frame{MinWidth: ora.L400})},
	).
		BackgroundColor(ora.Color("#000000").WithTransparency(40))

	return dlg.Render(ctx)
}
