package avatar

import (
	"go.wdy.de/nago/image"
	http_image "go.wdy.de/nago/image/http"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strings"
	"unicode"
)

type TAvatar struct {
	paraphe  string
	url      core.URI
	data     []byte
	action   func()
	frame    ui.Frame
	textSize ui.Length
	border   ui.Border
	color    ui.Color
}

func TextOrImage(text string, img image.ID) TAvatar {
	if img != "" {
		return URI(http_image.URI(img, image.FitCover, 64, 64))
	}

	return Text(text)
}

func Text(paraphe string) TAvatar {
	if paraphe == "" {
		paraphe = "?"
	} else {

		var sb strings.Builder
		if len(paraphe) > 2 {
			tokens := strings.Split(paraphe, " ")
			for i, token := range tokens {
				if i > 1 {
					break
				}

				for _, r := range token {
					sb.WriteRune(unicode.ToUpper(r))
					break
				}
			}
		}

		if sb.Len() == 0 {
			paraphe = "?"
		} else {
			paraphe = sb.String()
		}
	}

	return TAvatar{
		paraphe: paraphe,
		color:   ui.A0,
		border:  ui.Border{}.Circle(),
	}.Size(ui.L40)
}

func URI(uri core.URI) TAvatar {
	return TAvatar{
		url:    uri,
		color:  ui.A0,
		border: ui.Border{}.Circle(),
	}.Size(ui.L40)
}

func Embed(data []byte) TAvatar {
	return TAvatar{
		data:   data,
		color:  ui.A0,
		border: ui.Border{}.Circle(),
	}.Size(ui.L40)
}

func (c TAvatar) Border(border ui.Border) TAvatar {
	c.border = border
	return c
}

func (c TAvatar) Action(fn func()) TAvatar {
	c.action = fn
	return c
}

func (c TAvatar) Size(widthAndHeight ui.Length) TAvatar {
	c.frame = ui.Frame{}.Size(widthAndHeight, widthAndHeight)
	c.textSize = widthAndHeight.Mul(0.4)
	return c
}

func (c TAvatar) Render(ctx core.RenderContext) core.RenderNode {
	if c.paraphe != "" {
		return ui.VStack(ui.Text(c.paraphe).Font(ui.Font{Size: c.textSize}).Color("#000000")).
			Action(c.action).
			BackgroundColor(c.color).
			Frame(c.frame).
			Border(c.border).
			Render(ctx)
	}

	// must be an image
	img := ui.Image()
	if c.data != nil {
		img = img.Embed(c.data)
	}

	if c.url != "" {
		img = img.URI(c.url)
	}

	return ui.VStack(img).
		Action(c.action).
		Frame(c.frame).
		Border(c.border).
		Render(ctx)
}
