package application

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/container/serrors"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui2"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

type page struct {
	ID            string `json:"id"`
	Endpoint      string `json:"endpoint"`
	Anchor        string `json:"anchor"`
	Authenticated bool   `json:"authenticated"`
}

type routeIndex struct {
	Pages []page `json:"pages"`
}

func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.fsys = append(c.fsys, fsys)
	return c
}

func (c *Configurator) Page(h ui.PageHandler) *Configurator {
	panic("delete me")
	return c
}

func (c *Configurator) Page2(id ui2.PageID, authenticated bool, s ui2.Scaffold) *Configurator {
	if len(validPageIdRegex.FindAllStringSubmatch(string(id), -1)) != 1 {
		panic(fmt.Errorf("the id '%s' is invalid and must match the [a-z0-9_\\-{/}]+", string(id)))
	}

	if _, ok := c.pages[id]; ok {
		panic(fmt.Errorf("another page with the same id has already been declared: %v ", id))
	}

	c.pages[id] = s

	eps := s.Content.Endpoints(id, authenticated)
	c.endpoints = append(c.endpoints, s.Endpoints(id, authenticated)...)
	c.endpoints = append(c.endpoints, eps...)

	return c
}

type applicationResponse struct {
	Name  string            `json:"name"`
	Pages map[string]string `json:"pages"`
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

	for _, endpoint := range c.endpoints {
		r.Method(endpoint.Method, endpoint.Path, endpoint.Handler)
		c.defaultLogger().Info("registered", slog.String("route", endpoint.Path))
	}

	idxApp := "/api/v1/ui"

	r.Get(idxApp, func(writer http.ResponseWriter, request *http.Request) {
		app := applicationResponse{
			Name:  c.appName,
			Pages: make(map[string]string),
		}
		for id := range c.pages {
			app.Pages[string(id)] = filepath.Join("/api/v1/ui/", string(id))
		}

		buf, err := json.Marshal(app)
		serrors.OrPanic(err)
		writer.Write(buf)
	})

	c.defaultLogger().Info("registered", slog.String("route", idxApp))

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", assets)
	}

	return r
}
