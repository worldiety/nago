// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/xerror"
)

// ProblemDetails is the error body of the HTTP layer, modelled after RFC 9457.
//
// Only classified, display safe information is included, see [xerror.Present]. The original
// error is never serialized; Code correlates the response with the log entry instead.
type ProblemDetails struct {
	// Title is the short, human readable and localized summary.
	Title string `json:"title"`
	// Detail is the localized explanation.
	Detail string `json:"detail"`
	// Status repeats the HTTP status code.
	Status int `json:"status"`
	// Fields maps field names or paths to validation messages, empty unless Status is 422.
	Fields map[string]string `json:"fields,omitempty"`
	// Code is a stable token of the underlying error for support and log correlation.
	Code string `json:"code,omitempty"`
}

// statusOf maps a classified error onto an HTTP status code.
//
// Previously every business error became a 400, so a client could not tell a denied request
// from a malformed one, and a missing element from a validation failure.
func statusOf(p xerror.Presentation) int {
	switch p.Kind {
	case xerror.KindNotLoggedIn:
		return http.StatusUnauthorized
	case xerror.KindDenied:
		return http.StatusForbidden
	case xerror.KindNotFound:
		return http.StatusNotFound
	case xerror.KindAlreadyExists:
		return http.StatusConflict
	case xerror.KindValidation, xerror.KindPasswordTooWeak:
		return http.StatusUnprocessableEntity
	case xerror.KindLocalized:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// WriteError classifies err and writes it as a problem+json response.
//
// It must be called before anything else was written to writer. The full error is logged
// together with the token which the client receives.
func WriteError(writer http.ResponseWriter, request *http.Request, err error) {
	if err == nil {
		return
	}

	p := xerror.PresentOrGeneric(bundlerOf(request), err)
	status := statusOf(p)

	body := ProblemDetails{
		Title:  p.Title,
		Detail: p.Message,
		Status: status,
		Code:   p.Token(),
	}

	// Field messages are actionable for the client regardless of the status: a join of a
	// denial and a validation error still tells it which field to correct.
	if p.Validation() {
		body.Fields = p.Fields
	}

	buf, mErr := json.Marshal(body)
	if mErr != nil {
		slog.Error("hapi: cannot encode problem details", "err", mErr.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Error("hapi: failed to handle request",
		"err", err.Error(),
		"kind", p.Kind.String(),
		"status", status,
		"code", body.Code,
		"path", requestPath(request),
	)

	writer.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	writer.WriteHeader(status)

	if _, wErr := writer.Write(buf); wErr != nil {
		slog.Error("hapi: cannot write problem details", "err", wErr.Error())
	}
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}

	return r.URL.Path
}

// bundlerOf resolves the localization bundle of the request context, if the middleware chain
// installed one. Returning nil is fine, [xerror.Present] then falls back to a default bundle.
//
// TODO honour the Accept-Language header when no bundle was installed.
func bundlerOf(r *http.Request) i18n.Bundler {
	if r == nil {
		return nil
	}

	if b, ok := i18n.BundleFrom(r.Context()); ok {
		return b
	}

	return nil
}
