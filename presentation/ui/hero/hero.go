package hero

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// THero is a composite component (Hero).
// This component represents a prominent section with an image,
// title, subtitle, and optional action buttons.
type THero struct {
	title           string      // main headline text
	subtitle        core.View   // supporting subtitle text
	actions         []core.View // list of action buttons or links
	frame           ui.Frame    // layout frame for the hero section
	backgroundImage core.URI
	backgroundColor ui.Color
	textColor       ui.Color
	sideView        core.View
	alignment       ui.Alignment
	foregroundColor ui.Color
	border          ui.Border
	padding         ui.Padding
}

// Hero creates a new THero with the given title and a default full-width height of 320.
func Hero(title string) THero {
	return THero{
		title:           title,
		frame:           ui.Frame{MinHeight: ui.L320}.FullWidth(),
		backgroundColor: ui.M0,
		textColor:       ui.M1,
		alignment:       ui.Center,
		border:          ui.Border{}.Radius(ui.L32),
		padding:         ui.Padding{}.All(ui.L48),
	}
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

func (c THero) Padding(padding ui.Padding) THero {
	c.padding = padding
	return c
}

// Subtitle sets the subtitle text of the hero section.
func (c THero) Subtitle(text string) THero {
	return c.SubtitleView(ui.Text(text))
}

func (c THero) SubtitleView(subtitle core.View) THero {
	c.subtitle = subtitle
	return c
}

func (c THero) Alignment(alignment ui.Alignment) THero {
	c.alignment = alignment
	return c
}

// BackgroundImage places a fit-cover image into the background.
func (c THero) BackgroundImage(img core.URI) THero {
	c.backgroundImage = img
	return c
}

func (c THero) ForegroundColor(col ui.Color) THero {
	c.foregroundColor = col
	return c
}

func (c THero) BackgroundColor(color ui.Color) THero {
	c.backgroundColor = color
	return c
}

func (c THero) TextColor(color ui.Color) THero {
	c.textColor = color
	return c
}

func (c THero) SideView(img core.View) THero {
	c.sideView = img
	return c
}

func (c THero) SideSVG(svg core.SVG) THero {
	return c.SideView(ui.Image().Embed(svg).Frame(ui.Frame{Width: ui.Full, Height: ui.Full}))
}

// Render shows the hero section with title, subtitle, actions, and optional teaser image.
// On small screens, the text takes full width and the image is hidden.
func (c THero) Render(ctx core.RenderContext) core.RenderNode {
	winfo := ctx.Window().Info()
	small := winfo.SizeClass.Ordinal() <= core.SizeClassSmall.Ordinal()
	var heroTextWidth ui.Length

	if small {
		heroTextWidth = ui.Full
		c.alignment = ui.Center
	} else {
		heroTextWidth = "70%"
	}

	return ui.HStack(
		ui.VStack(
			ui.Text(c.title).Font(ui.DisplayLarge),
			c.subtitle,
			ui.HStack(c.actions...).FullWidth().Alignment(c.alignment),
		).Alignment(c.alignment).
			Gap(ui.L16).
			Padding(c.padding).
			Frame(ui.Frame{Width: heroTextWidth}),
		ui.IfFunc(c.sideView != nil, func() core.View {
			if small {
				return nil
			}
			return ui.VStack(
				c.sideView,
			).Alignment(ui.Stretch).Frame(ui.Frame{Width: "30%", Height: "100%"})
		}),
	).
		TextColor(c.textColor).
		With(func(stack ui.THStack) ui.THStack {
			if c.backgroundImage != "" {
				bg := ui.Background{}.AppendURI(c.backgroundImage).Fit(ui.FitCover)
				if c.foregroundColor != "" {
					bg = bg.AppendLinearGradient(c.foregroundColor, c.foregroundColor)
				}

				stack = stack.Background(bg)
			}

			return stack
		}).
		Alignment(c.alignment).
		BackgroundColor(c.backgroundColor).
		Frame(c.frame).
		Border(c.border).
		Render(ctx)
}
