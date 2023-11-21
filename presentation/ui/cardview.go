package ui

import (
	"fmt"
	"github.com/swaggest/openapi-go/openapi3"
	"go.wdy.de/nago/container/slice"
	"log/slog"
	"net/http"
	"path/filepath"
)

type CardView[Params any] struct {
	ID          ComponentID
	List        func(p Params) (slice.Slice[Card], error)
	Description string
}

func (lv CardView[Params]) ComponentID() ComponentID {
	return lv.ID
}

func (lv CardView[Params]) configure(parentSlug string, r router) {
	pattern := filepath.Join(parentSlug, string(lv.ID))

	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		metaCV := cardViewResponse{Type: "CardView"}

		if lv.List != nil {
			params := parseParams[Params](request)
			list, err := lv.List(params)
			if err != nil {
				slog.Default().Error("failed to list cards", slog.Any("err", err))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			metaCV.Cards = slice.UnsafeUnwrap(slice.Map(list, func(idx int, v Card) card {
				return card{
					Type:        "Card",
					Title:       v.Title,
					Subtitle:    v.Subtitle,
					PrependIcon: v.PrependIcon,
					AppendIcon:  v.AppendIcon,
					Content: func() any {
						switch t := v.Content.(type) {
						case CardText:
							return cardText{
								Type:  "CardText",
								Value: string(t),
							}
						default:
							panic(fmt.Errorf("implement me: %T", t))
						}
					}(),
					Actions: slice.UnsafeUnwrap(slice.Map(v.Actions, func(idx int, v Button) button {
						return button{
							Type:    "Button",
							Caption: v.Caption,
							Action:  v.Action,
						}
					})),
				}
			}))
		}

		writeJson(writer, request, metaCV)
	})
}

func (lv CardView[Params]) renderOpenAPI(p Params, tag string, parentSlug string, r *openapi3.Reflector) {
	pattern := filepath.Join(parentSlug, string(lv.ID))
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(p)
	oc.AddRespStructure(cardViewResponse{})
	oc.SetTags(tag)
	setSummaryAndDescription(oc, lv.Description)
	must(r.AddOperation(oc))
}

type Card struct {
	Title       string
	Subtitle    string
	PrependIcon Image
	AppendIcon  Image
	Content     CardContent
	Actions     slice.Slice[Button]
}

type CardContent interface {
	isCardContent()
}

type CardText string

func (CardText) isCardContent() {}

type cardText struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type cardViewResponse struct {
	Type  string `json:"type"`
	Cards []card `json:"cards"`
}

type card struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Content     any      `json:"content"`
	Actions     []button `json:"actions"`
	PrependIcon Image    `json:"prependIcon"`
	AppendIcon  Image    `json:"appendIcon"`
}

type button struct {
	Type    string `json:"type"`
	Caption string `json:"caption"`
	Action  Action `json:"action"`
}
