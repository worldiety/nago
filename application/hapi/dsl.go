// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"fmt"
	"go.wdy.de/nago/pkg/oas/v31"
	"net/http"
)

type EndpointBuilder[In any] struct {
	api      *API
	op       Operation
	response *ResponseBuilder[In]
	request  *RequestBuilder[In]
	handler  *HandleFunc[In]
}

func Post[In any](api *API, op Operation) *EndpointBuilder[In] {
	op.Method = http.MethodPost
	return Endpoint[In](api, op)
}

func Get[In any](api *API, op Operation) *EndpointBuilder[In] {
	op.Method = http.MethodGet
	return Endpoint[In](api, op)
}

func Delete[In any](api *API, op Operation) *EndpointBuilder[In] {
	op.Method = http.MethodDelete
	return Endpoint[In](api, op)
}

func Put[In any](api *API, op Operation) *EndpointBuilder[In] {
	op.Method = http.MethodPut
	return Endpoint[In](api, op)
}

func Endpoint[In any](api *API, op Operation) *EndpointBuilder[In] {
	b := &EndpointBuilder[In]{
		op:  op,
		api: api,
	}

	return b
}

// Request defined arbitrary options which are usually mapper functions from the raw request input to something
// more structured. The goal of this additional abstraction, is to collect specific information about the request,
// to automatically generate the according open api documentation. See also [Response]. If a mapper fails,
// this is usually due to a bad request, which is communicated accordingly. If all request options has been applied,
// the [Response] options are evaluated.
func (b *EndpointBuilder[In]) Request(opts ...RequestOption[In]) *EndpointBuilder[In] {
	if b.request == nil {
		b.request = newRequestBuilder[In](b)
	}

	for _, opt := range opts {
		opt(b.api.doc, b.request)
	}
	return b
}

// Response defines arbitrary options which are usually mapper functions to modify the raw response output to be more
// structured. The goal of this additional abstraction is to collect specific information about the request,
// to automatically generate the according open api documentation. Invalid subjects are intentionally passed through,
// so that the use case can decide to either reject the request entirely or to just return public data.
func (b *EndpointBuilder[In]) Response(opts ...ResponseOption[In]) {
	if b.response == nil {
		b.response = &ResponseBuilder[In]{}
	}

	for _, opt := range opts {
		opt(b.api.doc, b.response)
	}

	b.register(b.api)
}

func (b *EndpointBuilder[In]) register(api *API) {
	pathItem, ok := api.doc.Paths[b.op.Path]
	if !ok {
		pathItem = &oas.PathItem{}

		api.doc.Paths[b.op.Path] = pathItem
	}

	var itemOp **oas.Operation
	switch b.op.Method {
	case http.MethodPost:
		itemOp = &pathItem.Post
	default:
		// developer error
		panic(fmt.Errorf("unsupported method: %s", b.op.Method))
	}

	if *itemOp != nil {
		// developer error
		panic(fmt.Errorf("method %s already registered", b.op.Method))
	}

	op := &oas.Operation{
		Summary:     b.op.Summary,
		Description: b.op.Description,
		Deprecated:  b.op.Deprecated,
	}

	if b.request != nil {
		for _, strHeader := range b.request.inputStrHeaders {
			op.Parameters = append(op.Parameters, oas.Parameter{
				Name:            strHeader.Name,
				In:              oas.LocationHeader,
				Description:     strHeader.Description,
				Required:        strHeader.Required,
				Deprecated:      strHeader.Deprecated,
				Schema:          schemaOf[string](api.doc),
				AllowEmptyValue: true,
			})
		}
	}

	if b.response != nil && b.response.schema != nil {

		if op.Responses == nil {
			op.Responses = map[string]*oas.Response{}
		}

		api.doc.AddSchemaIfUnknown(b.response.schema)
		op.Responses["200"] = &oas.Response{
			Content: map[string]oas.MediaType{
				b.response.contentType: oas.MediaType{
					Schema: &oas.Schema{
						Ref: b.response.schema.RefName(),
					},
				},
			},
		}
	}

	if b.request != nil {
		multipartFields := map[string]requestSchema[In]{}
		for _, schema := range b.request.handlers {
			if schema.contentType == "multipart/form-data" && schema.fieldName != "" {
				multipartFields[schema.fieldName] = schema
			}

		}

		for _, schema := range b.request.handlers {
			if schema.schema == nil {
				continue
			}
			api.doc.AddSchemaIfUnknown(schema.schema)
			op.RequestBody = &oas.RequestBody{
				Content: map[string]oas.MediaType{
					schema.contentType: oas.MediaType{
						Schema: &oas.Schema{
							Ref: schema.schema.RefName(),
						},
					},
				},
			}
		}

		if len(multipartFields) > 0 {
			// requestBody:
			//  required: true
			//  content:
			//    multipart/form-data:
			//      schema:
			//        type: object
			//        properties:
			//          files:
			//            type: array
			//            items:
			//              type: string
			//              format: binary
			//          metadata:
			//            $ref: '#/components/schemas/Metadata'

			props := map[string]*oas.Schema{}
			for fname, field := range multipartFields {
				props[fname] = field.schema
			}
			op.RequestBody.Content["multipart/form-data"] = oas.MediaType{
				Schema: &oas.Schema{
					Type:       "object",
					Properties: props,
				},
			}
		}

	}

	op.Security = b.request.security
	if len(op.Security) > 0 {
		op.Responses["401"] = &oas.Response{
			Description: "The authorization is missing. This is usually a bearer token. It may also indicate a wrong token format.",
		}
		op.Responses["403"] = &oas.Response{
			Description: "An authorization was given, but the subject does not have the required permission. This may also mean, that the associated user or api token has expired.",
		}
	}

	*itemOp = op

	// actual handler implementation
	api.opts.RegisterHandler(b.op.Method, b.op.Path, func(writer http.ResponseWriter, request *http.Request) {
		var input In
		if b.request != nil {
			for _, handler := range b.request.handlers {
				if handler.requestDecorator != nil {
					r, err := handler.requestDecorator(writer, request)
					if err != nil {
						writer.WriteHeader(http.StatusForbidden)
						return
					}

					request = r
				}
			}

			for _, strHeader := range b.request.inputStrHeaders {
				if strHeader.IntoModel != nil {
					if err := strHeader.IntoModel(&input, request.Header.Get(strHeader.Name)); err != nil {
						writer.WriteHeader(http.StatusBadRequest)
						return
					}
				}
			}

			for _, handler := range b.request.handlers {
				if handler.intoModel != nil {
					if err := handler.intoModel(&input, writer, request); err != nil {
						writer.WriteHeader(http.StatusBadRequest)
						return
					}
				}

			}

		}

		b.response.handle(input, writer, request)
	})
}

type HandleFunc[In any] func(In, http.ResponseWriter, *http.Request)
