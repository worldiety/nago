// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"fmt"
	"go.wdy.de/nago/application/token"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/oas/v31"
	"net/http"
	"strings"
)

type requestSchema[In any] struct {
	contentType      string
	schema           *oas.Schema
	fieldName        string
	intoModel        func(dst *In, writer http.ResponseWriter, request *http.Request) error
	requestDecorator func(w http.ResponseWriter, r *http.Request) (*http.Request, error)
}

type RequestBuilder[In any] struct {
	ep              *EndpointBuilder[In]
	security        []*oas.SecurityRequirement
	inputStrHeaders map[string]StrParam[In]
	handlers        []requestSchema[In]
	//handler         func(in In, writer http.ResponseWriter, request *http.Request) error

}

func newRequestBuilder[In any](ep *EndpointBuilder[In]) *RequestBuilder[In] {
	return &RequestBuilder[In]{
		ep:              ep,
		inputStrHeaders: map[string]StrParam[In]{},
	}
}

func StrFromHeader[In any](field StrParam[In]) RequestOption[In] {
	return func(doc *oas.OpenAPI, b *RequestBuilder[In]) {
		if _, ok := b.inputStrHeaders[field.Name]; ok {
			panic(fmt.Errorf("duplicate header field %s", field.Name))
		}

		b.inputStrHeaders[field.Name] = field
	}
}

type RequestOption[In any] func(*oas.OpenAPI, *RequestBuilder[In])

type StrParam[T any] struct {
	Name        string
	Description string
	Required    bool
	Deprecated  bool
	IntoModel   func(dst *T, value string) error
}

const bearerAuthSecName = "bearerAuth"

// BearerAuth requires that an API bearer token must be submitted as header value for request authentication.
// This is the default. The api keys can be configured through the according web ui element.
func BearerAuth[In any](authenticate token.AuthenticateSubject) RequestOption[In] {
	return func(doc *oas.OpenAPI, b *RequestBuilder[In]) {
		if _, ok := doc.ComponentsSecurity()[bearerAuthSecName]; !ok {
			doc.ComponentsSecurity()[bearerAuthSecName] = &oas.SecurityScheme{
				Type:         "http",
				Description:  "A configured Nago API Token.",
				Name:         "API Token",
				Scheme:       "bearer",
				BearerFormat: "Opaque", // we have no special format, its just a random token
			}
		}

		b.handlers = append(b.handlers, requestSchema[In]{
			requestDecorator: func(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
				authHeader := r.Header.Get("Authorization")

				if authHeader == "" {
					http.Error(w, "Authorization header missing", http.StatusUnauthorized)
					return nil, user.InvalidSubjectErr
				}

				const prefix = "Bearer "

				if !strings.HasPrefix(authHeader, prefix) {
					http.Error(w, "Invalid auth header format", http.StatusUnauthorized)
					return nil, user.InvalidSubjectErr
				}

				tokenStr := strings.TrimPrefix(authHeader, prefix)
				subj, err := authenticate(token.Plaintext(tokenStr))
				if err != nil {
					http.Error(w, "Authorization header missing", http.StatusInternalServerError)
					return nil, err
				}

				newCtx := auth.WithSubject(r.Context(), subj)
				r = r.WithContext(newCtx)
				return r, nil
			},
		})

		b.security = append(b.security, &oas.SecurityRequirement{
			bearerAuthSecName: []string{},
		})
	}
}
