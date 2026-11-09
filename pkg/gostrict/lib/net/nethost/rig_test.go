// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nethost_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
)

// rigProgram offers each way of sending a request and reports all of its result as one text.
const rigProgram = `package p

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

func report(r net.Result) error {
	return errors.New(string(r.Outcome) + "|" + text.FromInt(r.Status) + "|" + r.Reason + "|" + r.Body + "|" + r.Stored)
}

func Send(method, path string) task.ID {
	return net.Send("api", net.NewRequest(method, path, ""), report)
}

func Header(path, name, value string) task.ID {
	r := net.NewRequest("GET", path, "")
	r.Header.Set(name, value)
	return net.Send("api", r, report)
}

func Timed(path string, timeout int32) task.ID {
	r := net.NewRequest("GET", path, "")
	r.Timeout = timeout
	return net.Send("api", r, report)
}

func Transfer(method, path, body, source, sink, contentType string, timeout int32) task.ID {
	r := net.NewRequest(method, path, body)
	r.Source = source
	r.Sink = sink
	r.Timeout = timeout
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	return net.Send("api", r, report)
}

func Download(path, sink string) task.ID {
	return Transfer("GET", path, "", "", sink, "", 0)
}

func Upload(path, source string) task.ID {
	return Transfer("POST", path, "", source, "", "", 0)
}

func Form(path, value, source, fileName, mime string, timeout int32) task.ID {
	r := net.NewRequest("POST", path, "")
	r.Timeout = timeout
	r.Parts.Append(net.Part{Name: "field", Value: value})
	if source != "" {
		r.Parts.Append(net.Part{Name: "file", Source: source, FileName: fileName, Mime: mime})
	}
	return net.Send("api", r, report)
}

func FormTyped(path, value, contentType string) task.ID {
	r := net.NewRequest("POST", path, "")
	r.Header.Set("content-type", contentType)
	r.Parts.Append(net.Part{Name: "field", Value: value})
	return net.Send("api", r, report)
}

func OnePart(path, name, value, source, fileName, mime, body, topSource string) task.ID {
	r := net.NewRequest("POST", path, body)
	r.Source = topSource
	r.Parts.Append(net.Part{Name: name, Value: value, Source: source, FileName: fileName, Mime: mime})
	return net.Send("api", r, report)
}
`

var rigBundle = sync.OnceValues(func() (*gostrict.Bundle, error) {
	prog, err := gostrict.LoadModules("llm.local/p", gostrict.Module{Path: "llm.local/p", FS: fstest.MapFS{"p.go": {Data: []byte(rigProgram)}}})
	if err != nil {
		return nil, err
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		return nil, errors.New(diags.Text())
	}
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		return nil, errors.New(diags.Text())
	}
	return b, nil
})

// rigConfig describes the host of a test. Everything but the url is optional.
type rigConfig struct {
	engine  gostrict.Engine
	opts    nethost.Options
	url     string
	conn    func(*nethost.Connection)
	resolve nethost.Resolver
	streams nethost.Streams
	defers  func(start, drop func())
	jobs    *taskhost.Jobs
}

// rig is a host with a pool of instances, which records all observations and ends.
type rig struct {
	t    *testing.T
	host *nethost.Host
	pool *gostrict.Pool
	ends chan string

	mu           sync.Mutex
	observations []nethost.Observation
	seenAtDone   []int
}

func newRig(t *testing.T, cfg rigConfig) *rig {
	t.Helper()
	if cfg.engine == nil {
		cfg.engine = &vm.Engine{}
	}
	base, err := url.Parse(cfg.url + "/api")
	if err != nil {
		t.Fatal(err)
	}
	r := &rig{t: t, ends: make(chan string, 1024), host: nethost.New(t.Context(), cfg.opts)}
	if cfg.resolve == nil {
		cfg.resolve = func(string) (nethost.Connection, error) {
			c := nethost.Connection{BaseURL: base, Secrets: []string{token}}
			if cfg.conn != nil {
				cfg.conn(&c)
			}
			return c, nil
		}
	}
	bundle, err := rigBundle()
	if err != nil {
		t.Fatal(err)
	}
	r.pool, err = gostrict.NewPoolFunc(t.Context(), cfg.engine, bundle, 4, func(*gostrict.Slot) gostrict.Funcs {
		b := nethost.Binding{
			Resolve:  cfg.resolve,
			Complete: r.complete,
			Streams:  cfg.streams,
			Defer:    cfg.defers,
			Observe: func(o nethost.Observation) {
				r.mu.Lock()
				defer r.mu.Unlock()
				r.observations = append(r.observations, o)
			},
		}
		if cfg.jobs != nil {
			b.Jobs = func() *taskhost.Jobs { return cfg.jobs }
		}
		return r.host.Funcs(b)
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.pool.Close() })
	return r
}

