package hero

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// THero is a composite component (Hero).
// This component represents a prominent section with an image,
// title, subtitle, and optional action buttons.
type THero struct {
	title                string      // main headline text
	subtitle             string      // supporting subtitle text
	subtitleView         core.View   // supporting subtitle view, replaces subtitle text if set
	actions              []core.View // list of action buttons or links
	frame                ui.Frame    // layout frame for the hero section
	backgroundImage      core.URI
	backgroundColor      ui.Color
	textColor            ui.Color
	sideView             core.View
	alignment            ui.Alignment
	foregroundColorLight ui.Color
	foregroundColorDark  ui.Color
	border               ui.Border
	padding              ui.Padding
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
		padding: ui.Padding{
			Top:    ui.L60,
			Left:   ui.L80,
			Right:  ui.L80,
			Bottom: ui.L48,
		},
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
	c.subtitle = text
	return c
}

func (c THero) SubtitleView(view core.View) THero {
	c.subtitleView = view
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
	c.foregroundColorLight = col
	return c
}

func (c THero) ForegroundColorAdaptive(onLight, onDark ui.Color) THero {
	c.foregroundColorLight = onLight
	c.foregroundColorDark = onDark
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

// SideImage sets a side image for the hero section, which is displayed alongside the text content.
func (c THero) SideImage(img ui.TImage) THero {
	return c.SideView(img.ObjectFit(ui.FitContain).Padding(ui.Padding{}.All(ui.L64)).Frame(ui.Frame{}.FullHeight()))
}

// Render shows the hero section with title, subtitle, actions, and optional teaser image.
// On small screens, the text takes full width and the image is hidden.
func (c THero) Render(ctx core.RenderContext) core.RenderNode {
	winfo := ctx.Window().Info()
	small := winfo.SizeClass.Ordinal() <= core.SizeClassSmall.Ordinal()

	contentAlignment := ui.Leading
	textAlignment := ui.TextAlignStart
	if c.alignment == ui.Trailing || c.alignment == ui.TopTrailing || c.alignment == ui.BottomTrailing {
		contentAlignment = ui.Trailing
		textAlignment = ui.TextAlignEnd
	}

	fgColor := c.foregroundColorLight
	if winfo.PrefersDark() && c.foregroundColorDark != "" {
		fgColor = c.foregroundColorDark
	}

	colors := core.Colors[ui.Colors](ctx.Window())

	if fgColor == "" {
		lightM8, err := colors.M0.WithChromaAndTone(8, 10)
		if err == nil {
			fgColor = lightM8.WithTransparency(30)
		}
	}

	if c.frame.MinHeight == "" {
		c.frame.MinHeight = c.minHeight(winfo)
	}

	textColor := c.textColor
	darkM8, err := colors.M8.WithChromaAndTone(8, 98)
	if err == nil {
		textColor = darkM8
	}

	c.textColor = textColor

	return ui.HStack(
		ui.VStack(
			ui.Text(c.title).Font(c.titleFont(winfo)).TextAlignment(textAlignment),
			ui.IfElse(c.subtitleView != nil, c.subtitleView, ui.Text(c.subtitle).TextAlignment(textAlignment)),
			ui.HStack(c.actions...).FullWidth().Alignment(c.alignment),
		).Alignment(contentAlignment).
			Gap(ui.L16).
			Padding(c.padding).
			Frame(ui.Frame{MaxWidth: c.contentWidth(winfo)}),
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
				if fgColor != "" {
					bg = bg.AppendLinearGradient(fgColor, fgColor)
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

func (c THero) titleFont(winfo core.WindowInfo) ui.Font {
	if winfo.SizeClass < core.SizeClassMedium {
		return ui.DisplaySmall
	}

	if winfo.SizeClass < core.SizeClassXL {
		return ui.DisplayMedium
	}

	return ui.DisplayLarge
}

func (c THero) contentWidth(winfo core.WindowInfo) ui.Length {
	if winfo.SizeClass >= core.SizeClassMedium && c.sideView != nil {
		return "70%"
	}

	if winfo.SizeClass >= core.SizeClassMedium {
		return "80%"
	}

	if winfo.SizeClass >= core.SizeClassMedium {
		return "90%"
	}

	return ui.Full
}

func (c THero) minHeight(winfo core.WindowInfo) ui.Length {
	if winfo.SizeClass >= core.SizeClassXL {
		return ui.L480
	}

	if winfo.SizeClass >= core.SizeClassLarge {
		return ui.L400
	}

	if winfo.SizeClass >= core.SizeClassMedium {
		return ui.L320
	}

	if winfo.SizeClass >= core.SizeClassSmall {
		return ui.L256
	}

	if winfo.SizeClass >= core.SizeClassMedium {
		return ui.L200
	}

	return ui.L0
}
