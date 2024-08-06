package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TDialog struct {
	uri    core.URI
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
	colors := core.Colors[Colors](ctx.Window())
	dlg := Box(BoxLayout{Center: VStack(
		If(c.title != nil, HStack(c.title).Alignment(Leading).BackgroundColor(colors.M4).Frame(Frame{}.FullWidth()).Padding(Padding{Left: L20, Top: L12, Bottom: L12})),
		VStack(
			c.body,
			If(c.footer != nil, HLine()),
			HStack(c.footer).Alignment(Trailing).Frame(Frame{}.FullWidth()),
		).
			Frame(Frame{MaxWidth: L400}.FullWidth()).
			Padding(Padding{Left: L20, Top: L16, Right: L20, Bottom: L20}),
	).
		BackgroundColor(colors.M1).
		Border(Border{}.Radius(L20).Elevate(4)).
		Frame(Frame{MinWidth: L400})},
	).
		BackgroundColor(Color("#000000").WithTransparency(40))

	return dlg.Render(ctx)
}
