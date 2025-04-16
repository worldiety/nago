// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"go.wdy.de/nago/pkg/oas/v31"
	"go.wdy.de/nago/pkg/std"
	"net/http"
)

type Options struct {
	RegisterHandler func(method string, pattern string, handler http.HandlerFunc)
}

type API struct {
	operations []Operation
	opts       Options
	doc        *oas.OpenAPI
}

func NewAPI(doc *oas.OpenAPI, opts Options) *API {
	return &API{doc: doc, opts: opts}
}

type Operation struct {
	// Method is the HTTP method like [http.MethodGet]. Default is GET.
	Method string
	// Path of the HTTP resource including a leading slash and any path params in brackets, e.g. /pet/{petId}.
	// This is required and must not be empty.
	Path string
	// Summary for the API documentation for this operation.
	Summary string
	// Description for the API documentation for this operation.
	Description string
	// Deprecated flag for the API documentation.
	Deprecated bool
}

const errorAlreadyHandled std.Error = "already handled"

// Doc applies the given function to mutate the current state of the OpenAPI. Note, that some sections
// may have not been initialized, because non-null or omitted elements are sometimes not allowed in the open api
// specification. Also, invocation and register order matters and parts may get overwritten.
func Doc(api *API, fn func(doc *oas.OpenAPI)) {
	fn(api.doc)
}
