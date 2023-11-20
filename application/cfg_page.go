package application

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/presentation/ui"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.fsys = append(c.fsys, fsys)
	return c
}

func (c *Configurator) Page(p ui.Pager) *Configurator {
	if err := p.PageID().Validate(); err != nil {
		panic(err)
	}
	c.uiApp.Pages = c.uiApp.Pages.AppendAll(p)
	return c
}

func (c *Configurator) newHandler() http.Handler {

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

	c.uiApp.ConfigureRouter(r)

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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

	for _, route := range r.Routes() {
		fmt.Println(route.Pattern)
	}

	return r
}
