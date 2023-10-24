package application

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.wdy.de/nago/internal/text"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"net/http"
)

type page struct {
	ID    string `json:"id"`
	Route string `json:"route"`
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
	var idx routeIndex
	for route, handler := range c.pages {
		c.defaultLogger().Info("registered", slog.String("route", route))
		r.Get(route, handler.ServeHTTP)
		idx.Pages = append(idx.Pages, page{
			ID:    handler.ID(),
			Route: route,
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
