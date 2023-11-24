package ui

import (
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"path/filepath"
	"strconv"
)

type SVG[Params any] struct {
	ID          ComponentID
	Render      func(p Params) (string, error)
	Description string
	MaxWidth    int
}

func (s SVG[Params]) ComponentID() ComponentID {
	return s.ID
}

func (s SVG[Params]) configure(app *Application, authRequired bool, parentSlug string, r router) {
	pattern := filepath.Join(parentSlug, string(s.ID))
	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		params := parseParams[Params](request, authRequired)
		buf, err := s.Render(params)
		if err != nil {
			panic(err) // TODO Torben
		}
		tmp := svgResponse{
			Type: "SVG",
			SVG:  buf,
		}
		if s.MaxWidth > 0 {
			tmp.MaxWidth = strconv.Itoa(s.MaxWidth)
		}

		writeJson(writer, request, tmp)
	})
}

func (s SVG[Params]) renderOpenAPI(p Params, tag string, parentSlug string, r *openapi3.Reflector) {
	pattern := filepath.Join(parentSlug, string(s.ID))
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(p)
	oc.AddRespStructure(svgResponse{})
	oc.SetTags(tag)
	setSummaryAndDescription(oc, s.Description)
	must(r.AddOperation(oc))
}

type svgResponse struct {
	Type     string `json:"type"`
	SVG      string `json:"svg"`
	MaxWidth string `json:"maxWidth"`
}
