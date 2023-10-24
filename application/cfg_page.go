package application

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.wdy.de/nago/internal/text"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"net/http"
)

type page struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
	Anchor   string `json:"anchor"`
}

type routeIndex struct {
	Pages []page `json:"pages"`
}

func (c *Configurator) Page(h ui.PageHandler) *Configurator {
	route := "/api/v1/ui/page/" + text.SafeName(h.ID())
	if _, ok := c.pages[route]; ok {
		panic(fmt.Errorf("another page with the same id->path conversion rule has already been declared: %s -> %s", h.ID(), route))
	}

	c.pages[route] = h
	return c
}

func (c *Configurator) newHandler() http.Handler {
	r := chi.NewRouter()
	if c.debug {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers

		}))
		c.defaultLogger().Warn("using debug cors settings")
	}
	var idx routeIndex
	for route, handler := range c.pages {
		c.defaultLogger().Info("registered", slog.String("route", route))
		r.Get(route, handler.ServeHTTP)
		idx.Pages = append(idx.Pages, page{
			ID:       handler.ID(),
			Endpoint: route,
			Anchor:   "/" + text.SafeName(handler.ID()),
		})
	}

	idxRoute := "/api/v1/ui/pages"
	r.Get(idxRoute, func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.Marshal(idx)
		if err != nil {
			panic(fmt.Errorf("internal error: %w", err))
		}

		if _, err := w.Write(buf); err != nil {
			c.defaultLogger().Error("cannot write", err)
		}
	})

	c.defaultLogger().Info("registered", slog.String("route", idxRoute))

	return r
}
