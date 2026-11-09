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

	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
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
)

const maxReason = 1024

// prepared is a checked request, or the result of a request, which was refused.
type prepared struct {
	conn    Connection
	req     net.Request
	target  *url.URL
	timeout time.Duration
	failed  net.Result
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

// prepare resolves the connection and checks the request during the call of the program.
func (h *Host) prepare(b Binding, name string, req net.Request) (res prepared, ok bool) {
	refuse := func(reason string) (prepared, bool) {
		return prepared{failed: net.Result{Outcome: net.Failed, Reason: reason}}, false
	}
	defer func() {
		// the counter of the open requests already contains this one
		if v := recover(); v != nil {
			slog.Error("the connection of a program can not be resolved", "connection", name, "panic", v)
			res, ok = refuse(reasonConnection)
		}
	}()
	conn, err := b.Resolve(name)
	if err != nil {
		slog.Warn("a program may not use a connection", "connection", name, "err", err)
		return refuse(reasonConnection)
	}
	if conn.BaseURL == nil || conn.BaseURL.Host == "" || (conn.BaseURL.Scheme != "http" && conn.BaseURL.Scheme != "https") {
		slog.Error("a connection has no valid address", "connection", name)
		return refuse(reasonConnection)
	}
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	if !methods[req.Method] {
		return refuse("the method is not allowed")
	}
	if reason := checkPath(req.Path); reason != "" {
		return refuse(reason)
	}
	if header, ok := forbiddenHeader(req.Header); ok {
		return refuse("the header " + header + " is not allowed")
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
	return prepared{conn: conn, req: req, target: &u, timeout: timeout}, true
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
	for k := range m.All() {
		name := http.CanonicalHeaderKey(k)
		if forbiddenHeaders[name] || strings.HasPrefix(name, "Proxy-") || strings.HasPrefix(name, "X-Forwarded-") {
			return name, true
		}
	}
	return "", false
}

// do sends the request and classifies its end.
func (h *Host) do(r prepared) net.Result {
	ctx, cancel := context.WithTimeout(h.ctx, r.timeout)
	defer cancel()

	var sent atomic.Bool
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{WroteHeaders: func() { sent.Store(true) }})
	var body io.Reader
	if r.req.Body != "" {
		body = strings.NewReader(r.req.Body)
	}
	req, err := http.NewRequestWithContext(ctx, r.req.Method, r.target.String(), body)
	if err != nil {
		return net.Result{Outcome: net.Failed, Reason: "the request is not valid"}
	}
	for k, v := range r.req.Header.All() {
		req.Header.Set(k, v)
	}

	var transport http.RoundTripper = h.base
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
		return transportFailure(ctx, err, sent.Load())
	}
	defer resp.Body.Close()
	return h.answer(ctx, resp)
}

// transportFailure classifies an error without an answer.
func transportFailure(ctx context.Context, err error, sent bool) net.Result {
	if !sent {
		return net.Result{Outcome: net.Failed, Reason: reasonUnreachable}
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return net.Result{Outcome: net.Unknown, Reason: reasonTimeout}
	}
	return net.Result{Outcome: net.Unknown, Reason: reasonAborted}
}

func (h *Host) answer(ctx context.Context, resp *http.Response) net.Result {
	res := net.Result{Status: int32(resp.StatusCode), Header: answerHeader(resp.Header)}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// the text of an error helps to understand it, but only a part of it is needed
		text, _ := io.ReadAll(io.LimitReader(resp.Body, maxReason))
		res.Outcome, res.Reason = net.Rejected, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.ToValidUTF8(string(text), "�"))
		return res
	}

	text, err := io.ReadAll(io.LimitReader(resp.Body, h.opts.MaxBody+1))
	if err != nil {
		res.Outcome, res.Reason = net.Unknown, reasonAborted
		if ctx.Err() != nil {
			res.Reason = reasonTimeout
		}
		return res
	}
	if int64(len(text)) > h.opts.MaxBody {
		res.Outcome, res.Reason = net.Unknown, reasonTooLarge
		return res
	}
	res.Outcome, res.Body = net.OK, strings.ToValidUTF8(string(text), "�")
	return res
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
	r.Reason, r.Body = scrub(r.Reason, secrets), scrub(r.Body, secrets)
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
