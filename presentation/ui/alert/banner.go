package alert

import (
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

type TBanner struct {
	title     string
	message   string
	presented *core.State[bool]
	onClosed  func()
}

func Banner(title, message string) TBanner {
	return TBanner{
		title:   title,
		message: message,
	}
}

func (t TBanner) Closeable(presented *core.State[bool]) TBanner {
	t.presented = presented
	return t
}

func (t TBanner) OnClosed(fn func()) TBanner {
	t.onClosed = fn
	return t
}

func (t TBanner) Render(ctx core.RenderContext) core.RenderNode {
	if t.presented != nil && !t.presented.Get() {
		return ui.HStack().Render(ctx)
	}

	const textColor = "#FF543E"

	// actually the color is "#FF543E" however, we don't want transparency
	var bgColor ui.Color
	if ctx.Window().Info().ColorScheme == core.Dark {
		bgColor = "#3b1812"
	} else {
		bgColor = "#F6d2de"
	}

	return ui.VStack(
		ui.HStack(
			ui.Image().
				FillColor(textColor).
				Embed(heroSolid.ExclamationTriangle).
				Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
			ui.Text(t.title).
				Font(ui.SubTitle).
				Color(textColor),
			ui.Spacer(),
			ui.If(t.presented != nil, ui.HStack(ui.Image().
				Embed(heroSolid.XMark).
				FillColor(textColor).
				Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
			).Action(func() {
				t.presented.Set(false)
				if t.onClosed != nil {
					t.onClosed()
				}
			})),
		).Gap(ui.L4).
			FullWidth(),
		ui.Text(t.message).Color(textColor),
	).Alignment(ui.Leading).
		Gap(ui.L8).
		BackgroundColor(bgColor).
		Border(ui.Border{}.Radius(ui.L12)).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(ui.Frame{Width: ui.L400}).Render(ctx)
}
