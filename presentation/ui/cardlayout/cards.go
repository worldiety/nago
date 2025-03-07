package cardlayout

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

type sizeClassColumns struct {
	SizeClass core.WindowSizeClass
	Columns   int
}

type TCardLayout struct {
	children       []core.View
	frame          ui.Frame
	padding        ui.Padding
	customColumns  []sizeClassColumns
	defaultColumns int
}

func Layout(children ...core.View) TCardLayout {
	tmp := make([]core.View, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}

		tmp = append(tmp, child)
	}

	return TCardLayout{children: tmp, defaultColumns: 3}
}

func (c TCardLayout) Columns(class core.WindowSizeClass, columns int) TCardLayout {
	c.customColumns = append(c.customColumns, sizeClassColumns{
		SizeClass: class,
		Columns:   columns,
	})

	return c
}

func (c TCardLayout) Frame(frame ui.Frame) TCardLayout {
	c.frame = frame
	return c
}

func (c TCardLayout) Padding(padding ui.Padding) TCardLayout {
	c.padding = padding
	return c
}

func (c TCardLayout) Render(ctx core.RenderContext) core.RenderNode {
	columns := c.defaultColumns
	wnd := ctx.Window()
	if len(c.customColumns) > 0 {
		for _, column := range c.customColumns {
			if column.SizeClass == wnd.Info().SizeClass {
				columns = column.Columns
				break
			}
		}
	} else {
		switch wnd.Info().SizeClass {
		case core.SizeClassSmall:
			columns = 1
		case core.SizeClassMedium:
			columns = 1
		case core.SizeClassLarge:
			columns = 2
		}
	}

	return ui.Grid(
		ui.Each(slices.Values(c.children), func(t core.View) ui.TGridCell {
			return ui.GridCell(t)
		})...,
	).Gap(ui.L16).
		Columns(columns).
		Frame(c.frame).
		Padding(c.padding).
		Render(ctx)
}

type TitleStyle int

const (
	TitleLarge TitleStyle = iota
	TitleCompact
)

type TCard struct {
	title         string
	body          core.View
	footer        core.View
	frame         ui.Frame
	style         TitleStyle
	padding       ui.Padding
	customPadding bool
}

func Card(title string) TCard {
	return TCard{title: title, padding: ui.Padding{Right: ui.L40, Left: ui.L40, Bottom: ui.L40, Top: ""}}
}

func (c TCard) Style(style TitleStyle) TCard {
	c.style = style
	return c
}

func (c TCard) Body(view core.View) TCard {
	c.body = view
	return c
}

func (c TCard) Padding(padding ui.Padding) TCard {
	c.customPadding = true
	c.padding = padding
	return c
}

func (c TCard) Footer(view core.View) TCard {
	c.footer = view
	return c
}

func (c TCard) Frame(frame ui.Frame) TCard {
	c.frame = frame
	return c
}

func (c TCard) Render(ctx core.RenderContext) core.RenderNode {
	var bodyTopPadding ui.Length
	var title core.View
	if c.title != "" {
		switch c.style {
		case TitleCompact:
			bodyTopPadding = ui.L16
			title = ui.HStack(
				ui.Text(c.title),
			).
				Alignment(ui.Leading).
				BackgroundColor(ui.ColorCardTop).
				Padding(ui.Padding{}.All(ui.L8)).
				Frame(ui.Frame{Height: c.padding.Left}.FullWidth())
		default:
			title = ui.VStack(
				ui.Text(c.title).Font(ui.Title),
				ui.HLineWithColor(ui.ColorAccent),
			).Padding(ui.Padding{Top: ui.L40, Left: c.padding.Left, Right: c.padding.Right})
		}
	}

	padding := c.padding
	if !c.customPadding {
		padding.Top = bodyTopPadding
		if padding.Top == "" && c.title == "" {
			padding.Top = padding.Left
		}
	}

	var border ui.Border
	if padding.Left != "" || padding.Bottom != "" {
		border = ui.Border{}.Radius(ui.L16)
	}

	return ui.VStack(
		title,
		ui.VStack(c.body).
			Alignment(ui.Leading).
			FullWidth().
			Padding(padding),
		ui.Spacer(),
		ui.If(c.footer != nil, ui.HStack(c.footer).
			FullWidth().
			Alignment(ui.Trailing).
			BackgroundColor(ui.ColorCardFooter).
			Padding(ui.Padding{}.All(ui.L12))),
	).Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardBody).
		Border(border).
		Frame(c.frame).
		Render(ctx)
}
