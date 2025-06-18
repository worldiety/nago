package hero

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type THero struct {
	teaserImg ui.DecoredView
	title     string
	subtitle  string
	actions   []core.View
	frame     ui.Frame
}

func Hero(title string) THero {
	return THero{
		title: title,
		frame: ui.Frame{Height: ui.L320}.FullWidth(),
	}
}

func (c THero) Frame(frame ui.Frame) THero {
	c.frame = frame
	return c
}

func (c THero) Actions(actions ...core.View) THero {
	c.actions = actions
	return c
}

func (c THero) Subtitle(subtitle string) THero {
	c.subtitle = subtitle
	return c
}

func (c THero) Teaser(img ui.DecoredView) THero {
	c.teaserImg = img
	return c
}

func (c THero) Render(ctx core.RenderContext) core.RenderNode {
	winfo := ctx.Window().Info()
	small := winfo.SizeClass.Ordinal() <= core.SizeClassSmall.Ordinal()
	var heroTextWidth ui.Length

	if small {
		heroTextWidth = ui.Full
	} else {
		heroTextWidth = "70%"
	}

	return ui.HStack(
		ui.VStack(
			ui.Text(c.title).Font(ui.Title),
			ui.Text(c.subtitle),
			ui.HStack(c.actions...).FullWidth().Alignment(ui.Leading),
		).Alignment(ui.Leading).
			Gap(ui.L16).
			Padding(ui.Padding{}.All(ui.L32)).
			Frame(ui.Frame{Width: heroTextWidth}),
		ui.IfFunc(c.teaserImg != nil, func() core.View {
			if small {
				return nil
			}
			return ui.VStack(
				c.teaserImg.Frame(ui.Frame{Width: ui.Full, Height: ui.Full}),
			).Alignment(ui.Stretch).Frame(ui.Frame{Width: "30%", Height: "100%"})
		}),
	).
		BackgroundColor(ui.ColorCardFooter).
		Frame(c.frame).
		Border(ui.Border{}.Radius(ui.L16)).
		Render(ctx)
}
