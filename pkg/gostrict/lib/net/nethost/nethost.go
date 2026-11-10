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
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/netip"
	"net/url"
	"runtime"
	"slices"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
)

// SendFunc is the full name of the function net.Send.
const SendFunc = "go.wdy.de/nago/pkg/gostrict/lib/net.Send"

// The full names of the functions, which read the values of a connection.
const (
	ValueFunc  = "go.wdy.de/nago/pkg/gostrict/lib/net.Value"
	ValuesFunc = "go.wdy.de/nago/pkg/gostrict/lib/net.Values"
)

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
	// Match finds the declared endpoint of a request. It gets the method, which is GET for an empty one, and the plain
	// path before its encoding, and returns the name of the endpoint for [Observation.Route] or false, if the request
	// must not be sent. Nil allows every request without a name. It runs before a connection exists and before Decorate.
	Match func(method, path string) (route string, ok bool)
	// PrivateOnly allows this connection only the addresses of the local machine and of private networks, e.g. to send
	// credentials without TLS. It is checked after the name was resolved and does not widen
	// [Options.AllowPrivateNetworks]: without that option such a connection reaches nothing.
	PrivateOnly bool
	// Values are the keys and values, which the program may read with net.Value. They must not contain a secret.
	Values map[string]string
	// Context stops all requests of the connection, when it ends, e.g. when the data, which they belong to, is removed.
	// Such a request ends without its function done and without an [Observation], and its job fails. Nil means, that only
	// the context of the [Host] stops the request.
	Context context.Context
}

// Refused is an error of a [Resolver], of [Streams] or of a [Sink], whose text the program may read, so it must not
// contain a secret. It says why, e.g. that a file is too large. Any other error stays in the logs of the host.
type Refused string

func (r Refused) Error() string { return string(r) }

// Resolver returns the connection with the exact name or an error, if the running program may not use it. It runs
// synchronously during the call of net.Send, thus in the frame of the program. The program only reads an error of type
// [Refused].
type Resolver func(name string) (Connection, error)

// Streams finds what the names of [net.Request.Source], [net.Request.Sink] and [net.Part.Source] stand for. Both run
// synchronously during the call of net.Send, thus in the frame of the program, so they may check the rights of the
// caller. The program only reads an error of type [Refused].
type Streams interface {
	Source(name string) (Source, error)
	Sink(name string) (Sink, error)
}

// Source is the content, which a request sends.
type Source struct {
	// Open returns the content from its beginning. It runs on another goroutine than the program and may run several
	// times, e.g. for a redirect, which needs the content again. Close must end a Read, which is blocked, because the
	// limits of a transfer can only end the request this way.
	Open func() (io.ReadCloser, error)
	// Size is the exact size of the content in bytes, or -1 if it is unknown. Zero is an empty content.
	Size int64
	// Mime is the type of the content or empty.
	Mime string
	// Name is the name of the file for a part of a form or empty.
	Name string
}

// Sink receives the answer of a request with a status of 2xx. The host calls Commit after the last byte or Abort, and
// Abort also after a failed Commit, so that no half written content stays behind. Write may return a [Refused] to tell
// why the answer can not be kept, e.g. that it is too large.
type Sink interface {
	io.Writer
	// Commit keeps the content and returns the text of [net.Result.Stored]. Its error ends the request as
	// [net.Unknown], because the server has already done its work.
	Commit(contentType string) (stored string, err error)
	// Abort discards the content.
	Abort()
}

// Class tells what an [Observation] has seen. It is never zero.
type Class int

const (
	// Answered means, that the server has answered. An answer, which the host could not keep, is answered as well.
	Answered Class = iota + 1
	// Network means, that no answer has arrived completely: the server can not be reached, the time ended or the
	// connection was lost, also after the status.
	Network
	// NotSent means, that the request ended before it was sent, or that its own content could not be read.
	NotSent
)

