// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Options struct {
	RegisterHandler     func(method string, pattern string, handler http.HandlerFunc)
	OperationConfigured func(op Operation)
}

type API struct {
	operations []Operation
	opts       Options
}

func NewAPI(opts Options) *API {
	return &API{opts: opts}
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

func Handle[In, Out any](api *API, op Operation, fn func(ctx context.Context, in *In) (*Out, error)) {
	if op.Path == "" {
		panic(fmt.Errorf("empty path is not allowed"))
	}

	if op.Method == "" {
		op.Method = http.MethodGet
	}

	for _, operation := range api.operations {
		if operation.Path == op.Path && op.Method == operation.Method {
			panic(fmt.Errorf("path %s has already been configured for method %s: %s", op.Path, op.Method, operation.Summary))
		}
	}

	api.operations = append(api.operations, op)
	api.opts.RegisterHandler(op.Method, op.Path, func(writer http.ResponseWriter, request *http.Request) {
		var in In
		out, err := fn(request.Context(), &in)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		buf, err := json.Marshal(out)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write(buf); err != nil {
			slog.Error("failed to write response", "err", err.Error())
			return
		}
	})

	if api.opts.OperationConfigured != nil {
		api.opts.OperationConfigured(op)
	}
}

type ContentType string

func (h *ContentType) Read(r *http.Request) error {
	*h = ContentType(r.Header.Get("Content-Type"))
	return nil
}

func (h *ContentType) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", string(*h))
	return nil
}

type Body[T any] struct {
	Body T
}

func (h *Body[T]) Read(r *http.Request) error {
	return nil
}

func (h *Body[T]) Write(w http.ResponseWriter) error {
	return nil
}
