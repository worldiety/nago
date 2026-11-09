// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nethost

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"golang.org/x/net/http/httpguts"
)

// The reasons of the outcomes. They are fixed texts, because the errors of Go contain the URL, which may contain
// a credential.
const (
	reasonTooMany     = "too many open requests"
	reasonConnection  = "the connection is not available"
	reasonUnreachable = "the server can not be reached"
	reasonTimeout     = "the request has timed out after it was sent"
	reasonAborted     = "the connection was lost after the request was sent"
	reasonTooLarge    = "the answer is too large"
	reasonNotDeclared = "the endpoint is not declared"
	reasonNoStreams   = "this host can not transfer files"
	reasonStream      = "the file is not available"
	reasonSource      = "the file can not be read"
	reasonStore       = "the answer can not be stored"
	reasonStalled     = "the transfer has stalled"
	reasonParts       = "the parts of the form are not valid"
	reasonSourceBody  = "a request can not have a source and a body"
	reasonPartsBody   = "a request with parts can not have a body or a source"
	reasonInvalid     = "the request is not valid"

	// the failures of a job, which ended without its function done
	reasonDropped   = "the request was not sent"
	reasonCancelled = "the request was cancelled"
)

const maxReason = 1024

// prepared is a checked request, or the result of a request, which was refused.
type prepared struct {
	conn    Connection
	req     net.Request
	target  *url.URL
	timeout time.Duration
	// method and route are the only parts of a request, which an observation may contain
	method string
	route  string

	refused bool
	failed  net.Result

	// streaming tells, that the request is a transfer, see [Options.StreamStall]
	streaming bool
	upload    *upload
	sink      Sink
	settled   bool // the sink has either kept its content or has been aborted
}

func (r *prepared) refuse(reason string) {
	r.refused, r.failed = true, net.Result{Outcome: net.Failed, Reason: reason}
}

func (r *prepared) cancelled() bool {
	return r.conn.Context != nil && r.conn.Context.Err() != nil
}

// discardSink aborts the sink, unless it was committed. It is safe to call it several times, but not concurrently.
func (r *prepared) discardSink() {
	if r.sink == nil || r.settled {
		return
	}
	r.settled = true
	defer func() {
		if v := recover(); v != nil {
			slog.Error("the sink of a request has failed", "panic", v)
		}
	}()
	r.sink.Abort()
}

// reasonOf returns the text of a [Refused] or the fixed reason otherwise.
func reasonOf(err error, hidden string) string {
	var refused Refused
	if errors.As(err, &refused) {
		return string(refused)
	}
	return hidden
}

// forbiddenHeaders are the headers, which a program must not set: credentials and what the transport controls.
var forbiddenHeaders = map[string]bool{
	"Authorization": true, "Proxy-Authorization": true, "Cookie": true, "Set-Cookie": true, "Host": true,
	"Connection": true, "Content-Length": true, "Transfer-Encoding": true, "Te": true, "Trailer": true, "Upgrade": true,
}

var methods = map[string]bool{
	http.MethodGet: true, http.MethodHead: true, http.MethodPost: true, http.MethodPut: true, http.MethodPatch: true,
	http.MethodDelete: true,
}

