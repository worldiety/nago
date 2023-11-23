package ui

import (
	"fmt"
	"github.com/flowchartsman/swaggerui"
	"github.com/go-chi/chi/v5"
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/container/slice"
	"net/http"
	"strings"
)

type Application struct {
	Name        string
	Version     string
	Description string
	Pages       slice.Slice[Pager]
	IndexTarget Target
}

func (a *Application) ConfigureRouter(router chi.Router) {
	router.Handle("/api/doc/*", http.StripPrefix("/api/doc", swaggerui.Handler(a.renderOpenAPI())))
	a.Pages.Each(func(idx int, v Pager) {
		v.Configure(a, router)
	})

	router.Get(apiAppSlug, func(writer http.ResponseWriter, request *http.Request) {
		writeJson(writer, request, appResponse{
			Name:        a.Name,
			Description: a.Description,
			Index:       string(a.IndexTarget),
			Pages: slice.UnsafeUnwrap(slice.Map(a.Pages, func(idx int, v Pager) appPage {
				return appPage{
					ID:            v.PageID(),
					Authenticated: v.Authenticated(),
					Link:          Link(v.Pattern()),
					Anchor:        "/" + strings.TrimPrefix(v.Pattern(), apiPageSlug),
				}
			})),
		})
	})
}

func (a *Application) renderOpenAPI() []byte {
	o3a := &openapi3.Reflector{}
	o3a.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	o3a.Spec.Info.
		WithTitle(a.Name + " API").
		WithVersion(a.Version).
		WithDescription("Copyright by worldiety GmbH")

	a.Pages.Each(func(idx int, v Pager) {
		v.renderOpenAPI(o3a)
	})

	oc := must2(o3a.NewOperationContext(http.MethodGet, apiAppSlug))
	oc.AddRespStructure(appResponse{})
	must(o3a.AddOperation(oc))

	schema, err := o3a.Spec.MarshalYAML()
	if err != nil {
		panic(fmt.Errorf("illegal state: %w", err))
	}

	return schema
}

type appResponse struct {
	Name        string    `json:"name" description:"The name of the entire application."`
	Pages       []appPage `json:"pages" description:"All known and configured pages. Not all pages are directly addressable and therefore require parameters."`
	Description string    `json:"description" description:"The applications purpose description."`
	Index       string    `json:"index" description:"The default index page target to load."`
}

type appPage struct {
	ID            PageID `json:"id" description:"unique page identifier"`
	Authenticated bool   `json:"authenticated" description:"If true, the client must provide an authenticated user, otherwise any resource requests will fail."`
	Link          Link   `json:"link" description:"TODO: how to handle pages with parameters? just query?"`
	Anchor        string `json:"anchor"`
}

type links map[string]Link
