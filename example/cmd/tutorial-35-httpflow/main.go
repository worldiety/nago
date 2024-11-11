package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		// simulate some http redirect endpoint
		cfg.HandleFunc("/api/v1/oauth2/login", func(w http.ResponseWriter, r *http.Request) {
			slog.Info("got login request")
			redirectUri := r.URL.Query().Get("redirect_uri")
			if redirectUri != "/redirect" {
				slog.Info("wrong redirect uri")
				_, _ = w.Write([]byte("permission denied"))
				return
			}

			u, err := url.Parse(redirectUri)
			if err != nil {
				slog.Info("bad request", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			q := u.Query()
			q.Set("response_type", "code id_token")
			q.Set("scope", "openid 1234")
			q.Set("state", r.URL.Query().Get("state"))
			u.RawQuery = q.Encode()

			slog.Info("redirected to", "uri", u.String())

			http.Redirect(w, r, u.String(), http.StatusFound)
		})

		// index page to start the flow
		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.PrimaryButton(func() {
					core.HTTPFlow(wnd.Navigation(), "/api/v1/oauth2/login?redirect_uri=/redirect&state=xyz", "/redirect", "redirect")
				}).Title("start http-flow")).
				Frame(ui.Frame{}.MatchScreen())
		})

		// receiver page, which will be redirected to, either by browser directly or from app indirectly after intercepting webview callback
		cfg.RootView("redirect", func(wnd core.Window) core.View {
			return ui.VStack(
				slices.Collect(func(yield func(view core.View) bool) {
					yield(ui.Text("redirect target"))
					for k, v := range wnd.Values() {
						yield(ui.Text(fmt.Sprintf("%s = %s", k, v)))
					}
				})...,
			).
				Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