// Observation describes the end of a request. It contains no value of the path, no header and no content.
type Observation struct {
	// Connection is the name of the connection.
	Connection string
	// Method is the method after its default or empty, if it is not one of the allowed methods.
	Method string
	// Route is the result of [Connection.Match], or empty without one.
	Route string
	// Class is the kind of the end.
	Class Class
	// Status is the status of the answer with [Answered] and otherwise 0.
	Status int32
	// End is the time of the last read or written byte or of the abort. Duration is the time from the begin of the
	// sending until End, over all redirects, and zero, if the request was refused before.
	End      time.Time
	Duration time.Duration
}

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
	// MaxBody limits the size of an answer, which reaches the program as text, in bytes, 16 MiB by default. Only a [Sink]
	// limits what it receives.
	MaxBody int64
	// DefaultTimeout is the limit of a request without one, 30 seconds by default.
	DefaultTimeout time.Duration
	// MaxTimeout is the largest limit of a request, 60 seconds by default.
	MaxTimeout time.Duration
	// CompleteTimeout limits the function, which handles the end of a request, 30 seconds by default.
	CompleteTimeout time.Duration
	// StreamStall ends a transfer, which has moved no byte in any direction for this time, 5 minutes by default. A
	// transfer is a request with a source, a sink or a source in its parts, and only other requests have a timeout.
	StreamStall time.Duration
	// MaxStream limits the whole time of a transfer, 24 hours by default.
	MaxStream time.Duration
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
	set(&o.StreamStall, 5*time.Minute)
	set(&o.MaxStream, 24*time.Hour)
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
	// Streams finds the content and the targets of a transfer. Nil refuses each request, which names one of them.
	Streams Streams
	// Defer postpones a request, e.g. until the transaction of the caller is committed. The host calls it during
	// net.Send exactly once, also for a request which is already refused. The binding must call exactly one of its
	// functions: start sends the request or delivers its refusal and returns at once, and drop ends it without a request,
	// without its function done and without an [Observation], frees its place among the open requests, aborts its sink
	// and lets its job fail. Calling both or calling one twice is a defect of the binding, only the first call has an
	// effect, and if a binding forgets both, the garbage collector finds it and the request ends like with drop. Nil
	// starts a request at once.
	Defer func(start, drop func())
	// Observe gets exactly one [Observation] for each request before its function done runs, unless the request was
	// dropped, ended with [Connection.Context] or by the end of the context of the [Host].
	Observe func(Observation)
}

// Host sends the requests of the programs of an application.
type Host struct {
	ctx  context.Context
	opts Options
	// base connects by the rules of the options and internal only to local and private addresses. Both are separate
	// transports, because a connection in a pool was checked once by the rule of its creator and would otherwise serve
	// a request of the other rule.
	base     http.RoundTripper
	internal http.RoundTripper
	// open counts the requests and the ends of refused requests, which are not delivered yet
	open atomic.Int64
}

// New returns a host. A request ends with the context, which belongs to the application, never to a window, because
// closing a window does not stop its work.
func New(ctx context.Context, opts Options) *Host {
	opts = opts.withDefaults()
	base := newTransport(func(ip netip.Addr) bool { return forbiddenAddress(ip, opts.AllowPrivateNetworks) })
	internal := newTransport(func(ip netip.Addr) bool { return forbiddenPrivateOnly(ip, opts.AllowPrivateNetworks) })
	context.AfterFunc(ctx, func() {
		base.CloseIdleConnections()
		internal.CloseIdleConnections()
	})
	return &Host{ctx: ctx, opts: opts, base: base, internal: internal}
}

// Funcs returns the functions of package lib/net for an instance.
func (h *Host) Funcs(b Binding) gostrict.Funcs {
	if b.Resolve == nil || b.Complete == nil {
		panic("nethost: a binding needs Resolve and Complete")
	}
	funcs := net.Host()
	funcs[SendFunc] = func(args []any) []any { return h.send(b, args) }
	funcs[ValueFunc] = func(args []any) []any {
		v, ok := h.values(b, args[0].(string))[args[1].(string)]
		return []any{v, ok}
	}
	funcs[ValuesFunc] = func(args []any) []any {
		values := h.values(b, args[0].(string))
		keys := slices.Sorted(maps.Keys(values))
		entries := make([]any, len(keys))
		for i, k := range keys {
			entries[i] = []any{k, values[k]}
		}
		return []any{entries}
	}
	return funcs
}

// values returns the values of the connection, which the program may read, and nil, if it can not be resolved.
func (h *Host) values(b Binding, name string) map[string]string {
	conn, err := b.Resolve(name)
	if err != nil {
		var refused Refused
		if !errors.As(err, &refused) {
			slog.Error("cannot resolve the values of a connection", "connection", name, "err", err)
		}
		return nil
	}
	return conn.Values
}

