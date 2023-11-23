package ui

import (
	"cmp"
	"encoding/json"
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/container/slice"
	dm "go.wdy.de/nago/domain"
	"log/slog"
	"net/http"
	"path/filepath"
)

type ListItem[Identity any] struct {
	ID     Identity `json:"id"`
	Title  string
	Action Action
}

func (l ListItem[Identity]) MarshalJSON() ([]byte, error) {
	return marshalJSON(l)
}

type ListView[E dm.Entity[ID], ID cmp.Ordered, Params any] struct {
	ID          ComponentID
	Delete      func(p Params, ids slice.Slice[ID]) error
	List        func(p Params) (slice.Slice[ListItem[ID]], error)
	Description string
}

func (lv ListView[E, ID, Params]) ComponentID() ComponentID {
	return lv.ID
}

func (lv ListView[E, ID, Params]) configure(app *Application, authRequired bool, parentSlug string, r router) {
	pattern := filepath.Join(parentSlug, string(lv.ID))
	metaLV := listViewResponse{Type: "ListView"}
	if lv.List != nil {
		metaLV.Links.List = Link(filepath.Join(pattern, "list"))
		r.MethodFunc(http.MethodGet, string(metaLV.Links.List), func(writer http.ResponseWriter, request *http.Request) {
			params := parseParams[Params](request, authRequired)
			items, _ := lv.List(params)
			s := slice.UnsafeUnwrap(items)
			resp := response[[]ListItem[ID]]{
				Data: s,
			}
			writeJson(writer, request, resp)
		})
	}

	if lv.Delete != nil {
		metaLV.Links.Delete = Link(filepath.Join(pattern, "delete-by-ids"))
		r.MethodFunc(http.MethodPost, string(metaLV.Links.Delete), func(writer http.ResponseWriter, request *http.Request) {
			params := parseParams[Params](request, authRequired)

			var idents deleteRequest[ID]
			dec := json.NewDecoder(request.Body)
			if err := dec.Decode(&idents); err != nil {
				slog.Default().Error("failed to decode json", slog.Any("err", err))
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := lv.Delete(params, slice.Of(idents.Identifiers...)); err != nil {
				slog.Default().Error("failed to delete entities", slog.Any("err", err))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			writeJson(writer, request, struct {
			}{})
		})
	}

	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		tmp := metaLV
		tmp.Links.List = Link(interpolatePathVariables[Params](string(tmp.Links.List), request))
		tmp.Links.Delete = Link(interpolatePathVariables[Params](string(tmp.Links.Delete), request))
		writeJson(writer, request, tmp)
	})
}

func (lv ListView[E, ID, Params]) renderOpenAPI(p Params, tag string, parentSlug string, r *openapi3.Reflector) {
	pattern := filepath.Join(parentSlug, string(lv.ID))
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(p)
	oc.AddRespStructure(listViewResponse{})
	oc.SetTags(tag)
	setSummaryAndDescription(oc, lv.Description)
	must(r.AddOperation(oc))

	if lv.List != nil {
		oc := must2(r.NewOperationContext(http.MethodGet, filepath.Join(pattern, "list")))
		oc.AddReqStructure(p)
		oc.AddRespStructure(response[[]ListItem[ID]]{})
		oc.SetTags(tag)
		must(r.AddOperation(oc))
	}

	if lv.Delete != nil {

		oc := must2(r.NewOperationContext(http.MethodPost, filepath.Join(pattern, "delete-by-ids")))
		oc.AddReqStructure(p)
		oc.AddReqStructure(deleteRequest[ID]{})
		oc.AddRespStructure(struct{}{})
		oc.SetTags(tag)
		must(r.AddOperation(oc))
	}
}

type listViewResponse struct {
	Type  string `json:"type"`
	Links struct {
		List   Link `json:"list,omitempty"`
		Delete Link `json:"delete,omitempty"`
	} `json:"links"`
}

type deleteRequest[Identity any] struct {
	Identifiers []Identity `json:"identifiers"`
}
