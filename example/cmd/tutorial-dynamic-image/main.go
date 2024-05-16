package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"
)

//go:embed animated-green-astronaut-helmet.svg
var mySvg string

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		var invocationCount int64

		cfg.Component(".", func(wnd core.Window) core.Component {
			return ui.NewImage(func(img *ui.Image) {
				// tell the runtime, that we want a dynamic resource and need an uri for that
				uri, err := wnd.AsURI(func() (io.Reader, error) {
					// on every reload, we modify our SVG image
					count := atomic.AddInt64(&invocationCount, 1)
					newSvg := strings.ReplaceAll(mySvg, "$TEXT", fmt.Sprintf("called %dx", count))
					return core.WithMimeType("image/svg+xml", bytes.NewBufferString(newSvg)), nil
				})

				if err != nil {
					slog.Error("cannot create uri", slog.Any("err", err))
					return
				}

				// set the initial uri into the image component
				img.URI().Set(uri)

				// now schedule a 5 sec invalidation event
				core.Schedule(wnd, 5*time.Second, func() {
					// use a little trick to trigger a reload in the native image view by changing the uri a bit
					img.URI().Set(ora.URI(fmt.Sprintf("%s&version=%d", uri, atomic.LoadInt64(&invocationCount))))
				})

			})
		})
	}).Run()
}
