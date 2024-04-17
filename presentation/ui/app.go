package ui

import (
	"fmt"
	"github.com/flowchartsman/swaggerui"
	"github.com/go-chi/chi/v5"
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"net/http"
	"regexp"
	"strings"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

const (
	apiSlug     = "/api/v1/"
	apiUiSlug   = apiSlug + "ui/"
	apiPageSlug = apiUiSlug + "page/"
	apiAppSlug  = apiUiSlug + "application"
)

type PageID string

func (p PageID) Validate() error {
	if len(validPageIdRegex.FindAllStringSubmatch(string(p), -1)) != 1 {
		return fmt.Errorf("the id '%s' is invalid and must match the [a-z0-9_\\-{/}]+", string(p))
	}

	return nil
}

// deprecated use core.Application
type Application struct {
	Name        string
	Version     string
	Description string
	//deprecated this is just the .
	IndexTarget string
	OIDC        []OIDCProvider //deprecated must be unified/abstracted away
	//deprecated we work with components now
	LivePages  map[PageID]func(Wire) *Page
	Components map[ora.ComponentFactoryId]func(realm core.Window) core.Component
}

func (a *Application) ConfigureRouter(router chi.Router) {
	router.Handle("/api/doc/*", http.StripPrefix("/api/doc", swaggerui.Handler(a.renderOpenAPI())))

	router.Get(apiAppSlug, func(writer http.ResponseWriter, request *http.Request) {
		res := appResponse{
			Name:        a.Name,
			Description: a.Description,
			Index:       string(a.IndexTarget),

			LivePages: func() []livePage {
				var tmp []livePage
				for id, _ := range a.LivePages {
					tmp = append(tmp, livePage{
						ID:            id,
						Authenticated: false,
						Anchor:        "/" + strings.TrimPrefix(string(id), apiPageSlug),
					})
				}

				return tmp
			}(),
			OIDC: a.OIDC,
		}

		writeJson(writer, request, res)
	})
}

func (a *Application) renderOpenAPI() []byte {
	o3a := &openapi3.Reflector{}
	o3a.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	o3a.Spec.Info.
		WithTitle(a.Name + " API").
		WithVersion(a.Version).
		WithDescription("Copyright by worldiety GmbH")

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
	Name        string         `json:"name" description:"The name of the entire application."`
	Description string         `json:"description" description:"The applications purpose description."`
	Index       string         `json:"index" description:"The default index page target to load."`
	OIDC        []OIDCProvider `json:"oidc"`
	LivePages   []livePage     `json:"livePages"`
}

type livePage struct {
	ID            PageID `json:"id"`
	Authenticated bool   `json:"authenticated"`
	Anchor        string `json:"anchor"`
}

// OIDCProvider does not perform any oauth workflow in the backend. Instead, it expects that workflow at the
// frontend-side and only gets the jwts
type OIDCProvider struct {
	Name                  string `json:"name"`
	Authority             string `json:"authority"`             // e.g. http://localhost:8080/realms/myapp for a local keycloak or https://accounts.google.com
	ClientID              string `json:"clientID"`              // used by the frontend
	ClientSecret          string `json:"clientSecret"`          // used by the frontend
	RedirectURL           string `json:"redirectURL"`           // used by the frontend
	PostLogoutRedirectUri string `json:"postLogoutRedirectUri"` // used by frontend
}
