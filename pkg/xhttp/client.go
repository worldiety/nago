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
	"encoding/base64"
	"encoding/json"
	"errors"
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
	// rpsSem serializes the throttled requests, so that the waiting time is computed correctly. A channel
	// instead of a mutex lets a cancelled request give up waiting.
	rpsSem  chan struct{}
	semOnce sync.Once
	rps     int
	// lastReqAt is when the last request has been released, guarded by rpsSem. The monotonic clock of a
	// time.Time keeps a stepped wall clock from stalling or bursting the group.
	lastReqAt time.Time
	debugLog  bool
	debugCtr  atomic.Int64
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

// throttle waits until the next request may be sent according to the rate limit. A cancelled context gives up
// waiting and does not consume a slot, so that e.g. stopped sub-agents do not delay the others.
func (r *RequestGroup) throttle(ctx context.Context) error {
	r.semOnce.Do(func() { r.rpsSem = make(chan struct{}, 1) })
	if ctx == nil {
		ctx = context.Background()
	}

	// A done context must win deterministically, otherwise a dead request would take a slot by chance and delay
	// the next living one by a whole interval.
	if err := ctx.Err(); err != nil {
		return err
	}

	// all goroutines are serialized into a sequence, so that the waiting time is computed correctly
	select {
	case r.rpsSem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-r.rpsSem }()

	if err := ctx.Err(); err != nil {
		return err
	}

	betweenRequest := time.Second / time.Duration(r.rps)
	if !r.lastReqAt.IsZero() {
		if waitTime := betweenRequest - time.Since(r.lastReqAt); waitTime > 0 {
			if r.debugLog {
				slog.Info("xhttp.Do throttle", "wait", waitTime, "rps", r.rps)
			}

			timer := time.NewTimer(waitTime)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	r.lastReqAt = time.Now()
	return nil
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

// Timeout bounds the whole request including reading the response, by default 60 seconds. For a response
// which is handed over to the caller (see [Request.ToCloser]) it only bounds the arrival of the response:
// the body is read for as long as the caller likes, bounded by the context and by the timeout of an
// [http.Client] the caller sets.
func (r *Request) Timeout(timeout time.Duration) *Request {
	r.timeout = timeout
	return r
}

// Retry enables an internal retry-mechanics which is used to retry on connection errors and a 503 status
// (with [Request.Assert2xx]), not on other protocol errors. The retry sleep time uses exponential backoff, see
// also [Request.RetryWait]. Each attempt counts against the rate limit of the [RequestGroup]. Negative values
// mean no retry.
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

func (r *Request) BasicAuthentication(username, password string) *Request {
	auth := username + ":" + password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	r.Header("Authorization", "Basic "+encoded)
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

// To processes the response body. A limit set by [Request.ToLimit] applies, in whatever order both are called.
func (r *Request) To(fn func(r io.Reader) error) *Request {
	r.respBody = func(body io.ReadCloser) error {
		return fn(body)
	}

	return r
}

// ToCloser hands the response body over to the caller, e.g. to read a stream of events. The caller must close
// it, which also releases the request and its connection. The request is not bounded by [Request.Timeout].
func (r *Request) ToCloser(body func(readCloser io.ReadCloser)) *Request {
	r.leakResponse = true
	r.respBody = func(closer io.ReadCloser) error {
		body(closer)
		return nil
	}

	return r
}

// ToLimit limits the response body read by any To* method to the given amount of bytes. A longer body is cut
// off silently; [Request.ToJSON] mentions the limit, when the cut body does not decode.
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
				if int64(len(buf)) >= r.respLimit {
					err = fmt.Errorf("%w (the response body reached the limit of %d bytes and may be cut off)", err, r.respLimit)
				}
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
		client = &http.Client{}
		if !r.leakResponse {
			// a handed over body is read for as long as the caller likes
			client.Timeout = timeout
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

	}

	throttle := func(ctx context.Context) error {
		if grp := r.group; grp != nil && grp.rps > 0 {
			return grp.throttle(ctx)
		}
		return nil
	}
	if err := throttle(ctx); err != nil {
		return err
	}

	// A response which is handed over to the caller is read for as long as the caller likes, e.g. a stream of
	// events, so only its arrival is bounded by the timeout. Its context is cancelled when the caller closes
	// the body, which releases the connection.
	var release context.CancelFunc
	var arrival *time.Timer
	var timedOut atomic.Bool
	if r.leakResponse {
		ctx, release = context.WithCancel(ctx)
		arrival = time.AfterFunc(timeout, func() {
			timedOut.Store(true)
			release()
		})
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	fail := func(err error) error {
		if arrival != nil {
			arrival.Stop()
		}
		if release != nil {
			release()
		}
		return err
	}

	// An error of the request construction cannot be fixed by a retry, so the first request is built in
	// advance. Every further attempt builds its own request with a fresh body, because a body reader cannot be
	// rewound in general, and counts against the rate limit like any other request. The wait between the
	// attempts grows and gives up with the context.
	retry := max(r.retry, 0)
	waitTime := r.retryWaitDuration()
	req, err := r.newRequest(ctx, method, reqUrl)
	if err != nil {
		return fail(err)
	}

	var resp *http.Response
	for attempt := range retry + 1 {
		if attempt > 0 {
			if err = throttle(ctx); err != nil {
				break
			}
			if req, err = r.newRequest(ctx, method, reqUrl); err != nil {
				return fail(err)
			}
		}

		resp, err = client.Do(req)
		if err == nil && r.assert2xx && resp.StatusCode == http.StatusServiceUnavailable {
			// try to work against unreliable services, but report status and body like for any other status
			buf, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
			_ = resp.Body.Close()
			err = UnexpectedStatusCodeError{resp.StatusCode, buf}
		}

		if err == nil {
			break
		}

		// A cancelled or expired context fails every further attempt the same way.
		if ctx.Err() != nil || attempt == retry {
			break
		}

		slog.Warn("request failed, wait and retry", "try", attempt+1, "wait", waitTime, "err", err.Error())
		if werr := sleepContext(ctx, waitTime); werr != nil {
			err = werr
			break
		}

		waitTime += waitTime * time.Duration(attempt+1)
	}

	if arrival != nil {
		arrival.Stop()
	}

	if err != nil {
		if timedOut.Load() {
			err = context.DeadlineExceeded
		}
		var status UnexpectedStatusCodeError
		if errors.As(err, &status) {
			return fail(status)
		}
		return fail(fmt.Errorf("request failed: %w", err))
	}

	body := resp.Body
	if r.leakResponse {
		body = &releasingBody{ReadCloser: body, release: release}
	} else {
		defer resp.Body.Close()
	}

	if r.assert2xx {
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			lr := io.LimitReader(resp.Body, 1024*1024)
			buf, _ := io.ReadAll(lr)
			_ = body.Close()

			if grp := r.group; grp != nil {
				if grp.debugLog && DevelopmentBuild() {
					fmt.Println(string(buf))
				}
			}

			return UnexpectedStatusCodeError{resp.StatusCode, buf}
		}
	}

	if r.respBody == nil {
		// nobody reads the body, so a handed over one is released right away
		if r.leakResponse {
			_ = body.Close()
		}
		return nil
	}

	if r.respLimit > 0 {
		body = &limitedBody{Reader: io.LimitReader(body, r.respLimit), Closer: body}
	}

	if err := r.respBody(body); err != nil {
		return fail(fmt.Errorf("failed to parse response body: %w", err))
	}

	return nil
}

// newRequest builds the request of one attempt.
func (r *Request) newRequest(ctx context.Context, method, reqUrl string) (*http.Request, error) {
	var body io.Reader
	if r.body != nil {
		b, err := r.body()
		if err != nil {
			return nil, fmt.Errorf("failed to create request body: %w", err)
		}

		body = b
	}

	req, err := http.NewRequestWithContext(ctx, method, reqUrl, body)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// sleepContext waits for d or until ctx is done, whichever comes first.
func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// releasingBody is a handed over response body, which releases the request when it is closed.
type releasingBody struct {
	io.ReadCloser
	release context.CancelFunc
	once    sync.Once
}

func (b *releasingBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(b.release)
	return err
}

// limitedBody reads at most a limit of bytes, but closes the whole body.
type limitedBody struct {
	io.Reader
	io.Closer
}

type UnexpectedStatusCodeError struct {
	StatusCode int
	Body       []byte
}

func (e UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d: %s", e.StatusCode, excerpt(e.Body))
}

type ErrorWithBody struct {
	Cause error
	Body  []byte
}

func (e ErrorWithBody) Error() string {
	return fmt.Sprintf("%s: %s", e.Cause.Error(), excerpt(e.Body))
}

// excerpt keeps error messages readable. The Body fields of the errors stay complete for the callers.
func excerpt(b []byte) string {
	const limit = 4 * 1024
	if len(b) <= limit {
		return string(b)
	}
	return fmt.Sprintf("%s... (%d bytes)", b[:limit], len(b))
}

func (e ErrorWithBody) Unwrap() error {
	return e.Cause
}

// DevelopmentBuild returns true if this process is likely run on a developers machine
func DevelopmentBuild() bool {
	_, ok := os.LookupEnv("XPC_SERVICE_NAME")
	return ok
}
