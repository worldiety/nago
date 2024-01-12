package pages

import (
	"bytes"
	_ "embed"
	"go.wdy.de/nago/presentation/ui"
	"io"
)

//go:embed assets/morty_vanilla.png
var mortyImg []byte

func Morty(wire ui.Wire) *ui.Page {
	return ui.NewPage(wire, func(page *ui.Page) {
		page.Body().Set(
			ui.NewScaffold(func(scaffold *ui.Scaffold) {
				scaffold.TopBar().Right.Set(ui.NewButton(func(btn *ui.Button) {
					btn.Caption().Set("zurück")
					btn.Action().Set(func() {
						page.History().Open("example", ui.Values{
							"a": "b",
						})
					})
					scaffold.Body().Set(
						ui.NewVBox(func(vbox *ui.VBox) {
							vbox.Append(
								ui.NewImage(func(img *ui.Image) {
									img.Source(func() (io.Reader, error) {
										return bytes.NewBuffer(mortyImg), nil
									})
								}),
							)
						}),
					)
				}))

			}),
		)
	})
}