// call is a single net.Send from its acceptance until its end.
type call struct {
	host    *Host
	binding Binding
	name    string
	done    gostrict.Detached
	jobs    *taskhost.Jobs
	id      task.ID
	plan    prepared
	// claimed is set by the first of start, drop and the end of the call of net.Send.
	claimed atomic.Bool
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

	c := &call{host: h, binding: b, name: name, done: done}
	defer func() {
		// the program is aborted, so its request is the same as dropped
		if v := recover(); v != nil {
			if c.claimed.CompareAndSwap(false, true) {
				c.abandon(reasonDropped)
			}
			panic(v)
		}
	}()

	started := true
	if b.Jobs != nil {
		c.jobs = b.Jobs()
	}
	if c.jobs != nil {
		c.id, started = c.jobs.Start(h.opts.MaxOpenPerWindow)
		if !started {
			c.jobs = nil
		}
	}

	// the limits come after the connection, so that a refused request is still counted at its endpoint
	c.plan = h.prepare(b, name, req)
	switch {
	case c.plan.refused:
	case !started, int(h.open.Load()) > h.opts.MaxOpen:
		c.plan.refuse(reasonTooMany)
	default:
		h.attach(b, &c.plan)
	}

	if b.Defer == nil {
		c.start()
	} else {
		c.deferTo(b.Defer)
	}
	return []any{string(c.id)}
}

// deferToken is only reachable through the functions of a deferred call, so the garbage collector finds a binding,
// which has lost them. It holds a pointer, because a cleanup is unreliable for the tiny allocations of pointer free
// objects.
type deferToken struct {
	call *call
}

func (c *call) deferTo(fn func(start, drop func())) {
	token := &deferToken{call: c}
	runtime.AddCleanup(token, (*call).forgotten, c)
	fn(func() { token.call.start() }, func() { token.call.drop() })
}

func (c *call) forgotten() {
	if c.claimed.CompareAndSwap(false, true) {
		slog.Error("a deferred request was neither started nor dropped", "connection", c.name)
		c.abandon(reasonDropped)
	}
}

// claim tells, whether this is the first decision about the deferred call. A second one is a defect of the binding,
// which must not end a request twice.
func (c *call) claim(action string) bool {
	if c.claimed.CompareAndSwap(false, true) {
		return true
	}
	slog.Error("a deferred request was decided twice", "connection", c.name, "action", action)
	return false
}

// start runs the request on a goroutine of its own, because the caller may hold a transaction or a window.
func (c *call) start() {
	if c.claim("start") {
		go c.run()
	}
}

func (c *call) drop() {
	if c.claim("drop") {
		c.abandon(reasonDropped)
	}
}

// abandon ends the request without its function done and without an observation.
func (c *call) abandon(failure string) {
	c.plan.discardSink()
	c.finish(failure)
}

// finish frees the place of the request before its window hears about the end, so that a window, which reacts with the
// next request, never finds its own place still taken.
func (c *call) finish(failure string) {
	c.host.open.Add(-1)
	if c.jobs != nil {
		c.jobs.Finish(c.id, failure)
	}
}

// run sends the request or delivers its refusal and runs the end of the request.
func (c *call) run() {
	e := c.exchange()
	// also after a complete answer, because what done would write belongs to something, which is gone now
	if c.plan.cancelled() {
		c.abandon(reasonCancelled)
		return
	}
	// a request, which the end of the server has cut off, says nothing about the network
	if c.host.ctx.Err() == nil {
		c.observe(e)
	}
	secrets := c.plan.conn.Secrets
	failure := c.host.complete(c.binding, c.done, scrubResult(e.result, secrets))
	c.finish(scrub(failure, secrets))
}

func (c *call) exchange() ending {
	switch {
	case c.plan.refused:
		return ending{result: c.plan.failed, class: NotSent, end: time.Now()}
	case c.plan.cancelled():
		return ending{}
	}
	return c.host.do(&c.plan)
}

func (c *call) observe(e ending) {
	if c.binding.Observe == nil {
		return
	}
	defer func() {
		if v := recover(); v != nil {
			slog.Error("the observation of a request has failed", "connection", c.name, "panic", v)
		}
	}()
	c.binding.Observe(Observation{
		Connection: c.name, Method: c.plan.method, Route: c.plan.route,
		Class: e.class, Status: e.status, End: e.end, Duration: e.duration,
	})
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
