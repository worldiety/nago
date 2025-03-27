// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"context"
	"errors"
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/oas/v30"
	"go.wdy.de/nago/pkg/std"
	"log/slog"
	"net/http"
	"reflect"
)

type Options struct {
	RegisterHandler     func(method string, pattern string, handler http.HandlerFunc)
	OperationConfigured func(op Operation, in Input, out Output)
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

type Input interface {
	Read(w http.ResponseWriter, t *http.Request) error
	DescribeInput(op *oas.Operation)
}

type Output interface {
	Write(w http.ResponseWriter, t *http.Request) error
	DescribeOutput(op *oas.Operation)
}

func Handle[In Input, Out Output](api *API, op Operation, fn func(ctx context.Context, in In) (Out, error)) {
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
		if err := in.Read(writer, request); err != nil {
			if errors.Is(err, errorAlreadyHandled) {
				return
			}

			if errors.Is(err, user.PermissionDeniedErr) {
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			if errors.Is(err, user.InvalidSubjectErr) {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			slog.Error("failed to read input", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		out, err := fn(request.Context(), in)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := out.Write(writer, request); err != nil {
			if errors.Is(err, errorAlreadyHandled) {
				return
			}

			if errors.Is(err, user.PermissionDeniedErr) {
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			if errors.Is(err, user.InvalidSubjectErr) {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			slog.Error("failed to write output", "err", err.Error())
			writer.WriteHeader(http.StatusInternalServerError) // be have already sent the header, but we don't know
			return
		}
	})

	if api.opts.OperationConfigured != nil {
		var zeroIn Input
		var zeroOut Output

		tin := reflect.TypeFor[In]()
		if tin.Kind() == reflect.Ptr {
			zeroIn = reflect.New(tin.Elem()).Interface().(Input)
		}

		tout := reflect.TypeFor[Out]()
		if tout.Kind() == reflect.Ptr {
			zeroOut = reflect.New(tout.Elem()).Interface().(Output)
		}

		api.opts.OperationConfigured(op, zeroIn, zeroOut)
	}
}

const errorAlreadyHandled std.Error = "already handled"
