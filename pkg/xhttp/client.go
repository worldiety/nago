// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	client    *http.Client
	ctx       context.Context
	timeout   time.Duration
	url       string
	baseUrl   string
	headers   map[string]string
	query     map[string]string
	body      func() (io.Reader, error)
	respBody  func(io.Reader) error
	assert2xx bool
}

func NewRequest() *Request {
	return &Request{}
}

func (r *Request) Context(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) URL(url string) *Request {
	r.url = url
	return r
}

func (r *Request) BaseURL(base string) *Request {
	r.baseUrl = base
	return r
}

// Client uses the given Client (and transport pool) for communication. May be nil to create a new client for
// each request on the fly.
func (r *Request) Client(c *http.Client) *Request {
	r.client = c
	return r
}

// Timeout sets the timeout to use. By default, the timeout is
func (r *Request) Timeout(timeout time.Duration) *Request {
	r.timeout = timeout
	return r
}

func (r *Request) Header(key, value string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *Request) Query(key, value string) *Request {
	if r.query == nil {
		r.query = map[string]string{}
	}
	r.query[key] = value
	return r
}

func (r *Request) BearerAuthentication(token string) *Request {
	r.Header("Authorization", "Bearer "+token)
	return r
}

func (r *Request) Assert2xx(assert2xx bool) *Request {
	r.assert2xx = assert2xx
	return r
}

// BodyJSON marshals the given value as json and encodes it as the request body.
func (r *Request) BodyJSON(v any) *Request {
	r.body = func() (io.Reader, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(b), nil
	}

	r.Header("Content-Type", "application/json")

	return r
}

func (r *Request) Body(fn func() (io.Reader, error)) *Request {
	r.body = fn
	return r
}

func (r *Request) To(fn func(r io.Reader) error) *Request {
	r.respBody = fn
	return r
}

// ToJSON accepts a json response and unmarshal into the given pointer.
func (r *Request) ToJSON(v any) *Request {
	r.respBody = func(r io.Reader) error {
		return json.NewDecoder(r).Decode(v)
	}

	r.Header("Accept", "application/json")
	return r
}

func (r *Request) Post() error {
	return r.Do(http.MethodPost)
}

func (r *Request) Get() error {
	return r.Do(http.MethodGet)
}

func (r *Request) Patch() error {
	return r.Do(http.MethodPatch)
}

func (r *Request) Delete() error {
	return r.Do(http.MethodDelete)
}

func (r *Request) Put() error {
	return r.Do(http.MethodPut)
}

func (r *Request) Do(method string) error {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := r.timeout
	if timeout == 0 {
		timeout = time.Second * 60
	}

	client := r.client
	if client == nil {
		client = &http.Client{
			Timeout: timeout,
		}
	}

	reqUrl := r.url
	if r.baseUrl != "" {
		a := strings.TrimRight(r.baseUrl, "/")
		b := strings.TrimLeft(r.url, "/")
		reqUrl = a + "/" + b
	}

	if len(r.query) > 0 {
		u, err := url.Parse(r.url)
		if err != nil {
			return fmt.Errorf("invalid url %q: %w", r.url, err)
		}

		queryValues := u.Query()
		for key, value := range r.query {
			queryValues.Set(key, value)
		}

		u.RawQuery = queryValues.Encode()
		reqUrl = u.String()
	}

	var body io.Reader
	if r.body != nil {
		b, err := r.body()
		if err != nil {
			return fmt.Errorf("failed to create request body: %w", err)
		}

		body = b
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, reqUrl, body)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if r.assert2xx {
		lr := io.LimitReader(resp.Body, 4*1024)
		buf, _ := io.ReadAll(lr)
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("unexpected status: got %v: %s", resp.Status, string(buf))
		}
	}

	if r.respBody != nil {
		if err := r.respBody(resp.Body); err != nil {
			return fmt.Errorf("failed to parse response body: %w", err)
		}
	}

	return nil
}
