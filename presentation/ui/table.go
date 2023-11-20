package ui

import (
	"cmp"
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/container/slice"
	dm "go.wdy.de/nago/domain"
	"net/http"
	"path/filepath"
)

type Table[E dm.Entity[ID], ID cmp.Ordered, Params any] struct {
	ID          ComponentID
	Description string
	Headers     slice.Slice[TableHeader]
	Delete      func(p Params, ids slice.Slice[ID]) error
	List        func(p Params) (slice.Slice[TableRow[ID]], error)
}

func (t Table[E, ID, Params]) ComponentID() ComponentID {
	return t.ID
}

func (t Table[E, ID, Params]) configure(parentSlug string, r router) {
	pattern := filepath.Join(parentSlug, string(t.ID))
	metaT := tableResponse{Type: "Table"}
	if t.List != nil {
		metaT.Links.List = Link(filepath.Join(pattern, "list"))
		r.MethodFunc(http.MethodGet, string(metaT.Links.List), func(writer http.ResponseWriter, request *http.Request) {
			params := parseParams[Params](request)
			items, _ := t.List(params)
			resp := tableDataListResponse{}
			if t.Headers.Len() == 0 && items.Len() > 0 {
				var tmp []tableHead
				items.At(0).Cells.Each(func(idx int, v TableCell) {
					tmp = append(tmp, tableHead{
						Title: v.Key,
						Key:   v.Key,
						Align: string(AlignStart),
					})
				})
				resp.Headers = tmp
			} else {
				t.Headers.Each(func(idx int, v TableHeader) {
					resp.Headers = append(resp.Headers, tableHead{
						Title: v.Title,
						Key:   v.Key,
						Align: string(v.Align),
					})
				})
			}

			items.Each(func(idx int, v TableRow[ID]) {
				obj := map[string]any{}
				obj["_id"] = v.ID
				v.Cells.Each(func(idx int, v TableCell) {
					obj[v.Key] = v.Value
				})
				resp.Rows = append(resp.Rows, obj)
			})

			writeJson(writer, request, resp)
		})
	}

	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		tmp := metaT
		tmp.Links.List = Link(interpolatePathVariables[Params](string(tmp.Links.List), request))
		tmp.Links.Delete = Link(interpolatePathVariables[Params](string(tmp.Links.Delete), request))
		writeJson(writer, request, tmp)
	})
}

func (t Table[E, ID, Params]) renderOpenAPI(p Params, tag string, parentSlug string, r *openapi3.Reflector) {
	pattern := filepath.Join(parentSlug, string(t.ID))
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(p)
	oc.AddRespStructure(tableResponse{})
	oc.SetTags(tag)
	setSummaryAndDescription(oc, t.Description)
	must(r.AddOperation(oc))

	if t.List != nil {
		oc := must2(r.NewOperationContext(http.MethodGet, filepath.Join(pattern, "list")))
		oc.AddReqStructure(p)
		oc.AddRespStructure(tableResponse{})
		oc.SetTags(tag)
		must(r.AddOperation(oc))
	}
}

type TableRow[ID any] struct {
	ID    ID
	Cells slice.Slice[TableCell]
}

type TableCell struct {
	Key   string
	Value any // valid are string, int and float. Other types may break or create rubbish view.
}

type TableCellAlignment string

const (
	AlignStart TableCellAlignment = "start"
	AlignEnd   TableCellAlignment = "end"
)

type TableHeader struct {
	Title string
	Key   string
	Align TableCellAlignment
}

type tableResponse struct {
	Type  string `json:"type"`
	Links struct {
		List   Link `json:"list,omitempty"`
		Delete Link `json:"delete,omitempty"`
	} `json:"links"`
}

type tableDataListResponse struct {
	Headers []tableHead `json:"headers"`
	Rows    []tableRow  `json:"rows"`
}

type tableRow map[string]any

type tableHead struct {
	Title string `json:"title"`
	Align string `json:"align"` // start or end
	Key   string `json:"key"`
}
