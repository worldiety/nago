package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TWindowTitle struct {
	title string
}

func WindowTitle(title string) TWindowTitle {
	return TWindowTitle{title: title}
}

func (c TWindowTitle) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.WindowTitle{
		Value: proto.Str(c.title),
	}
}

func H1(title string) core.View {
	return Heading(1, title)
}

func H2(title string) core.View {
	return Heading(2, title)
}

// Heading returns a default formatted heading text. Level 1 is page heading H1 and so forth. H1 levels also
// set automatically the window title.
func Heading(level int, title string) core.View {
	switch level {
	case 1:
		return VStack(
			WindowTitle(title),
			Text(title).Font(Font{
				Size:   "2rem",
				Weight: BoldFontWeight,
			}),
			HLineWithColor(ColorAccent),
		).Alignment(Leading).Padding(Padding{Bottom: Length("2rem")})
	case 2:
		return VStack(
			Text(title).Font(Font{
				Size:   "1.2rem",
				Weight: BoldFontWeight,
			}),
			HLine(),
		).Alignment(Leading).Padding(Padding{Bottom: Length("2rem")})
	case 3:
		return VStack(Text(title).Font(Title))
	default:
		return VStack(Text(title).Font(SubTitle))
	}
}