// prepare resolves the connection and checks the request during the call of the program. It has no side effects
// besides the resolver, so a request, which the limits refuse afterwards, leaves nothing behind.
func (h *Host) prepare(b Binding, name string, req net.Request) (res prepared) {
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	if methods[req.Method] {
		res.method = req.Method
	}
	defer func() {
		// the counter of the open requests already contains this one
		if v := recover(); v != nil {
			if _, aborted := gostrict.AbortedErr(v); aborted {
				panic(v)
			}
			slog.Error("the connection of a program can not be resolved", "connection", name, "panic", v)
			res.refuse(reasonConnection)
		}
	}()
	conn, err := b.Resolve(name)
	if err != nil {
		var refused Refused
		if !errors.As(err, &refused) {
			slog.Warn("a program may not use a connection", "connection", name, "err", err)
		}
		res.refuse(reasonOf(err, reasonConnection))
		return res
	}
	res.conn = conn
	if conn.BaseURL == nil || conn.BaseURL.Host == "" || (conn.BaseURL.Scheme != "http" && conn.BaseURL.Scheme != "https") {
		slog.Error("a connection has no valid address", "connection", name)
		res.refuse(reasonConnection)
		return res
	}
	if !methods[req.Method] {
		res.refuse("the method is not allowed")
		return res
	}
	if reason := checkPath(req.Path); reason != "" {
		res.refuse(reason)
		return res
	}
	if conn.Match != nil {
		route, ok := conn.Match(req.Method, req.Path)
		if !ok {
			res.refuse(reasonNotDeclared)
			return res
		}
		res.route = route
	}
	if header, ok := forbiddenHeader(req.Header); ok {
		res.refuse("the header " + header + " is not allowed")
		return res
	}
	if reason := checkContent(req); reason != "" {
		res.refuse(reason)
		return res
	}

	// the fields are set directly, so that the path is never parsed as an address and can not change the host
	u := *conn.BaseURL
	u.Path = strings.TrimSuffix(u.Path, "/") + req.Path
	u.RawPath = ""
	u.Fragment = ""
	u.User = nil
	query := u.Query()
	for k, v := range req.Query.All() {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	timeout := h.opts.DefaultTimeout
	if req.Timeout > 0 {
		timeout = min(time.Duration(req.Timeout)*time.Millisecond, h.opts.MaxTimeout)
	}
	res.req, res.target, res.timeout = req, &u, timeout
	res.streaming = isTransfer(req)
	return res
}

// isTransfer tells, whether the request has a source or a sink.
func isTransfer(req net.Request) bool {
	if req.Source != "" || req.Sink != "" {
		return true
	}
	for p := range req.Parts.Values() {
		if p.Source != "" {
			return true
		}
	}
	return false
}

// checkContent returns, why the content of the request is contradictory, or "".
func checkContent(req net.Request) string {
	if req.Body != "" && req.Source != "" {
		return reasonSourceBody
	}
	if req.Parts.Len() > 0 && (req.Body != "" || req.Source != "") {
		return reasonPartsBody
	}
	for p := range req.Parts.Values() {
		if p.Name == "" || (p.Value != "" && p.Source != "") || !validMime(p.Mime) {
			return reasonParts
		}
	}
	return ""
}

// validMime tells, whether a type can be a value of a header. It is empty, if there is none.
func validMime(mime string) bool {
	return httpguts.ValidHeaderFieldValue(mime)
}

// attach finds the content and the target of a transfer. It comes after all limits, because a sink may create a file.
func (h *Host) attach(b Binding, r *prepared) {
	defer func() {
		if v := recover(); v != nil {
			if _, aborted := gostrict.AbortedErr(v); aborted {
				panic(v)
			}
			slog.Error("the streams of a request have failed", "panic", v)
			r.refuse(reasonStream)
		}
	}()
	if r.streaming && b.Streams == nil {
		r.refuse(reasonNoStreams)
		return
	}
	switch {
	case r.req.Parts.Len() > 0:
		sources := make([]Source, r.req.Parts.Len())
		for i, p := range r.req.Parts.All() {
			if p.Source == "" {
				continue
			}
			src, err := b.Streams.Source(p.Source)
			if err != nil {
				refuseStream(r, err)
				return
			}
			sources[i] = src
		}
		body, ok := newMultipart(r.req.Parts, sources)
		if !ok {
			r.refuse(reasonParts)
			return
		}
		r.upload = body.upload()
	case r.req.Source != "":
		src, err := b.Streams.Source(r.req.Source)
		if err != nil {
			refuseStream(r, err)
			return
		}
		if !validMime(src.Mime) {
			r.refuse(reasonStream)
			return
		}
		r.upload = &upload{open: src.Open, size: src.Size, contentType: src.Mime}
	}
	// the sink comes last, so that nothing refuses the request after it exists
	if r.req.Sink != "" {
		sink, err := b.Streams.Sink(r.req.Sink)
		if err != nil {
			refuseStream(r, err)
			return
		}
		r.sink = sink
	}
}

func refuseStream(r *prepared, err error) {
	var refused Refused
	if !errors.As(err, &refused) {
		slog.Warn("a request can not use a stream", "err", err)
	}
	r.refuse(reasonOf(err, reasonStream))
}

// checkPath returns, why the path is not valid, or "".
func checkPath(p string) string {
	if !strings.HasPrefix(p, "/") {
		return "the path must begin with /"
	}
	if strings.Contains(p, "//") || strings.Contains(p, `\`) {
		return "the path must not contain // or a backslash"
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." || seg == "." {
			return "the path must not contain . or .."
		}
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return "the path must not contain control characters"
		}
	}
	return ""
}

func forbiddenHeader(m dict.Map[string, string]) (string, bool) {
	for k, v := range m.All() {
		name := http.CanonicalHeaderKey(k)
		if forbiddenHeaders[name] || strings.HasPrefix(name, "Proxy-") || strings.HasPrefix(name, "X-Forwarded-") {
			return name, true
		}
		// the transport would refuse them after the connection, and then it is no refused header anymore
		if !httpguts.ValidHeaderFieldName(k) || !httpguts.ValidHeaderFieldValue(v) {
			return name, true
		}
	}
	return "", false
}

// ending is what a request leaves behind for its observation and its function.
type ending struct {
	result   net.Result
	class    Class
	status   int32
	end      time.Time
	duration time.Duration
}

// exchange is the state of a request on the network.
type exchange struct {
	host  *Host
	r     *prepared
	ctx   context.Context
	dog   *watchdog // nil for a request, which is no transfer
	begin time.Time
	sent  atomic.Bool
	// fault is the first error of reading the content of the request, which the transport has no type for.
	fault atomic.Pointer[error]
	// body is the content, which was opened last. The transport waits for its write loop, which waits for a read of the
	// content, so only closing it here ends a request, whose source hangs.
	body atomic.Pointer[uploadBody]
}

func (x *exchange) finishAt(last time.Time, res net.Result, class Class) ending {
	e := ending{result: res, class: class, end: last, duration: last.Sub(x.begin)}
	if class == Answered {
		e.status = res.Status
	}
	return e
}

func (x *exchange) finish(res net.Result, class Class) ending {
	return x.finishAt(time.Now(), res, class)
}

// do sends the request and classifies its end.
func (h *Host) do(r *prepared) (e ending) {
	x := &exchange{host: h, r: r, begin: time.Now()}
	defer r.discardSink()
	defer x.closeBody()
	// this goroutine has no caller, which could recover a panic of a decorator or a source
	defer func() {
		if v := recover(); v != nil {
			slog.Error("a request of a program has failed", "panic", v)
			class := Network
			if !x.sent.Load() {
				class = NotSent
			}
			e = x.finish(net.Result{Outcome: net.Unknown, Reason: "the request has failed"}, class)
		}
	}()

	ctx, cancel := context.WithCancelCause(h.ctx)
	defer cancel(nil)
	if r.conn.Context != nil {
		stop := context.AfterFunc(r.conn.Context, func() { cancel(context.Canceled) })
		defer stop()
	}
	limit := r.timeout
	if r.streaming {
		limit = h.opts.MaxStream
		x.dog = startWatchdog(h.opts.StreamStall, func() { cancel(errStalled) })
		defer x.dog.stop()
	}
	ctx, stopLimit := context.WithTimeout(ctx, limit)
	defer stopLimit()
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{WroteHeaders: func() { x.sent.Store(true) }})
	x.ctx = ctx

	var body io.Reader
	if r.req.Body != "" {
		body = strings.NewReader(r.req.Body)
	}
	req, err := http.NewRequestWithContext(ctx, r.req.Method, r.target.String(), body)
	if err != nil {
		return x.finish(net.Result{Outcome: net.Failed, Reason: reasonInvalid}, NotSent)
	}
	for k, v := range r.req.Header.All() {
		req.Header.Set(k, v)
	}
	if r.upload != nil {
		if err := x.attachUpload(req); err != nil {
			return x.finish(net.Result{Outcome: net.Failed, Reason: reasonOf(err, reasonSource)}, NotSent)
		}
		defer context.AfterFunc(ctx, x.closeBody)()
	}

	transport := h.base
	if r.conn.PrivateOnly {
		transport = h.internal
	}
	if r.conn.Decorate != nil {
		transport = r.conn.Decorate(transport)
	}
	client := http.Client{Transport: transport, CheckRedirect: func(next *http.Request, via []*http.Request) error {
		if len(via) >= 5 || next.URL.Scheme != r.target.Scheme || next.URL.Host != r.target.Host {
			return http.ErrUseLastResponse
		}
		return nil
	}}

	resp, err := client.Do(req)
	if err != nil {
		return x.failed(err)
	}
	defer resp.Body.Close()
	return x.answer(resp)
}

// attachUpload sets the content of a transfer on the request.
func (x *exchange) attachUpload(req *http.Request) error {
	u := x.r.upload
	if u.contentType != "" && (u.replaceType || req.Header.Get("Content-Type") == "") {
		req.Header.Set("Content-Type", u.contentType)
	}
	if u.size == 0 {
		req.Body, req.ContentLength = http.NoBody, 0
		req.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		return nil
	}
	body, err := x.openUpload()
	if err != nil {
		return err
	}
	req.Body, req.GetBody, req.ContentLength = body, x.openUpload, u.size
	return nil
}

func (x *exchange) openUpload() (io.ReadCloser, error) {
	rc, err := x.r.upload.open()
	if err != nil {
		x.fault.Store(&err)
		return nil, err
	}
	body := &uploadBody{rc: rc, x: x}
	x.body.Store(body)
	if x.ctx.Err() != nil {
		_ = body.Close()
	}
	return body, nil
}

func (x *exchange) closeBody() {
	if body := x.body.Load(); body != nil {
		_ = body.Close()
	}
}

// failed classifies an error without an answer.
func (x *exchange) failed(err error) ending {
	outcome, reason := net.Unknown, x.lostReason()
	if !x.sent.Load() {
		outcome, reason = net.Failed, reasonUnreachable
	}
	class := Network
	if errors.Is(err, errForbiddenAddress) {
		class = NotSent
	}
	if fault := x.fault.Load(); fault != nil {
		class, reason = NotSent, reasonOf(*fault, reasonSource)
	}
	return x.finish(net.Result{Outcome: outcome, Reason: reason}, class)
}

// lostReason tells, why the connection has ended without an error of the server.
func (x *exchange) lostReason() string {
	cause := context.Cause(x.ctx)
	switch {
	case errors.Is(cause, errStalled):
		return reasonStalled
	case errors.Is(cause, context.DeadlineExceeded):
		return reasonTimeout
	}
	return reasonAborted
}

func (x *exchange) answer(resp *http.Response) ending {
	res := net.Result{Status: int32(resp.StatusCode), Header: answerHeader(resp.Header)}
	var body io.Reader = resp.Body
	if x.dog != nil {
		body = progressReader{r: body, dog: x.dog}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return x.rejected(res, body)
	}
	if x.r.req.Sink != "" {
		return x.store(res, resp, body)
	}

	text, err := io.ReadAll(io.LimitReader(body, x.host.opts.MaxBody+1))
	if err != nil {
		res.Outcome, res.Reason = net.Unknown, x.lostReason()
		return x.finish(res, Network)
	}
	if int64(len(text)) > x.host.opts.MaxBody {
		res.Outcome, res.Reason = net.Unknown, reasonTooLarge
		return x.finish(res, Answered)
	}
	res.Outcome, res.Body = net.OK, strings.ToValidUTF8(string(text), "�")
	return x.finish(res, Answered)
}

// rejected reads a part of the text of an error, which helps to understand it.
func (x *exchange) rejected(res net.Result, body io.Reader) ending {
	text, err := io.ReadAll(io.LimitReader(body, maxReason))
	res.Outcome, res.Reason = net.Rejected, fmt.Sprintf("HTTP %d: %s", res.Status, strings.ToValidUTF8(string(text), "�"))
	if err != nil {
		return x.finish(res, Network)
	}
	return x.finish(res, Answered)
}

// store copies the answer to the sink.
func (x *exchange) store(res net.Result, resp *http.Response, body io.Reader) ending {
	w := &sinkWriter{sink: x.r.sink}
	_, err := io.CopyBuffer(w, body, make([]byte, 32<<10))
	last := time.Now()
	if w.err != nil {
		res.Outcome, res.Reason = net.Unknown, reasonOf(w.err, reasonStore)
		return x.finishAt(last, res, Answered)
	}
	if err != nil || x.ctx.Err() != nil {
		res.Outcome, res.Reason = net.Unknown, x.lostReason()
		return x.finishAt(last, res, Network)
	}
	stored, err := x.r.sink.Commit(resp.Header.Get("Content-Type"))
	if err != nil {
		res.Outcome, res.Reason = net.Unknown, reasonStore
		return x.finishAt(last, res, Answered)
	}
	x.r.settled = true
	res.Outcome, res.Stored = net.OK, stored
	return x.finishAt(last, res, Answered)
}

// answerHeader returns the headers of an answer without its cookies, which are credentials, too.
func answerHeader(h http.Header) dict.Map[string, string] {
	res := dict.New[string, string]()
	for k, vs := range h {
		if k != "Set-Cookie" {
			res.Set(k, strings.Join(vs, ", "))
		}
	}
	return res
}

// scrubResult removes the secrets from the texts of the result.
func scrubResult(r net.Result, secrets []string) net.Result {
	if len(secrets) == 0 {
		return r
	}
	r.Reason, r.Body, r.Stored = scrub(r.Reason, secrets), scrub(r.Body, secrets), scrub(r.Stored, secrets)
	for k, v := range r.Header.All() {
		r.Header.Set(k, scrub(v, secrets))
	}
	return r
}

func scrub(s string, secrets []string) string {
	for _, secret := range secrets {
		if secret != "" {
			s = strings.ReplaceAll(s, secret, "***")
		}
	}
	return s
}
