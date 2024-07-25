package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TFilledButton struct {
	fillColor ora.Color
	textColor ora.Color
	title     string
	preIcon   ora.SVG
	postIcon  ora.SVG
	frame     ora.Frame
	action    func()
}

func FilledButton(fillColor ora.Color, action func()) TFilledButton {
	return TFilledButton{fillColor: fillColor, action: action}
}

func (c TFilledButton) Title(text string) TFilledButton {
	c.title = text
	return c
}

func (c TFilledButton) PreIcon(svg ora.SVG) TFilledButton {
	c.preIcon = svg
	return c
}

func (c TFilledButton) PostIcon(svg ora.SVG) TFilledButton {
	c.postIcon = svg
	return c
}

func (c TFilledButton) TextColor(color ora.Color) TFilledButton {
	c.textColor = color
	return c
}

func (c TFilledButton) Frame(frame ora.Frame) TFilledButton {
	c.frame = frame
	return c
}

func (c TFilledButton) Render(context core.RenderContext) ora.Component {
	decView := customButton(c.fillColor, HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(ora.Frame{}.Size(ora.L16, ora.L16))),
		If(c.title != "", btnTitle(c.title, c.textColor)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(ora.Frame{}.Size(ora.L16, ora.L16))),
	).Action(c.action))

	var zero ora.Frame
	if c.frame != zero {
		decView = decView.Frame(c.frame)
	}

	return decView.Render(context)
}

func btnTitle(text string, color ora.Color) TText {
	t := Text(text).Font(ora.Font{Size: ora.L14, Weight: 500})
	if color != "" {
		t = t.Color(color)
	}

	return t
}

func customButton(fillColor ora.Color, hstack THStack) core.DecoredView {
	// TODO the pressed+focus logic is (perhaps?) broken in the frontend
	return hstack.
		HoveredBackgroundColor(fillColor.WithTransparency(25)).
		PressedBackgroundColor(fillColor.WithTransparency(35)).
		PressedBorder(ora.Border{}.Circle().Color("#00000000").Width(ora.L2)).
		FocusedBorder(ora.Border{}.Circle().Color("#ffffff").Width(ora.L2)).
		BackgroundColor(fillColor).
		Frame(ora.Frame{Height: "2.375rem"}).
		Padding(ora.Padding{}.Horizontal("1.125rem")).
		// add invisible default border, to avoid dimension changes,
		// note, that we need to fix that with frame and padding above
		Border(ora.Border{}.Circle().Color("#00000000").Width(ora.L2))
}
