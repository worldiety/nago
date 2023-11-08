package ui2

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/logging"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

type Scaffold struct {
	ApplicationName string
	Navigation      func(Context) slice.Slice[NavItem]
	Content         Persona
}

type responsePage struct {
	Type    string          `json:"type"`
	Content responsePersona `json:"content"`
	Model   string          `json:"model"`
}

type responsePersona struct {
	Type       string   `json:"type"`
	TypeParams []string `json:"typeParams,omitempty"`
	Model      string   `json:"model"`
}

func (v Scaffold) endpointsPage(page PageID) []Endpoint {
	var res []Endpoint
	base := filepath.Join(apiUIPrefix, string(page))
	name, tp := getGenericTypeName(v.Content)
	rPage := responsePage{
		Type: "Scaffold",
		Content: responsePersona{
			Type:       name,
			TypeParams: tp,
			Model:      filepath.Join(base, strings.ToLower(name)),
		},
		Model: filepath.Join(base, "scaffold"),
	}

	pageBuf, err := json.Marshal(rPage)
	if err != nil {
		panic(err)
	}
	res = append(res, Endpoint{
		Method: http.MethodGet,
		Path:   base,
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			if _, err := writer.Write(pageBuf); err != nil {
				logging.FromContext(request.Context()).Error("failed to write meta response", slog.Any("err", err))
			}
		},
	})

	return res
}

type responseScaffoldMeta struct {
	Navigation    string `json:"navigation"`
	Authenticated bool   `json:"authenticated"`
}

func (v Scaffold) Endpoints(page PageID, authenticated bool) []Endpoint {
	var res []Endpoint
	var meta responseScaffoldMeta
	meta.Authenticated = authenticated
	base := filepath.Join(apiUIPrefix, string(page), "scaffold")

	res = append(res, v.endpointsPage(page)...)

	if v.Navigation != nil {
		meta.Navigation = filepath.Join(base, "navigation")
		ep := Endpoint{
			Method: http.MethodGet,
			Path:   meta.Navigation,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ctx := newContext(writer, request)
				items := v.Navigation(ctx) // TODO do we have an error here?
				s := slice.UnsafeUnwrap(items)
				resp := response[[]NavItem]{
					Data: s,
				}
				enc := json.NewEncoder(writer)
				if err := enc.Encode(resp); err != nil {
					logging.FromContext(request.Context()).Error("failed to encode 'scaffold' json response", slog.Any("err", err))
				}
			},
		}

		res = append(res, ep)
	}

	metaBuf, err := json.Marshal(meta)
	if err != nil {
		panic("unreachable")
	}

	res = append(res, Endpoint{
		Method: http.MethodGet,
		Path:   base,
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			if _, err := writer.Write(metaBuf); err != nil {
				logging.FromContext(request.Context()).Error("failed to write meta response", slog.Any("err", err))
			}
		},
	})

	return res
}
