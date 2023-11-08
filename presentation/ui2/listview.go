package ui2

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/logging"
	"log/slog"
	"net/http"
	"path/filepath"
)

const apiUIPrefix = "/api/v1/ui"

type ListItem[Identity any] struct {
	ID     Identity `json:"id"`
	Title  string
	Action Navigation
}

func (l ListItem[Identity]) MarshalJSON() ([]byte, error) {
	return marshalJSON(l)
}

type ListView[Identity any] struct {
	//Delete func(id ...Identity) Status
	List func() (slice.Slice[ListItem[Identity]], Status) `json:"-"`
}

func (lv ListView[Identity]) MarshalJSON() ([]byte, error) {
	return marshalJSON(lv)
}

func (ListView[Identity]) isPersona() {}

type responseListViewMeta struct {
	List          string `json:"list"`
	Authenticated bool   `json:"authenticated"`
}

func (lv ListView[Identity]) Endpoints(page PageID, authenticated bool) []Endpoint {
	var res []Endpoint
	var meta responseListViewMeta
	meta.Authenticated = authenticated
	base := filepath.Join(apiUIPrefix, string(page), "listview")

	if lv.List != nil {
		meta.List = filepath.Join(base, "list-all")
		ep := Endpoint{
			Method: http.MethodGet,
			Path:   meta.List,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				items, _ := lv.List() // TODO fix status
				s := slice.UnsafeUnwrap(items)
				resp := response[[]ListItem[Identity]]{
					Data: s,
				}
				enc := json.NewEncoder(writer)
				if err := enc.Encode(resp); err != nil {
					logging.FromContext(request.Context()).Error("failed to encode json response", slog.Any("err", err))
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
