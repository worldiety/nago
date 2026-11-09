// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package nethost implements package lib/net for a host in Go. The credentials of a connection stay here: they never
// enter an instance, so neither a program nor its logs or results can see them.
package nethost

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
)

// SendFunc is the full name of the function net.Send.
const SendFunc = "go.wdy.de/nago/pkg/gostrict/lib/net.Send"

// Connection is what the name of a connection stands for. It never reaches the program.
type Connection struct {
	// BaseURL is the address of the connection, which the path of each request extends. A request and its redirects
	// never leave its origin.
	BaseURL *url.URL
	// Decorate wraps the safe transport of the [Host], e.g. to set the credentials, to renew a token or to write a
	// protocol. It sees each request after the program, so the program can not overwrite what it sets. Nil adds nothing.
	Decorate func(next http.RoundTripper) http.RoundTripper
	// Secrets are the values, which must never reach the program, also if the server mirrors them.
	Secrets []string
}

// Resolver returns the connection with the exact name or an error, if the running program may not use it. It runs
// synchronously during the call of net.Send, thus in the frame of the program. Its error never reaches the program.
type Resolver func(name string) (Connection, error)

// Completer runs the function done with the arguments as a call of its own, which owns its frame and its transaction,
// and returns the error of the function or of the commit. A nil error means, that its work is committed. The context
// ends with the application and after [Options.CompleteTimeout].
type Completer func(ctx context.Context, done gostrict.Detached, args []any) error

// PoolCompleter runs the function in a free instance of the pool, without a frame or a transaction of its own.
func PoolCompleter(p *gostrict.Pool) Completer {
	return func(ctx context.Context, done gostrict.Detached, args []any) error {
		res, err := p.CallDetached(ctx, done, args...)
		if err != nil {
			return err
		}
		return gostrict.ErrorOf(res[0])
	}
}

// Options are the limits of a [Host]. A zero value means its default.
type Options struct {
	// MaxOpen limits the open requests of all programs, 256 by default.
	MaxOpen int
	// MaxOpenPerWindow limits the open requests of a window, 16 by default.
	MaxOpenPerWindow int
	// MaxBody limits the size of an answer in bytes, 16 MiB by default.
	MaxBody int64
	// DefaultTimeout is the limit of a request without one, 30 seconds by default.
	DefaultTimeout time.Duration
	// MaxTimeout is the largest limit of a request, 60 seconds by default.
	MaxTimeout time.Duration
	// CompleteTimeout limits the function, which handles the end of a request, 30 seconds by default.
	CompleteTimeout time.Duration
	// AllowPrivateNetworks allows the addresses of the local machine and of private networks, e.g. for a connection
	// to a server in the same network. Link-local addresses and those of cloud metadata services stay forbidden.
	AllowPrivateNetworks bool
}

func (o Options) withDefaults() Options {
	set := func(v *time.Duration, d time.Duration) {
		if *v <= 0 {
			*v = d
		}
	}
	if o.MaxOpen <= 0 {
		o.MaxOpen = 256
	}
	if o.MaxOpenPerWindow <= 0 {
		o.MaxOpenPerWindow = 16
	}
	if o.MaxBody <= 0 {
		o.MaxBody = 16 << 20
	}
	set(&o.DefaultTimeout, 30*time.Second)
	set(&o.MaxTimeout, 60*time.Second)
	set(&o.CompleteTimeout, 30*time.Second)
	return o
}

// Binding connects the functions of an instance with its host, e.g. the slot of a pool.
type Binding struct {
	// Resolve finds the connections of the instance.
	Resolve Resolver
	// Complete runs the end of a request.
	Complete Completer
	// Jobs returns the jobs of the window, which runs the instance now, so that the window can wait for the end. Nil or
	// a nil result means, that no window is involved.
	Jobs func() *taskhost.Jobs
}

// Host sends the requests of the programs of an application.
type Host struct {
	ctx  context.Context
	opts Options
	base http.RoundTripper
	// open counts the requests and the ends of refused requests, which are not delivered yet
	open atomic.Int64
}

