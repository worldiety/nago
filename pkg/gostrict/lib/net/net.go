// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package net sends HTTP requests of a program on a server. A program never sees an address or a credential: it names
// a connection, which an administrator has set up, and a path of it. The host adds the credentials outside of the
// program, so a defect or an injected instruction can not send them elsewhere. A connection to a public API simply has
// no credentials. A connection may offer further values, e.g. the number of a tenant, which the program builds into a
// request, see [Value].
//
// Send returns at once, and the function of its last parameter runs later as a call of its own with its own
// transaction, e.g. to save the result. It runs on another goroutine, so it can only use what can not change:
// the compiler rejects anything else (GS046).
//
//	func invoke(in Input) (Output, error) {
//		id := net.Send("crm", net.NewRequest("GET", "/customers/"+in.ID, bin.Empty()), func(r net.Result) error {
//			if r.Outcome != net.OK {
//				return errors.New(r.Reason)
//			}
//			return save(r.Body.Text())
//		})
//		return Output{Job: id}, nil
//	}
//
// A view waits for the end with ui.OnDone(wnd, out.Job, func(failure string) { ... }).
//
// Large content does not enter the program. A host may name it: [Request.Source] sends what a name stands for,
// [Request.Parts] sends a form with such files and [Request.Sink] writes an answer with a status of 2xx there, so that
// the program only sees [Result.Stored]. A request with one of them is a transfer: instead of [Request.Timeout] it ends
// when no byte moved for a while or after a very long time, and only the host knows both limits.
//
//gostrict:header
package net

import (
	"go.wdy.de/nago/pkg/gostrict/std/bin"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
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
	// Body is the content of the request, which may be binary. Use bin.FromText for a text and json.Encode(v) for
	// JSON.
	Body bin.Bytes
	// Timeout is the limit in milliseconds, 30 seconds if 0, and the host limits it. A transfer ignores it.
	Timeout int32
	// Source names the content, which the host sends instead of Body. The name comes from the host, e.g. from a
	// function of a file store. A Body and a Source together are [Failed].
	Source string
	// Sink names, where the host writes the answer of a status of 2xx. [Result.Body] stays empty then.
	Sink string
	// Parts is a form in multipart/form-data, whose Content-Type the host sets. If there is a part, Body and Source must
	// be empty, otherwise the request is [Failed].
	Parts list.List[Part]
}

// Part is a field of a form of [Request.Parts].
type Part struct {
	// Name is the name of the field and must not be empty.
	Name string
	// Value is the text of the field. A Value and a Source together are [Failed].
	Value string
	// Source names the content of the field, like [Request.Source].
	Source string
	// FileName is the name of the file, or empty for the name, which the host knows for Source.
	FileName string
	// Mime is the type of the content, or empty for the type, which the host knows for Source.
	Mime string
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
	// Body is the content of the answer, which may be binary, but only with [OK], so that a forgotten check never parses
	// an error page. Body.Text() returns it as text, json.Decode[T](r.Body) reads JSON into T and a program reads a
	// binary format with bin.Decode(r.Body, ...).
	Body bin.Bytes
	// Stored is what the host reports about the answer, which it wrote to [Request.Sink], but only with [OK].
	Stored string
}

// NewRequest returns a request with empty query parameters, headers and parts, which the program may set.
//
//gostrict:host
func NewRequest(method, path string, body bin.Bytes) Request {
	return Request{Method: method, Path: path, Query: dict.New[string, string](), Header: dict.New[string, string](), Body: body, Parts: list.New[Part]()}
}

// NewTextRequest returns a request like [NewRequest], whose content is the UTF-8 of the text.
//
//gostrict:host
func NewTextRequest(method, path, text string) Request {
	return NewRequest(method, path, bin.FromText(text))
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

// Value returns the value of the key of the connection and true, or "" and false, if the connection does not exist or
// does not offer this key: a program never gets a value, which its administrator has marked as write-only, and none of
// the keys, which the host itself applies to each request, e.g. the address and the headers of the login. So the value
// may be built into a header, a parameter or the content of a request, but it is no secret: the program can show it.
//
//gostrict:host
//gostrict:const connection
//gostrict:zero
func Value(connection, key string) (string, bool) {
	panic("net.Value needs a host, see package nethost")
}

// Values returns all keys and values of the connection, which [Value] finds, and is empty for an unknown connection.
//
//gostrict:host
//gostrict:const connection
func Values(connection string) dict.Map[string, string] {
	panic("net.Values needs a host, see package nethost")
}
