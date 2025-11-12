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
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RequestGroup struct {
	rpsMutex  sync.Mutex
	rps       int
	lastReqAt atomic.Int64
	debugLog  bool
	debugCtr  atomic.Int64
	leakBody  bool
}

func NewRequestGroup() *RequestGroup {
	return &RequestGroup{}
}

func (r *RequestGroup) DebugLog(debugLog bool) *RequestGroup {
	r.debugLog = debugLog
	return r
}

func (r *RequestGroup) RateLimit(rps int) *RequestGroup {
	r.rps = rps
	return r
}

type Request struct {
	client       *http.Client
	ctx          context.Context
	timeout      time.Duration
	url          string
	baseUrl      string
	headers      map[string]string
	query        map[string]string
	body         func() (io.Reader, error)
	respBody     func(closer io.ReadCloser) error
	assert2xx    bool
	respLimit    int64
	retry        int
	retryWait    time.Duration
	group        *RequestGroup
	leakResponse bool
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

// Retry enables an internal retry-mechanics which is used to retry on connection errors, not higher level
// protocol errors. The retry sleep time uses exponential backoff. See also [Request.RetryWait].
func (r *Request) Retry(retry int) *Request {
	r.retry = retry
	return r
}

// Group sets the group to which this request shall belong.
func (r *Request) Group(group *RequestGroup) *Request {
	r.group = group
	return r
}

// RetryWait sets the base duration for retries. Defaults to 50ms.
func (r *Request) RetryWait(retryWait time.Duration) *Request {
	r.retryWait = retryWait
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

		var debug bool
		if r.group != nil {
			debug = r.group.debugLog
		}
		if debug {
			if DevelopmentBuild() {
				fmt.Println(string(b))
			}

			slog.Info("prepared JSON request body", "body", string(b))
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
	if r.respLimit == 0 {
		r.respBody = func(closer io.ReadCloser) error {
			return fn(closer)
		}
	} else {
		r.respBody = func(body io.ReadCloser) error {
			lr := io.LimitReader(body, r.respLimit)
			return fn(lr)
		}
	}

	return r
}

// ToCloser keeps the body open and leaks the entire returned response reader. The caller is responsible for closing
// and releasing the associated resources.
func (r *Request) ToCloser(body func(readCloser io.ReadCloser)) *Request {
	r.leakResponse = true
	r.respBody = func(closer io.ReadCloser) error {
		body(closer)
		return nil
	}

	return r
}

// ToLimit installs a limited reader when processing with any To* response.
func (r *Request) ToLimit(limit int64) *Request {
	r.respLimit = limit
	return r
}

// ToJSON accepts a json response and unmarshal into the given pointer. If a limit is configured, the
// response will be buffered and returned in the error for debugging purpose. Otherwise, the stream decoder
// is used.
func (r *Request) ToJSON(v any) *Request {
	r.To(func(reader io.Reader) error {
		if r.respLimit > 0 {
			buf, err := io.ReadAll(reader)
			if err != nil {
				return err
			}

			var debug bool
			if r.group != nil {
				debug = r.group.debugLog
			}
			if debug {
				if DevelopmentBuild() {
					fmt.Println(string(buf))
				}

				slog.Info("received JSON response body", "body", string(buf))
			}

			err = json.Unmarshal(buf, v)
			if err != nil {
				return ErrorWithBody{
					Cause: err,
					Body:  buf,
				}
			}

			return err
		}

		return json.NewDecoder(reader).Decode(v)
	})

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

func (r *Request) retryWaitDuration() time.Duration {
	if r.retryWait == 0 {
		return time.Millisecond * 50
	}

	return r.retryWait
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
		u, err := url.Parse(reqUrl)
		if err != nil {
			return fmt.Errorf("invalid url %s: %w", r.url, err)
		}

		queryValues := u.Query()
		for key, value := range r.query {
			queryValues.Set(key, value)
		}

		u.RawQuery = queryValues.Encode()
		reqUrl = u.String()
	}

	if grp := r.group; grp != nil {

		if grp.debugLog {
			id := grp.debugCtr.Add(1)
			slog.Info("xhttp.Do", "id", id, "method", method, "url", reqUrl, "timeout", timeout)
			start := time.Now()

			defer func() {
				slog.Info("xhttp.Do done", "id", id, "method", method, "url", reqUrl, "duration", time.Since(start))
			}()
		}

		if grp.rps > 0 {
			// we must serialize all goroutines into a sequence so that waiting duration calculation is correct
			grp.rpsMutex.Lock()

			delta := time.Duration(time.Now().UnixMilli()-grp.lastReqAt.Load()) * 1000 * 1000
			betweenRequest := time.Second / time.Duration(grp.rps)
			if delta < betweenRequest {
				waitTime := betweenRequest - delta
				if grp.debugLog {
					slog.Info("xhttp.Do throttle", "wait", waitTime, "rps", grp.rps)
				}

				time.Sleep(waitTime)
			}

			grp.lastReqAt.Store(time.Now().UnixMilli())

			grp.rpsMutex.Unlock()
		}

	}

	var body io.Reader
	if r.body != nil {
		b, err := r.body()
		if err != nil {
			return fmt.Errorf("failed to create request body: %w", err)
		}

		body = b
	}
	if !r.leakResponse {
		// TODO it is unclear, how this should behave. We must actually couple the close of the leaked reader with this cancel to release earlier
		// this way we don't have forced timeouts at all
		c, cancel := context.WithTimeout(ctx, timeout)
		ctx = c
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqUrl, body)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	waitTime := r.retryWaitDuration()
	try := time.Duration(0)
	doFn := func() (*http.Response, error) {
		var resp *http.Response
		var err error
		for range r.retry + 1 {
			try++
			resp, err = client.Do(req)
			if r.assert2xx && err == nil && resp.StatusCode == http.StatusServiceUnavailable {
				// try work against unreliable services
				_ = resp.Body.Close()
				err = fmt.Errorf("request failed with status code %v", resp.StatusCode)
			}

			if err != nil {
				if r.retry > 0 {
					slog.Warn("request failed, wait and retry", "try", try, "wait", waitTime, "err", err.Error())
					time.Sleep(waitTime)
					waitTime = waitTime + waitTime*try
				} else {
					break
				}

				continue
			}

			break
		}

		return resp, err
	}

	resp, err := doFn()
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if !r.leakResponse {
		defer resp.Body.Close()
	}

	if r.assert2xx {
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			lr := io.LimitReader(resp.Body, 1024*1024)
			buf, _ := io.ReadAll(lr)

			if grp := r.group; grp != nil {
				if grp.debugLog && DevelopmentBuild() {
					fmt.Println(string(buf))
				}
			}

			return UnexpectedStatusCodeError{resp.StatusCode, buf}
		}
	}

	if r.respBody != nil {
		if err := r.respBody(resp.Body); err != nil {
			return fmt.Errorf("failed to parse response body: %w", err)
		}
	}

	return nil
}

type UnexpectedStatusCodeError struct {
	StatusCode int
	Body       []byte
}

func (e UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d: %s", e.StatusCode, string(e.Body))
}

type ErrorWithBody struct {
	Cause error
	Body  []byte
}

func (e ErrorWithBody) Error() string {
	return fmt.Sprintf("%s: %s", e.Cause.Error(), string(e.Body))
}

func (e ErrorWithBody) Unwrap() error {
	return e.Cause
}

// DevelopmentBuild returns true if this process is likely run on a developers machine
func DevelopmentBuild() bool {
	_, ok := os.LookupEnv("XPC_SERVICE_NAME")
	return ok
}