// New returns a host. A request ends with the context, which belongs to the application, never to a window, because
// closing a window does not stop its work.
func New(ctx context.Context, opts Options) *Host {
	opts = opts.withDefaults()
	transport := newTransport(opts.AllowPrivateNetworks)
	context.AfterFunc(ctx, transport.CloseIdleConnections)
	return &Host{ctx: ctx, opts: opts, base: transport}
}

// Funcs returns the functions of package lib/net for an instance.
func (h *Host) Funcs(b Binding) gostrict.Funcs {
	if b.Resolve == nil || b.Complete == nil {
		panic("nethost: a binding needs Resolve and Complete")
	}
	funcs := net.Host()
	funcs[SendFunc] = func(args []any) []any { return h.send(b, args) }
	return funcs
}

// send never starts more than [Options.MaxOpen] requests. A refused request ends like any other, but its end needs a
// goroutine, too, so a program, which refuses to stop, is aborted at twice the limit.
func (h *Host) send(b Binding, args []any) []any {
	name := args[0].(string)
	var req net.Request
	if err := gostrict.FromNeutral(args[1], &req); err != nil {
		gostrict.Abort(fmt.Errorf("net.Send: %w", err))
	}
	done, err := gostrict.Detach(args[2].(gostrict.Callback))
	if err != nil {
		gostrict.Abort(fmt.Errorf("net.Send: %w", err))
	}
	if int(h.open.Add(1)) > 2*h.opts.MaxOpen {
		h.open.Add(-1)
		gostrict.Abort(errors.New("net.Send: too many open requests"))
	}

	var jobs *taskhost.Jobs
	if b.Jobs != nil {
		jobs = b.Jobs()
	}
	var id task.ID
	started := true
	if jobs != nil {
		id, started = jobs.Start(h.opts.MaxOpenPerWindow)
	}

	switch {
	case !started:
		go h.deliver(b, done, nil, "", net.Result{Outcome: net.Failed, Reason: reasonTooMany}, nil)
	case int(h.open.Load()) > h.opts.MaxOpen:
		go h.deliver(b, done, jobs, id, net.Result{Outcome: net.Failed, Reason: reasonTooMany}, nil)
	default:
		r, ok := h.prepare(b, name, req)
		if !ok {
			go h.deliver(b, done, jobs, id, r.failed, nil)
			break
		}
		go func() {
			h.deliver(b, done, jobs, id, h.protect(r), r.conn.Secrets)
		}()
	}
	return []any{string(id)}
}

// protect sends the request and turns a panic of a decorator into the end of the request, because this goroutine has
// no caller, which could recover it.
func (h *Host) protect(r prepared) (res net.Result) {
	defer func() {
		if v := recover(); v != nil {
			slog.Error("a request of a program has failed", "panic", v)
			res = net.Result{Outcome: net.Unknown, Reason: "the request has failed"}
		}
	}()
	return h.do(r)
}

// deliver runs the end of a request and reports it to the window.
func (h *Host) deliver(b Binding, done gostrict.Detached, jobs *taskhost.Jobs, id task.ID, res net.Result, secrets []string) {
	defer h.open.Add(-1)
	failure := h.complete(b, done, scrubResult(res, secrets))
	if jobs != nil {
		jobs.Finish(id, scrub(failure, secrets))
	}
}

func (h *Host) complete(b Binding, done gostrict.Detached, res net.Result) (failure string) {
	defer func() {
		if v := recover(); v != nil {
			slog.Error("the end of a request of a program has failed", "panic", v)
			failure = "the job has failed"
		}
	}()
	arg, err := gostrict.ToNeutral(res)
	if err != nil {
		return err.Error()
	}
	ctx, cancel := context.WithTimeout(h.ctx, h.opts.CompleteTimeout)
	defer cancel()
	if err := b.Complete(ctx, done, []any{arg}); err != nil {
		return err.Error()
	}
	return ""
}
