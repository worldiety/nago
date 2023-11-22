package ui

import (
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/container/slice"
	"net/http"
	"path/filepath"
	"strconv"
)

type Timeline[Params any] struct {
	ID          ComponentID
	Items       slice.Slice[TimelineItem]
	Description string
}

func (t Timeline[Params]) ComponentID() ComponentID {
	return t.ID
}

func (t Timeline[Params]) configure(parentSlug string, r router) {
	pattern := filepath.Join(parentSlug, string(t.ID))
	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		tResp := timelineResponse{
			Type: "Timeline",
			Items: slice.UnsafeUnwrap(slice.Map(t.Items, func(idx int, v TimelineItem) timelineItem {
				color := v.Color
				if color == "" {
					color = "grey"
				}
				alternateText := ""
				if v.Icon == nil {
					alternateText = strconv.Itoa(idx + 1)
				}
				return timelineItem{
					Type:             "TimelineItem",
					Icon:             v.Icon,
					AlternateDotText: alternateText,
					Color:            color,
					Title:            v.Title,
					Target:           string(v.Target),
				}
			})),
		}
		writeJson(writer, request, tResp)
	})
}

func (t Timeline[Params]) renderOpenAPI(p Params, tag string, parentSlug string, r *openapi3.Reflector) {
	pattern := filepath.Join(parentSlug, string(t.ID))
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(p)
	oc.AddRespStructure(timelineResponse{})
	oc.SetTags(tag)
	setSummaryAndDescription(oc, t.Description)
	must(r.AddOperation(oc))
}

type TimelineItem struct {
	Icon   Image
	Color  string // use a different dot color, e.g. primary or error
	Title  string
	Target Target // alternative clickable for navigation
}

type timelineItem struct {
	Type             string `json:"type"`
	Icon             Image  `json:"icon"`
	Color            string `json:"color"`
	Title            string `json:"title"`
	AlternateDotText string `json:"alternateDotText"`
	Target           string `json:"target"`
}

type timelineResponse struct {
	Type  string         `json:"type"`
	Items []timelineItem `json:"items"`
}
