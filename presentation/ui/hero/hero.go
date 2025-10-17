package hero

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type HeroType int

const (
	ImageRight HeroType = iota
	ImageFull
)

// THero is a composite component (Hero).
// This component represents a prominent section with an image,
// title, subtitle, and optional action buttons.
type THero struct {
	teaserImg ui.DecoredView // hero image or visual teaser
	title     string         // main headline text
	subtitle  string         // supporting subtitle text
	actions   []core.View    // list of action buttons or links
	frame     ui.Frame       // layout frame for the hero section
	herotype  HeroType
	alignment ui.Alignment
}

// Hero creates a new THero with the given title and a default full-width height of 320.
func Hero(title string) THero {
	return THero{
		title: title,
		frame: ui.Frame{Height: ui.L320}.FullWidth(),
	}
}

// HeroType sets the herotype of the THero element.
func (c THero) HeroType(t HeroType) THero {
	c.herotype = t
	return c
}

// Alignment sets the Alignment of the THero element.
func (c THero) Alignment(alignment ui.Alignment) THero {
	c.alignment = alignment
	return c
}

// Frame sets the frame of the hero section.
func (c THero) Frame(frame ui.Frame) THero {
	c.frame = frame
	return c
}

// Actions sets the action buttons or links of the hero section.
func (c THero) Actions(actions ...core.View) THero {
	c.actions = actions
	return c
}

// Subtitle sets the subtitle text of the hero section.
func (c THero) Subtitle(subtitle string) THero {
	c.subtitle = subtitle
	return c
}

// Teaser sets the teaser image of the hero section.
func (c THero) Teaser(img ui.DecoredView) THero {
	c.teaserImg = img
	return c
}

// Render shows the hero section with title, subtitle, actions, and optional teaser image.
// On small screens, the text takes full width and the image is hidden.
func (c THero) Render(ctx core.RenderContext) core.RenderNode {
	winfo := ctx.Window().Info()
	small := winfo.SizeClass.Ordinal() <= core.SizeClassSmall.Ordinal()
	var heroTextWidth ui.Length

	if small {
		heroTextWidth = ui.Full
	} else {
		heroTextWidth = "70%"
	}

	if c.herotype == ImageFull {
		return imageFullView(c, small).Render(ctx)
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

func imageFullView(
	c THero,
	small bool,
) core.View {

	opacity := 0.5
	width := ui.Length("80%")
	if small {
		width = ui.Full
	}

	textAlignment := ui.TextAlignStart
	if c.alignment == ui.TopTrailing || c.alignment == ui.Trailing || c.alignment == ui.BottomTrailing {
		textAlignment = ui.TextAlignEnd
	} else if c.alignment == ui.Top || c.alignment == ui.Center || c.alignment == ui.Bottom {
		textAlignment = ui.TextAlignCenter
	}

	return ui.VStack(

		ui.IfFunc(c.teaserImg != nil, func() core.View {

			if small {
				return ui.VStack(
					c.teaserImg.Frame(ui.Frame{}.Size(ui.Full, ui.Full)),
				).
					Alignment(ui.Stretch).
					Opacity(opacity).
					Frame(ui.Frame{}.Size(ui.Full, ui.Full))
			}

			return ui.VStack(
				c.teaserImg.Frame(ui.Frame{}.FullWidth()),
			).
				Alignment(ui.Stretch).
				Opacity(opacity).
				Frame(ui.Frame{}.Size(ui.Full, ui.Full))
		}),

		ui.VStack(

			ui.VStack(
				ui.Text(c.title).
					Font(ui.Font{Weight: ui.DisplayAndLabelFontWeight, Size: ui.L60}).
					TextAlignment(textAlignment).
					Frame(ui.Frame{MaxWidth: "90%"}.FullWidth()),
				ui.Text(c.subtitle).
					TextAlignment(textAlignment).
					Frame(ui.Frame{MaxWidth: "90%"}.FullWidth()),
				ui.HStack(c.actions...).FullWidth().Alignment(c.alignment),
			).Alignment(ui.Leading).
				Gap(ui.L16).
				Frame(ui.Frame{}.FullWidth()),
		).
			Alignment(c.alignment).
			Position(ui.Position{
				Type:   ui.PositionAbsolute,
				ZIndex: 1,
			}).
			Padding(ui.Padding{}.All(ui.L32)).
			Frame(ui.Frame{MaxWidth: width}.FullWidth()),
	).
		Alignment(c.alignment).
		BackgroundColor(ui.ColorCardFooter).
		WithFrame(func(frame ui.Frame) ui.Frame {
			frame.Height = ui.L560 // limit the height of the hero element, since otherwise it would be depending on the image height
			frame.Width = c.frame.Width
			frame.MinHeight = c.frame.MinHeight
			frame.MaxHeight = c.frame.MaxHeight
			frame.MinWidth = c.frame.MinWidth
			frame.MaxWidth = c.frame.MaxWidth
			return frame
		}).
		Border(ui.Border{}.Radius(ui.L16))
}