func (r *rig) complete(ctx context.Context, done gostrict.Detached, args []any) error {
	res, err := r.pool.CallDetached(ctx, done, args...)
	if err != nil {
		r.ends <- "engine: " + err.Error()
		return err
	}
	err = gostrict.ErrorOf(res[0])
	r.mu.Lock()
	r.seenAtDone = append(r.seenAtDone, len(r.observations))
	r.mu.Unlock()
	r.ends <- err.Error()
	return err
}

// send calls a function of the program and returns the id of its job.
func (r *rig) send(name string, args ...any) task.ID {
	r.t.Helper()
	res, err := r.pool.Call(r.t.Context(), name, args...)
	if err != nil {
		r.t.Fatal(err)
	}
	return task.ID(res[0].(string))
}

func (r *rig) end() string {
	r.t.Helper()
	select {
	case s := <-r.ends:
		return s
	case <-time.After(20 * time.Second):
		r.t.Fatal("the request did not end")
		return ""
	}
}

func (r *rig) observed() []nethost.Observation {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]nethost.Observation(nil), r.observations...)
}

func (r *rig) doneCalls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.seenAtDone)
}

// waitOpen waits until the host counts n open requests, which it frees after the end of a request.
func (r *rig) waitOpen(n int64) {
	r.t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for r.host.OpenRequests() != n {
		if time.Now().After(deadline) {
			r.t.Fatalf("open requests: got %d, want %d", r.host.OpenRequests(), n)
		}
		time.Sleep(time.Millisecond)
	}
}

// queue is an event loop, whose functions run when a test takes them, so that a wait is never too late.
type queue struct{ fns chan func() }

func newQueue() *queue { return &queue{fns: make(chan func(), 16)} }

func (q *queue) post(fn func()) bool {
	q.fns <- fn
	return true
}

// next runs the next posted function.
func (q *queue) next(t *testing.T) {
	t.Helper()
	select {
	case fn := <-q.fns:
		fn()
	case <-time.After(20 * time.Second):
		t.Fatal("the job has not ended")
	}
}

// jobFailure waits for the end of the job and returns its failure.
func jobFailure(t *testing.T, q *queue, jobs *taskhost.Jobs, id task.ID) string {
	t.Helper()
	var failure string
	jobs.Wait(id, func(f string) { failure = f })
	q.next(t)
	return failure
}

// memSink keeps its content in memory and records how it ended.
type memSink struct {
	mu          sync.Mutex
	data        bytes.Buffer
	writeErr    error
	shortWrite  bool
	commitErr   error
	stored      string
	commits     int
	aborts      int
	contentType string
	onWrite     func()
	onCommit    func()
}

func (s *memSink) Write(p []byte) (int, error) {
	if s.onWrite != nil {
		s.onWrite()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	if s.shortWrite {
		return len(p) - 1, nil
	}
	return s.data.Write(p)
}

func (s *memSink) Commit(contentType string) (string, error) {
	if s.onCommit != nil {
		s.onCommit()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contentType = contentType
	if s.commitErr != nil {
		return "", s.commitErr
	}
	s.commits++
	return s.stored, nil
}

func (s *memSink) Abort() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aborts++
}

func (s *memSink) state() (commits, aborts int, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commits, s.aborts, s.data.String()
}

// fakeStreams finds sources and sinks by name.
type fakeStreams struct {
	sources map[string]nethost.Source
	sinks   map[string]nethost.Sink
	err     error
}

func (s *fakeStreams) Source(name string) (nethost.Source, error) {
	if s.err != nil {
		return nethost.Source{}, s.err
	}
	src, ok := s.sources[name]
	if !ok {
		return nethost.Source{}, errors.New("unknown source " + name)
	}
	return src, nil
}

func (s *fakeStreams) Sink(name string) (nethost.Sink, error) {
	if s.err != nil {
		return nil, s.err
	}
	sink, ok := s.sinks[name]
	if !ok {
		return nil, errors.New("unknown sink " + name)
	}
	return sink, nil
}

// textSource is a source with the text. opens counts how often it was opened.
func textSource(text string, size int64, mime, name string, opens *atomic.Int32) nethost.Source {
	return nethost.Source{Size: size, Mime: mime, Name: name, Open: func() (io.ReadCloser, error) {
		if opens != nil {
			opens.Add(1)
		}
		return io.NopCloser(strings.NewReader(text)), nil
	}}
}

// countingHandler counts the requests of the wrapped handler.
func countingHandler(n *atomic.Int32, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		h(w, r)
	}
}
