package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"log/slog"
	"net/url"
	"strings"
)

type TextAlignment uint

const (
	TextAlignInherit TextAlignment = TextAlignment(proto.TextAlignInherit)
	TextAlignStart   TextAlignment = TextAlignment(proto.TextAlignStart)
	TextAlignEnd     TextAlignment = TextAlignment(proto.TextAlignEnd)
	TextAlignCenter  TextAlignment = TextAlignment(proto.TextAlignCenter)
	TextAlignJustify TextAlignment = TextAlignment(proto.TextAlignJustify)
)

type TText struct {
	content                string
	color                  proto.Color
	backgroundColor        proto.Color
	hoveredBackgroundColor proto.Color
	pressedBackgroundColor proto.Color
	focusedBackgroundColor proto.Color
	font                   proto.Font
	invisible              bool
	onClick                func()

	padding            proto.Padding
	frame              Frame
	border             proto.Border
	hoveredBorder      proto.Border
	focusedBorder      proto.Border
	pressedBorder      proto.Border
	accessibilityLabel string
	textAlignment      TextAlignment
	action             func()
	lineBreak          bool
	underline          bool
}

func MailTo(wnd core.Window, name string, email string) TText {
	return Link(wnd, name, "mailto:"+email, "_parent")
}

func LinkWithAction(text string, action func()) TText {
	return Text(text).Underline(true).Action(action).Color(ColorInteractive)
}

// Link performs a best guess based on the given href. If the href starts with http or https
// the window will perform an Open call. Otherwise, a local forward navigation is applied.
func Link(wnd core.Window, text string, href string, target string) TText {
	return LinkWithAction(text, func() {
		if wnd == nil {
			slog.Error("cannot execute link action: window is nil")
			return
		}

		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			core.HTTPOpen(wnd.Navigation(), core.URI(href), target)
		} else {
			u, err := url.Parse(href)
			if err != nil {
				slog.Error("Failed to parse href link URL", err, href)
				wnd.Navigation().ForwardTo(core.NavigationPath(href), nil)
				return
			}

			q := u.Query()
			if len(q) > 0 {
				tmp := core.Values{}
				for k, v := range q {
					if len(v) > 0 {
						tmp[k] = v[0]
					}
				}
				wnd.Navigation().ForwardTo(core.NavigationPath(u.Path), tmp)
			} else {
				wnd.Navigation().ForwardTo(core.NavigationPath(u.Path), nil)
			}

		}
	})
}

func Text(content string) TText {
	return TText{content: content}
}

func (c TText) Underline(b bool) TText {
	c.underline = b
	return c
}

func (c TText) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TText) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

func (c TText) FullWidth() TText {
	c.frame = Frame{}.FullWidth()
	return c
}

func (c TText) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TText) LineBreak(lb bool) TText {
	c.lineBreak = lb
	return c
}

func (c TText) HoveredBorder(border Border) TText {
	c.hoveredBorder = border.ora()
	return c
}

func (c TText) PressedBorder(border Border) TText {
	c.pressedBorder = border.ora()
	return c
}

func (c TText) FocusedBorder(border Border) TText {
	c.focusedBorder = border.ora()
	return c
}

func (c TText) TextAlignment(align TextAlignment) TText {
	c.textAlignment = align
	return c
}

func (c TText) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

func (c TText) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TText) Font(font Font) TText {
	c.font = font.ora()
	return c
}

func (c TText) Color(color Color) TText {
	c.color = color.ora()
	return c
}

func (c TText) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TText) Action(f func()) TText {
	c.action = f
	return c
}

func (c TText) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.TextView{
		Value:              proto.Str(c.content),
		Color:              c.color,
		BackgroundColor:    c.backgroundColor,
		Font:               c.font,
		OnClick:            ctx.MountCallback(c.onClick),
		Invisible:          proto.Bool(c.invisible),
		Border:             c.border,
		Padding:            c.padding,
		Frame:              c.frame.ora(),
		AccessibilityLabel: proto.Str(c.accessibilityLabel),

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		TextAlignment:          proto.TextAlignment(c.textAlignment),
		Action:                 ctx.MountCallback(c.action),
		LineBreak:              proto.Bool(c.lineBreak),
		Underline:              proto.Bool(c.underline),
	}
}
