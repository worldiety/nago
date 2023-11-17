package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/logging"
	"log/slog"
	"net/http"
	"path/filepath"
)

type Scaffold struct {
	Navigation func(Context) slice.Slice[NavItem]
	Content    Persona
	Title      string
}

type responseScaffold struct {
	Type          string          `json:"type"`
	Content       responsePersona `json:"content"`
	Title         string          `json:"title"`
	Navigation    []NavItem       `json:"navigation"`
	Authenticated bool            `json:"authenticated"`
}

type responsePersona struct {
	Type       string   `json:"type"`
	TypeParams []string `json:"typeParams,omitempty"`
	Model      string   `json:"model"`
}

func (v Scaffold) Endpoints(page PageID, authenticated bool) []Endpoint {
	var res []Endpoint
	base := filepath.Join(apiUIPrefix, "page", string(page))
	name, tp := getGenericTypeName(v.Content)

	res = append(res, Endpoint{
		Method: http.MethodGet,
		Path:   base,
		Handler: func(writer http.ResponseWriter, request *http.Request) {

			rPage := responseScaffold{
				Type: "Scaffold",
				Content: responsePersona{
					Type:       name,
					TypeParams: tp,
					Model:      filepath.Join(base, string(v.Content.Id())),
				},
				Title:         v.Title,
				Authenticated: authenticated,
			}

			if v.Navigation != nil {
				rPage.Navigation = slice.UnsafeUnwrap(v.Navigation(Context{
					writer:  writer,
					request: request,
				}))
			}

			pageBuf, err := json.Marshal(rPage)
			if err != nil {
				panic(err)
			}

			if _, err := writer.Write(pageBuf); err != nil {
				logging.FromContext(request.Context()).Error("failed to write meta response", slog.Any("err", err))
			}
		},
	})

	return res
}
