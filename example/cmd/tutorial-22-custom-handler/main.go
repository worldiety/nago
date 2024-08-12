package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
)

//go:embed animated-green-astronaut-helmet.svg
var mySvg string

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		var invocationCount int64

		// define a raw custom api handler
		uri := "/api/v1/astronaut.svg"
		cfg.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
			// on every reload, we modify our SVG image
			// CAUTION: be thread safe here
			count := atomic.AddInt64(&invocationCount, 1)
			newSvg := strings.ReplaceAll(mySvg, "$TEXT", fmt.Sprintf("called %dx", count))
			w.Header().Set("Content-Type", "image/svg+xml")
			if _, err := w.Write([]byte(newSvg)); err != nil {
				slog.Error("failed to write svg", "err", err)
			}
		})

		// press reload in your browser and watch how the image is loaded
		cfg.Component(".", func(wnd core.Window) core.View {
			// use a little trick to trigger a reload in the native image view by changing the uri a bit
			return ui.Image().URI(core.URI(fmt.Sprintf("%s?version=%d", uri, atomic.LoadInt64(&invocationCount))))
		})
	}).Run()
}
