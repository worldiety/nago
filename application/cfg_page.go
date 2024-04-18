package application

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/gorilla"
	"go.wdy.de/nago/presentation/ora"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"time"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.fsys = append(c.fsys, fsys)
	return c
}

func (c *Configurator) Index(target string) *Configurator {
	c.uiApp.IndexTarget = target
	return c
}

func (c *Configurator) newHandler() http.Handler {

	factories := map[ora.ComponentFactoryId]core.ComponentFactory{}
	for id, f := range c.uiApp.Components {
		factories[id] = func(scope core.Window, requested ora.NewComponentRequested) core.Component {
			return f(scope)
		}
	}

	app2 := core.NewApplication(c.ctx, factories)
	r := chi.NewRouter()

	if c.debug {
		r.Use(
			cors.Handler(cors.Options{
				AllowedOrigins:   []string{"http://*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: true,
				MaxAge:           300, // Maximum value not ignored by any of major browsers

			}),
		)
		c.defaultLogger().Warn("using debug cors settings")
	}
	r.Use(
		c.loggerMiddleware,
		c.keycloakMiddleware,
	)

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			cookie, err := request.Cookie("wdy-ora-access")
			if err != nil {
				cookie = &http.Cookie{}
				cookie.Name = "wdy-ora-access"
				cookie.Value = string(ora.NewScopeID())
				cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
				cookie.Secure = false //TODO in release-mode this must be true
				cookie.HttpOnly = true
				cookie.Path = "/"
				http.SetCookie(writer, cookie)
			}

			dir := filepath.Dir(request.URL.Path)
			/*if strings.HasPrefix(base,"index"){
				request.URL.Path = "/"
			}*/
			if dir != "" && dir != "/assets" {
				request.URL.Path = "/"
				assets.ServeHTTP(writer, request)
				return
			}

			log.Println(request.URL.Path)

			assets.ServeHTTP(writer, request)
		}))

	}

	r.Mount("/wire", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.FromContext(r.Context())
		_ = logger
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true //TODO security implications?
			},
		} // use default options
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()

		queryParams := r.URL.Query()
		scopeID := queryParams.Get("_sid")

		// todo new
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
			}
		}()
		channel := gorilla.NewWebsocketChannel(conn)
		scope := app2.Connect(channel, ora.ScopeID(scopeID))
		defer scope.Destroy()

		cookie, _ := r.Cookie("wdy-ora-access")
		if err := channel.PublishLocal(ora.Marshal(ora.SessionAssigned{
			Type:      ora.SessionAssignedT,
			SessionID: cookie.Value,
		})); err != nil {
			slog.Error("cannot publish session assigned to local channel", slog.Any("err", err))
			return
		}

		if err := channel.Loop(); err != nil {
			fmt.Println(err)
			return
		}

	}))
	/*
		TODO
			r.Mount("/api/v1/upload", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				pageToken := r.Header.Get("x-page-token")
				page := appSrv.getPage(ui.PageInstanceToken(pageToken))
				if page == nil {
					logging.FromContext(r.Context()).Error("invalid page token for upload") //, slog.String("token", pageToken))
					w.WriteHeader(http.StatusNotFound)
					return
				}

				page.HandleHTTP(w, r)
			}))

			r.Mount("/api/v1/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				pageToken := r.Header.Get("x-page-token")
				if pageToken == "" {
					pageToken = r.URL.Query().Get("page")
				}
				page := appSrv.getPage(ui.PageInstanceToken(pageToken))
				if page == nil {
					logging.FromContext(r.Context()).Error("invalid page token for upload", slog.String("token", pageToken))
					w.WriteHeader(http.StatusNotFound)
					return
				}

				page.HandleHTTP(w, r)
			}))
	*/
	for _, route := range r.Routes() {
		fmt.Println(route.Pattern)
	}

	return r
}
