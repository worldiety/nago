// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"fmt"
	"net/http"
	"strings"

	"go.wdy.de/nago/pkg/oas/v31"
)

// Spec is an endpoint, which brings its documentation as OpenAPI types instead of deriving it from Go types by
// reflection like the DSL ([Get], [Request], [ToJSON]). This fits endpoints, whose input and output are only known at
// runtime, e.g. those of a script. Nothing of Operation is derived or checked, it is the documentation as it is.
type Spec struct {
	// Method is the HTTP method like [http.MethodGet]. Default is GET.
	Method string
	// Path of the HTTP resource with a leading slash and any path params in brackets, e.g. /pet/{petId}.
	Path string
	// Operation is the documentation of the endpoint in the document.
	Operation *oas.Operation
	// Schemas are the named schemas, which Operation references as #/components/schemas/<name>. A name, which the
	// components of the document have already, keeps its schema.
	Schemas map[string]*oas.Schema
	// Handler serves the endpoint. Within a [Group], it reads the values of the path params with
	// [http.Request.PathValue].
	Handler http.HandlerFunc
}

// Handle registers an endpoint with an explicit schema at api. It fails, if the method is unknown, the path does not
// start with a slash, the handler is missing or the document describes this method at this path already.
func Handle(api *API, spec Spec) error {
	method := spec.Method
	if method == "" {
		method = http.MethodGet
	}

	if !strings.HasPrefix(spec.Path, "/") {
		return fmt.Errorf("hapi: the path %q of an endpoint does not start with a slash", spec.Path)
	}

	if spec.Handler == nil {
		return fmt.Errorf("hapi: the endpoint %s %s has no handler", method, spec.Path)
	}

	if api.doc.Paths == nil {
		api.doc.Paths = oas.Paths{}
	}

	item, ok := api.doc.Paths[spec.Path]
	if !ok {
		item = &oas.PathItem{}
	}

	slot := operationOf(item, method)
	if slot == nil {
		return fmt.Errorf("hapi: unsupported method %s of the endpoint %s", method, spec.Path)
	}

	if *slot != nil {
		return fmt.Errorf("hapi: the endpoint %s %s is described already", method, spec.Path)
	}

	op := spec.Operation
	if op == nil {
		op = &oas.Operation{}
	}

	*slot = op
	api.doc.Paths[spec.Path] = item

	if len(spec.Schemas) > 0 {
		schemas := api.doc.ComponentsSchemas()
		for name, s := range spec.Schemas {
			if _, ok := schemas[name]; !ok && s != nil {
				schemas[name] = s
			}
		}
	}

	if api.opts.RegisterHandler != nil {
		api.opts.RegisterHandler(method, spec.Path, spec.Handler)
	}

	return nil
}

// operationOf is the field of the method in the path item, nil for an unknown method.
func operationOf(item *oas.PathItem, method string) **oas.Operation {
	switch method {
	case http.MethodGet:
		return &item.Get
	case http.MethodPut:
		return &item.Put
	case http.MethodPost:
		return &item.Post
	case http.MethodDelete:
		return &item.Delete
	case http.MethodPatch:
		return &item.Patch
	case http.MethodHead:
		return &item.Head
	case http.MethodOptions:
		return &item.Options
	case http.MethodTrace:
		return &item.Trace
	default:
		return nil
	}
}
