package ui

import (
	"fmt"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/container/slice"
	"net/http"
	"path/filepath"
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

type router interface {
	MethodFunc(method, pattern string, h http.HandlerFunc)
}

type Pager interface {
	renderOpenAPI(r *openapi3.Reflector)
	PageID() PageID
	PageDescription() string
	Configure(r router)
	Authenticated() bool
	Pattern() string
}

type Page[Params any] struct {
	ID              PageID
	Title           string
	Description     string
	Children        slice.Slice[Component[Params]]
	Navigation      slice.Slice[PageNavTarget]
	Unauthenticated bool // secure by design, requires opt-out
}

func (p Page[Params]) Pattern() string {
	var zeroParams Params
	fields := pathNames(zeroParams)
	var pathParams string
	if len(fields) > 0 {
		pathParams = "{" + strings.Join(fields, "}/{") + "}"
	}

	return filepath.Join(apiPageSlug, string(p.ID), pathParams)
}

func (p Page[P]) Authenticated() bool {
	return !p.Unauthenticated
}

func (p Page[P]) PageID() PageID {
	return p.ID
}

func (p Page[Params]) PageDescription() string {
	return p.Description
}

func (p Page[P]) renderOpenAPI(r *openapi3.Reflector) {
	var zeroParams P

	pattern := p.Pattern()
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(zeroParams)

	oc.AddRespStructure(pageResponse{}, func(cu *openapi.ContentUnit) {
		cu.Description = fmt.Sprintf("This response contains a machine readable description of the *%s* (%s) page.", p.ID, p.Title)

	})
	oc.SetTags(string(p.ID))

	setSummaryAndDescription(oc, p.Description)
	must(r.AddOperation(oc))

	p.Children.Each(func(idx int, v Component[P]) {
		v.renderOpenAPI(zeroParams, string(p.ID), pattern, r)
	})
}

func (p Page[P]) Configure(r router) {
	var zeroParams P
	fields := pathNames(zeroParams)
	var pathParams string
	if len(fields) > 0 {
		pathParams = "{" + strings.Join(fields, "}/{") + "}"
	}

	pattern := filepath.Join(apiPageSlug, string(p.ID), pathParams)
	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		writeJson(writer, request, pageResponse{
			Type:  "Page",
			Title: p.Title,
			Children: slice.UnsafeUnwrap(slice.Map(p.Children, func(idx int, v Component[P]) Link {
				if err := v.ComponentID().Validate(); err != nil {
					panic(err)
				}

				actualPath := interpolatePathVariables[P](pattern, request)
				return Link(filepath.Join(actualPath, string(v.ComponentID())))
			})),
			Navigation: slice.UnsafeUnwrap(slice.Map(p.Navigation, func(idx int, v PageNavTarget) pageNavTarget {
				return pageNavTarget{
					Target: Link(filepath.Join(apiPageSlug, interpolatePathVariables[P](string(v.Target), request))),
					Anchor: filepath.Join("/", interpolatePathVariables[P](string(v.Target), request)),
					Icon:   v.Icon,
					Title:  v.Title,
				}
			})),
		})
	})

	p.Children.Each(func(idx int, v Component[P]) {
		v.configure(pattern, r)
	})
}

type PageNavTarget struct {
	Target Target
	Icon   Image
	Title  string
}

type Target string // either an absolute link or a pageid or something relativ

// actual page response
type pageResponse struct {
	Type       TypeDiscriminator `json:"type" pattern:"page" description:"This is always 'page'."`
	Title      string            `json:"title" description:"The title of the page."`
	Children   []Link            `json:"children" description:"A bunch of dynamic subcomponents links."`
	Navigation []pageNavTarget   `json:"navigation" description:"The primary navigation targets."`
}

type pageNavTarget struct {
	Target Link   `json:"link" description:"The page api target link."`
	Anchor string `json:"anchor" description:"The page anchor link."`
	Icon   Image  `json:"icon" description:"The icon to display."`
	Title  string `json:"title" description:"The caption of the page link."`
}
