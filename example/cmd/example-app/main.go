package main

import (
	"encoding/json"
	"io"
	"net/http"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/example/dashboard/web"
	"go.wdy.de/nago/presentation/rest"
	"go.wdy.de/nago/presentation/ui"
)

func main() {

	type bla struct {
		Blub int `json:"blub,string"`
	}

	var x bla
	if err := json.Unmarshal([]byte(`{"Blub":"2"}`), &x); err != nil {
		panic(err)
	}

	app := application.NewApplication(
		application.PresentationLayer{
			Pages: slice.Of(ui.Route{
				Pattern: "/",
				Render:  nil,
			}),
			Events: slice.Slice[ui.Event]{},
			Http: slice.Of(rest.Route{
				Pattern: "/api/v1/",
				Method:  http.MethodPost,
				Handler: rest.HandleFileUpload(1024*1024, func(name string, size int64, r io.ReaderAt) error {
					return nil
				}),
			}),
		},
	)

	ui.Handler(
		web.Render,
		ui.OnEvent(func(model web.DashboardModel, evt any) web.DashboardModel {
			model.Count++
			return model
		}),
	)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
