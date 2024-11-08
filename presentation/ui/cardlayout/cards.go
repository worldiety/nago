package cardlayout

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

type TCardLayout struct {
	children []core.View
}

func CardLayout(children ...core.View) TCardLayout {
	tmp := make([]core.View, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}

		tmp = append(tmp, child)
	}

	return TCardLayout{children: tmp}
}

func (c TCardLayout) Render(ctx core.RenderContext) core.RenderNode {
	columns := 3
	wnd := ctx.Window()
	switch wnd.Info().SizeClass {
	case core.SizeClassSmall:
		columns = 1
	case core.SizeClassMedium:
		columns = 1
	case core.SizeClassLarge:
		columns = 2
	}

	return ui.Grid(
		ui.Each(slices.Values(c.children), func(t core.View) ui.TGridCell {
			return ui.GridCell(t)
		})...,
	).Gap(ui.L16).Columns(columns).Render(ctx)
}

type TCard struct {
	title  string
	body   core.View
	footer core.View
}

func Card(title string) TCard {
	return TCard{title: title}
}

func (c TCard) Body(view core.View) TCard {
	c.body = view
	return c
}

func (c TCard) Footer(view core.View) TCard {
	c.footer = view
	return c
}

func (c TCard) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		ui.VStack(
			ui.Text(c.title).Font(ui.Title),
			ui.HLineWithColor(ui.ColorAccent),
		).Padding(ui.Padding{Top: ui.L40, Left: ui.L40, Right: ui.L40}),
		ui.VStack(c.body).
			Padding(ui.Padding{Right: ui.L40, Left: ui.L40, Bottom: ui.L40}),
		ui.Spacer(),
		ui.HStack(c.footer).
			FullWidth().
			Alignment(ui.Trailing).
			BackgroundColor(ui.ColorCardFooter).
			Padding(ui.Padding{}.All(ui.L12)),
	).Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Render(ctx)
}
