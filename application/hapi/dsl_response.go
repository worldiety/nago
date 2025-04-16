// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"go.wdy.de/nago/pkg/oas/v31"
	"net/http"
)

type ResponseOption[In any] func(*oas.OpenAPI, *ResponseBuilder[In])
type ResponseBuilder[In any] struct {
	m *EndpointBuilder[In]

	contentType string
	schema      *oas.Schema
	handler     func(in In, writer http.ResponseWriter, request *http.Request)
}

func (r *ResponseBuilder[In]) handle(in In, writer http.ResponseWriter, request *http.Request) {
	if r == nil {
		writer.WriteHeader(http.StatusTeapot)
		return
	}

	r.handler(in, writer, request)
}
