package list

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type TEntry struct {
	headline       string
	supportingText string
	leading        core.View
	trailing       core.View
	action         func()
	frame          ui.Frame
}

func Entry() TEntry {
	return TEntry{}.Frame(ui.Frame{}.FullWidth())
}

func (c TEntry) Headline(s string) TEntry {
	c.headline = s
	return c
}

func (c TEntry) SupportingText(s string) TEntry {
	c.supportingText = s
	return c
}

func (c TEntry) Leading(v core.View) TEntry {
	c.leading = v
	return c
}

func (c TEntry) Trailing(v core.View) TEntry {
	c.trailing = v
	return c
}

func (c TEntry) Action(fn func()) TEntry {
	c.action = fn
	return c
}

func (c TEntry) Frame(frame ui.Frame) TEntry {
	c.frame = frame
	return c
}

func (c TEntry) Render(ctx core.RenderContext) core.RenderNode {

	return ui.HStack(
		c.leading,
		ui.VStack(
			ui.If(c.headline != "", ui.Text(c.headline).Font(ui.SubTitle)),
			ui.If(c.supportingText != "", ui.Text(c.supportingText)),
		).Alignment(ui.Leading),
		ui.Spacer(),
		c.trailing,
	).Action(c.action).
		Gap(ui.L16).
		Frame(c.frame).
		Render(ctx)
}

type TList struct {
	caption core.View
	rows    []core.View
	frame   ui.Frame
	footer  core.View
}

func List(entries ...core.View) TList {
	return TList{rows: entries}
}

func (c TList) Caption(s core.View) TList {
	c.caption = s
	return c
}

func (c TList) Frame(frame ui.Frame) TList {
	c.frame = frame
	return c
}

func (c TList) Footer(s core.View) TList {
	c.footer = s
	return c
}

func (c TList) Render(ctx core.RenderContext) core.RenderNode {
	rows := make([]core.View, 0, len(c.rows)*2+3)
	if c.caption != nil {
		rows = append(rows, ui.HStack(c.caption).Alignment(ui.Leading).FullWidth().BackgroundColor(ui.ColorCardTop).Padding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L16)))
	}

	for idx, row := range c.rows {
		rows = append(rows, ui.HStack(row).HoveredBackgroundColor(ui.ColorCardFooter).Padding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L16)).Frame(ui.Frame{}.FullWidth()))
		if idx < len(c.rows)-1 {
			rows = append(rows, ui.HStack(ui.HLine().Padding(ui.Padding{})).FullWidth().Padding(ui.Padding{}.Horizontal(ui.L16)))
		}
	}

	if c.footer != nil {
		rows = append(rows, ui.HStack(c.footer).Alignment(ui.Leading).FullWidth().BackgroundColor(ui.ColorCardFooter).Padding(ui.Padding{}.Vertical(ui.L16).Horizontal(ui.L16)))
	}

	return ui.VStack(rows...).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Frame(c.frame).
		Render(ctx)
}
