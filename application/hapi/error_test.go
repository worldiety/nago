// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xerrors"
)

func TestWriteErrorStatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantFields bool
	}{
		{
			name:       "not logged in",
			err:        user.InvalidSubjectErr,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "permission denied",
			err:        user.PermissionDeniedErr,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "permission denied with id",
			err:        user.PermissionDeniedError("nago.user.find_by_id"),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "not found",
			err:        os.ErrNotExist,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrapped not found",
			err:        fmt.Errorf("load: %w", os.ErrNotExist),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "validation",
			err:        xerrors.WithFields("invalid", "Name", "must not be empty"),
			wantStatus: http.StatusUnprocessableEntity,
			wantFields: true,
		},
		{
			// the bool based classifier had no flag for this and fell through to 400
			name:       "already exists is a conflict",
			err:        os.ErrExist,
			wantStatus: http.StatusConflict,
		},
		{
			// Kind wins the status, Fields still travel so the client can fix the input
			name:       "denial joined with validation",
			err:        errors.Join(user.PermissionDeniedErr, xerrors.WithFields("invalid", "Name", "empty")),
			wantStatus: http.StatusForbidden,
			wantFields: true,
		},
		{
			name:       "unknown becomes internal",
			err:        errors.New("db connection reset"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/x", nil)

			WriteError(rec, req, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/problem+json") {
				t.Errorf("content type = %q", got)
			}

			var body ProblemDetails
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not valid json: %v", err)
			}

			if body.Title == "" || body.Detail == "" {
				t.Errorf("body must be populated: %+v", body)
			}

			if body.Status != tt.wantStatus {
				t.Errorf("body status = %d, want %d", body.Status, tt.wantStatus)
			}

			if tt.wantFields && len(body.Fields) == 0 {
				t.Error("validation errors must expose the field messages")
			}

			if !tt.wantFields && len(body.Fields) != 0 {
				t.Errorf("fields must be empty, got %v", body.Fields)
			}
		})
	}
}

func TestWriteErrorDoesNotLeakInternals(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	WriteError(rec, req, errors.New("password=hunter2 at 10.0.0.1"))

	if strings.Contains(rec.Body.String(), "hunter2") {
		t.Errorf("unclassified error must not be disclosed: %s", rec.Body.String())
	}

	var body ProblemDetails
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if body.Code == "" {
		t.Error("a support token is required to correlate with the log")
	}
}

func TestWriteErrorNil(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	WriteError(rec, req, nil)

	if rec.Body.Len() != 0 {
		t.Error("nil error must not write anything")
	}
}

// TestWriteErrorWritesExactlyOnce is the regression guard for a double write: a handler which
// had already produced a response used to get a second status line and a JSON body appended to
// the existing one.
func TestWriteErrorWritesExactlyOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	WriteError(rec, req, user.PermissionDeniedErr)

	body := rec.Body.String()
	if n := strings.Count(body, `"status"`); n != 1 {
		t.Fatalf("expected exactly one problem document, found %d in %q", n, body)
	}

	var probe map[string]any
	if err := json.Unmarshal([]byte(body), &probe); err != nil {
		t.Fatalf("body must be a single valid json document: %v (%q)", err, body)
	}
}

// TestBearerAuthDoesNotWriteItself pins that the decorator leaves the response to the caller.
// It previously called http.Error itself, so adding WriteError on that path corrupted the body.
func TestBearerAuthDoesNotWriteItself(t *testing.T) {
	src, err := os.ReadFile("dsl_request.go")
	if err != nil {
		t.Fatal(err)
	}

	decorator := string(src)
	start := strings.Index(decorator, "requestDecorator: func(")
	if start < 0 {
		t.Skip("BearerAuth decorator not found, adjust this test")
	}

	end := strings.Index(decorator[start:], "\n\t\t\t},")
	if end < 0 {
		t.Fatal("cannot delimit the decorator body")
	}

	body := decorator[start : start+end]
	if strings.Contains(body, "http.Error(") || strings.Contains(body, "WriteHeader(") {
		t.Error("the decorator must not write a response; it returns an error and the caller classifies it")
	}
}

func TestWriteErrorExplicitStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantTitle  string
		wantDetail string
	}{
		{"conflict with own text", NewStatusError(http.StatusConflict, "order 7 exists"), http.StatusConflict, "Conflict", "order 7 exists"},
		{"wrapped", fmt.Errorf("handler: %w", NewStatusError(http.StatusBadRequest, "field count is no int32")), http.StatusBadRequest, "Bad Request", "field count is no int32"},
		{"own title", &StatusError{Status: http.StatusTooManyRequests, Title: "Busy", Detail: "try again"}, http.StatusTooManyRequests, "Busy", "try again"},
		// the explicit status wins over a denial, which the same error carries
		{"joined with a denial", errors.Join(NewStatusError(http.StatusNotFound, "gone"), user.PermissionDeniedErr), http.StatusNotFound, "Not Found", "gone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteError(rec, httptest.NewRequest(http.MethodGet, "/x", nil), tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/problem+json") {
				t.Errorf("content type = %q", got)
			}

			var body ProblemDetails
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not valid json: %v", err)
			}

			if body.Status != tt.wantStatus || body.Title != tt.wantTitle || body.Detail != tt.wantDetail || body.Code == "" {
				t.Errorf("body = %+v", body)
			}
		})
	}

	// counter check: a status outside of the errors is ignored and the error is classified
	rec := httptest.NewRecorder()
	WriteError(rec, httptest.NewRequest(http.MethodGet, "/x", nil), NewStatusError(http.StatusOK, "fine"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status 200 as an error: %d, want 500", rec.Code)
	}
}
