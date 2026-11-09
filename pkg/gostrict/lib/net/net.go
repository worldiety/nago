// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package net sends HTTP requests of a program on a server. A program never sees an address or a credential: it names
// a connection, which an administrator has set up, and a path of it. The host adds the credentials outside of the
// program, so a defect or an injected instruction can not send them elsewhere. A connection to a public API simply has
// no credentials.
//
// Send returns at once, and the function of its last parameter runs later as a call of its own with its own
// transaction, e.g. to save the result. It runs on another goroutine, so it can only use what can not change:
// the compiler rejects anything else (GS046).
//
//	func invoke(in Input) (Output, error) {
//		id := net.Send("crm", net.NewRequest("GET", "/customers/"+in.ID, ""), func(r net.Result) error {
//			if r.Outcome != net.OK {
//				return errors.New(r.Reason)
//			}
//			return save(r.Body)
//		})
//		return Output{Job: id}, nil
//	}
//
// A view waits for the end with ui.OnDone(wnd, out.Job, func(failure string) { ... }).
//
//gostrict:header
package net

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/task"
)

// Outcome tells, how a request ended.
type Outcome string

const (
	// OK means, that the server answered with a status of 2xx.
	OK Outcome = "ok"
	// Rejected means, that the server answered with another status, see [Result.Status] and [Result.Reason].
	Rejected Outcome = "rejected"
	// Failed means, that the request never reached the server, so it may safely be repeated, e.g. an unknown
	// connection, an invalid path, no connection to the server or too many open requests.
	Failed Outcome = "failed"
	// Unknown means, that the request ended after it was sent, e.g. by the timeout, so the server may have executed it.
	// Do not simply repeat a request, which changes something.
	Unknown Outcome = "unknown"
)

// Request is a request of a connection, see [NewRequest].
type Request struct {
	// Method is the HTTP method: GET if empty, or HEAD, POST, PUT, PATCH or DELETE.
	Method string
	// Path is the path relative to the address of the connection, e.g. "/customers/17". It is a plain text, not an
	// encoded URL, and must begin with / and contain no .. and no //.
	Path string
	// Query are the parameters, which the host encodes.
	Query dict.Map[string, string]
	// Header are additional headers, which must not be credentials: Authorization, Cookie and Host are rejected.
	Header dict.Map[string, string]
	// Body is the text of the request.
	Body string
	// Timeout is the limit in milliseconds, 30 seconds if 0, and the host limits it.
	Timeout int32
}

// Result is the end of a request.
type Result struct {
	Outcome Outcome
	// Reason explains any other outcome than [OK] and is empty otherwise.
	Reason string
	// Status is the HTTP status or 0, if the server did not answer.
	Status int32
	// Header are the headers of the answer.
	Header dict.Map[string, string]
	// Body is the text of the answer, but only with [OK], so that a forgotten check never parses an error page.
	Body string
}

// NewRequest returns a request with empty query parameters and headers, which the program may set.
//
//gostrict:host
func NewRequest(method, path, body string) Request {
	return Request{Method: method, Path: path, Query: dict.New[string, string](), Header: dict.New[string, string](), Body: body}
}

// Send sends the request over the connection and returns at once. The function done runs later as a call of its own,
// exactly once, also if the request fails, and its error rolls back its transaction and is the failure of ui.OnDone.
// The name of the connection is a constant, so that a host knows the connections of a program at compile time.
//
//gostrict:host
//gostrict:const connection
//gostrict:calls concurrent
func Send(connection string, req Request, done func(r Result) error) task.ID {
	panic("net.Send needs a host, see package nethost")
}
