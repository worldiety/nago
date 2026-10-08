// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nais

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// transport puts the current access token on every request of the Anthropic client. If the service refuses it, the
// token is renewed, or the instance enrolls again, and the request is sent once more. So the Anthropic provider works
// against the service as it is, without knowing anything about tokens.
type transport struct {
	service *Service
	base    http.RoundTripper
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := rewindable(req)
	if err != nil {
		return nil, err
	}

	for attempt := 0; ; attempt++ {
		access, err := t.service.Access(req.Context())
		if err != nil {
			return nil, err
		}

		r := req.Clone(req.Context())
		r.Header.Set("x-api-key", access)
		r.Header.Del("Authorization")
		if body != nil {
			r.Body, _ = body()
		}

		resp, err := t.base.RoundTrip(r)
		if err != nil || resp.StatusCode != http.StatusUnauthorized || attempt > 0 {
			return resp, err
		}

		buf, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
		t.service.Reject(access, errorOf(resp.StatusCode, buf))
	}
}

// rewindable returns a function that yields the body anew for each attempt.
func rewindable(req *http.Request) (func() (io.ReadCloser, error), error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}

	if req.GetBody != nil {
		return req.GetBody, nil
	}

	buf, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot read the request body: %w", err)
	}

	return func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buf)), nil }, nil
}
