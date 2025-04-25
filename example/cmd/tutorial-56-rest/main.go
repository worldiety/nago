// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/hapi"
	cfghapi "go.wdy.de/nago/application/hapi/cfg"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/stoplight"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"mime/multipart"
	"net/url"
	"slices"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_56")
		cfg.SetSemanticVersion("0.1.2")
		cfg.SetName("Tutorial 56")

		cfg.Serve(vuejs.Dist())
		// we have multiple frontend openapi distributions provided, e.g. swagger, redocly or stoplight.
		// Note, that ALL frontends are broken in one or another way. E.g. swagger does not support even simplest
		// recursions and stoplight does not support multipart files.
		cfg.Serve(stoplight.Dist())
		//cfg.Serve(swagger.Dist())
		//cfg.Serve(redocly.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		option.MustZero(cfg.StandardSystems())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		api := std.Must(cfghapi.Enable(cfg)).API
		tokens := std.Must(cfg.TokenManagement())
		configureMyAPI(api, tokens)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return VStack(Text("hello world")).
				Frame(Frame{}.MatchScreen())

		})
	}).
		Run()
}

func configureMyAPI(api *hapi.API, tokens application.TokenManagement) {
	type StackElement struct {
		Line int
		File string
	}

	type Exception struct {
		Name  string         `json:"name"`
		Stack []StackElement `json:"stack"`
		Cause *Exception     `json:"cause"`
	}

	type UploadMetadata struct {
		DeviceName string
		AppVersion string
		KeyValues  map[string]string
		Exception  *Exception
	}

	type UploadRequest struct {
		TestHeader string
		Metadata   UploadMetadata
		Files      []*multipart.FileHeader
		Subject    auth.Subject
	}

	type SomeID string
	type NestedResponse struct {
		ID       SomeID  `json:"id,omitempty" example:"1234"`
		Url      url.URL `json:"url,omitempty"`
		OtherNum int32   `json:"other_num,omitempty"`
	}

	type UploadResponse struct {
		ID     string         `json:"id"`
		Yes    bool           `json:"yes" doc:"say no" required:"true"`
		Num    int            `json:"num" supportingText:"Irgendeine Nummer"`
		When   time.Time      `json:"when"`
		Nested NestedResponse `json:"nested"`
	}

	hapi.Post[UploadRequest](api, hapi.Operation{Path: "/api/v1/events", Summary: "Create a new event", Description: "A post will take the given meta data and files and persists it as an event. A unique tracking code is returned."}).
		Request(
			hapi.BearerAuth[UploadRequest](tokens.UseCases.AuthenticateSubject, func(dst *UploadRequest, subject auth.Subject) error {
				dst.Subject = subject
				return nil
			}),

			hapi.StrFromHeader(hapi.StrParam[UploadRequest]{Name: "test-header", IntoModel: func(dst *UploadRequest, value string) error {
				dst.TestHeader = value
				return nil
			}}),
			// this can be a simple alternative
			/*hapi.JSONFromBody(func(dst *UploadRequest, model UploadMetadata) error {
				dst.Metadata = model
				return nil
			}),*/
			hapi.JSONFromFormField("meta", func(dst *UploadRequest, model UploadMetadata) error {
				dst.Metadata = model
				return nil
			}),
			hapi.FilesFromFormField("files", func(dst *UploadRequest, files []*multipart.FileHeader) error {
				dst.Files = files
				return nil
			}),
		).
		Response(hapi.ToJSON[UploadRequest, UploadResponse](func(in UploadRequest) (UploadResponse, error) {
			fmt.Println(in.Subject.Valid(), in.Subject.ID(), slices.Collect(in.Subject.Permissions()))
			return UploadResponse{ID: "1234-" + in.TestHeader, When: time.Now()}, nil
		}))
}
