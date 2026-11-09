// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nethost_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
)

func TestDownload(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		fmt.Fprint(w, "\x00\xff binary")
	})
	mux.HandleFunc("/api/nf", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusNotFound) })
	mux.HandleFunc("/api/cut", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		fmt.Fprint(w, "half")
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	for _, tt := range []struct {
		name        string
		path        string
		sink        *memSink
		maxBody     int64
		end         string
		commits     int
		aborts      int
		content     string
		contentType string
	}{
		{name: "the answer goes to the sink, not to the program", path: "/ok", sink: &memSink{stored: "file-1"}, end: "ok|200|||file-1", commits: 1, content: "\x00\xff binary", contentType: "image/png"},
		{name: "the text limit does not bind a sink", path: "/ok", sink: &memSink{stored: "file-1"}, maxBody: 2, end: "ok|200|||file-1", commits: 1, content: "\x00\xff binary", contentType: "image/png"},
		{name: "what the sink reports loses its secrets", path: "/ok", sink: &memSink{stored: "at " + token}, end: "ok|200|||at ***", commits: 1, content: "\x00\xff binary", contentType: "image/png"},
		{name: "an error status aborts the sink", path: "/nf", sink: &memSink{}, end: "rejected|404|HTTP 404: no\n||", aborts: 1},
		{name: "a failing write aborts the sink", path: "/ok", sink: &memSink{writeErr: errors.New("disk at /secret/path full")}, end: "unknown|200|the answer can not be stored||", aborts: 1},
		{name: "a refused write shows its reason", path: "/ok", sink: &memSink{writeErr: nethost.Refused("the answer is too large")}, end: "unknown|200|the answer is too large||", aborts: 1},
		{name: "a sink, which writes less, is no silent loss", path: "/ok", sink: &memSink{shortWrite: true}, end: "unknown|200|the answer can not be stored||", aborts: 1},
		{name: "a commit never shows its error", path: "/ok", sink: &memSink{commitErr: nethost.Refused("internal text")}, end: "unknown|200|the answer can not be stored||", aborts: 1, content: "\x00\xff binary", contentType: "image/png"},
		{name: "a failing commit keeps the status and aborts", path: "/ok", sink: &memSink{commitErr: errors.New("disk at /secret/path full")}, end: "unknown|200|the answer can not be stored||", aborts: 1, content: "\x00\xff binary", contentType: "image/png"},
		{name: "a lost connection aborts the sink", path: "/cut", sink: &memSink{}, end: "unknown|200|the connection was lost after the request was sent||", aborts: 1, content: "half"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true, MaxBody: tt.maxBody}, streams: &fakeStreams{sinks: map[string]nethost.Sink{"s": tt.sink}}})
			r.send("Download", tt.path, "s")
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			commits, aborts, content := tt.sink.state()
			if commits != tt.commits || aborts != tt.aborts || content != tt.content {
				t.Errorf("commits %d aborts %d content %q", commits, aborts, content)
			}
			if tt.sink.contentType != tt.contentType {
				t.Errorf("content type: %q", tt.sink.contentType)
			}
		})
	}
}

func TestStreamsAreChecked(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(countingHandler(&hits, func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "hello") }))
	t.Cleanup(srv.Close)
	file := textSource("content", 7, "", "", nil)
	sources := map[string]nethost.Source{"f": file}
	ok := &fakeStreams{sources: sources, sinks: map[string]nethost.Sink{"s": &memSink{}}}

	for _, tt := range []struct {
		name    string
		streams nethost.Streams
		fn      string
		args    []any
		end     string
		hits    int32
	}{
		{name: "a source with streams", streams: ok, fn: "Upload", args: []any{"/x", "f"}, end: "ok|200||hello|", hits: 1},
		{name: "a source without streams", fn: "Upload", args: []any{"/x", "f"}, end: "failed|0|this host can not transfer files||"},
		{name: "a sink with streams", streams: ok, fn: "Download", args: []any{"/x", "s"}, end: "ok|200|||", hits: 1},
		{name: "a sink without streams", fn: "Download", args: []any{"/x", "s"}, end: "failed|0|this host can not transfer files||"},
		{name: "a part with a source with streams", streams: ok, fn: "Form", args: []any{"/x", "v", "f", "", "", int32(0)}, end: "ok|200||hello|", hits: 1},
		{name: "a part with a source without streams", fn: "Form", args: []any{"/x", "v", "f", "", "", int32(0)}, end: "failed|0|this host can not transfer files||"},
		{name: "a form of values needs no streams", fn: "Form", args: []any{"/x", "v", "", "", "", int32(0)}, end: "ok|200||hello|", hits: 1},
		{name: "a refusal of the streams is visible", streams: &fakeStreams{err: nethost.Refused("no right to read this file")}, fn: "Upload", args: []any{"/x", "f"}, end: "failed|0|no right to read this file||"},
		{name: "a refusal of the sink is visible", streams: &fakeStreams{err: nethost.Refused("no right to write")}, fn: "Download", args: []any{"/x", "s"}, end: "failed|0|no right to write||"},
		{name: "any other error stays hidden", streams: &fakeStreams{err: errors.New("/var/lib/secret: denied")}, fn: "Upload", args: []any{"/x", "f"}, end: "failed|0|the file is not available||"},
		{name: "a body and a source contradict", streams: ok, fn: "Transfer", args: []any{"POST", "/x", "text", "f", "", "", int32(0)}, end: "failed|0|a request can not have a source and a body||"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hits.Store(0)
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: tt.streams})
			r.send(tt.fn, tt.args...)
			if got := r.end(); got != tt.end {
				t.Errorf("got %q, want %q", got, tt.end)
			}
			if hits.Load() != tt.hits {
				t.Errorf("hits: got %d, want %d", hits.Load(), tt.hits)
			}
			if len(r.observed()) != 1 {
				t.Errorf("observations: %+v", r.observed())
			}
		})
	}
}

// drain reads the body and hangs up after a broken one, because a server would otherwise wait half a second for it.
func drain(r *http.Request) {
	if _, err := io.Copy(io.Discard, r.Body); err != nil {
		panic(http.ErrAbortHandler)
	}
}

// failingReader delivers its text and then fails.
type failingReader struct {
	text string
	err  error
}

func (f *failingReader) Read(p []byte) (int, error) {
	if f.text == "" {
		return 0, f.err
	}
	n := copy(p, f.text)
	f.text = f.text[n:]
	return n, nil
}

func (f *failingReader) Close() error { return nil }

func TestUpload(t *testing.T) {
	type seen struct {
		method, contentType, body string
		length                    int64
		chunked                   bool
	}
	var mu sync.Mutex
	var got []seen
	mux := http.NewServeMux()
	record := func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		defer mu.Unlock()
		got = append(got, seen{method: r.Method, contentType: r.Header.Get("Content-Type"), body: string(b), length: r.ContentLength, chunked: len(r.TransferEncoding) > 0})
		fmt.Fprint(w, "stored")
	}
	mux.HandleFunc("/api/x", record)
	mux.HandleFunc("/api/moved", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		http.Redirect(w, r, "/api/x", http.StatusTemporaryRedirect)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	const text = "a file"
	for _, tt := range []struct {
		name        string
		path        string
		source      nethost.Source
		contentType string
		opens       int32
		end         string
		want        *seen
	}{
		{name: "known size", path: "/x", source: textSource(text, 6, "text/plain", "", nil), opens: 1, end: "ok|200||stored|",
			want: &seen{method: "POST", contentType: "text/plain", body: text, length: 6}},
		{name: "the type of the program wins", path: "/x", source: textSource(text, 6, "text/plain", "", nil), contentType: "application/x-own", opens: 1, end: "ok|200||stored|",
			want: &seen{method: "POST", contentType: "application/x-own", body: text, length: 6}},
		{name: "unknown size", path: "/x", source: textSource(text, -1, "", "", nil), opens: 1, end: "ok|200||stored|",
			want: &seen{method: "POST", body: text, length: -1, chunked: true}},
		{name: "empty source is sent without opening", path: "/x", source: textSource("", 0, "", "", nil), opens: 0, end: "ok|200||stored|",
			want: &seen{method: "POST", length: 0}},
		{name: "a redirect, which needs the content, opens it again", path: "/moved", source: textSource(text, 6, "text/plain", "", nil), opens: 2, end: "ok|200||stored|",
			want: &seen{method: "POST", contentType: "text/plain", body: text, length: 6}},
		{name: "a type of the host, which can not be a header, is refused", path: "/x", source: textSource(text, 6, "a\nb", "", nil), opens: 0, end: "failed|0|the file is not available||"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mu.Lock()
			got = nil
			mu.Unlock()
			var opens atomic.Int32
			src := tt.source
			open := src.Open
			src.Open = func() (io.ReadCloser, error) {
				opens.Add(1)
				return open()
			}
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: map[string]nethost.Source{"f": src}}})
			r.send("Transfer", "POST", tt.path, "", "f", "", tt.contentType, int32(0))
			if end := r.end(); end != tt.end {
				t.Errorf("end: got %q, want %q", end, tt.end)
			}
			if opens.Load() != tt.opens {
				t.Errorf("opens: got %d, want %d", opens.Load(), tt.opens)
			}
			mu.Lock()
			defer mu.Unlock()
			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("nothing must arrive: %+v", got)
				}
				return
			}
			if len(got) != 1 || got[0] != *tt.want {
				t.Errorf("got %+v, want %+v", got, *tt.want)
			}
		})
	}
}

func TestSourceFails(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(countingHandler(&hits, func(w http.ResponseWriter, r *http.Request) {
		drain(r)
		if r.URL.Path == "/api/moved" {
			http.Redirect(w, r, "/api/x", http.StatusTemporaryRedirect)
			return
		}
		fmt.Fprint(w, "stored")
	}))
	t.Cleanup(srv.Close)
	onlyOnce := func() nethost.Source {
		var opened atomic.Bool
		return nethost.Source{Size: 3, Open: func() (io.ReadCloser, error) {
			if opened.Swap(true) {
				return nil, errors.New("gone")
			}
			return io.NopCloser(strings.NewReader("abc")), nil
		}}
	}
	for _, tt := range []struct {
		name   string
		path   string
		source nethost.Source
		end    string
		class  nethost.Class
	}{
		{name: "a redirect, whose content can not be opened again", path: "/moved", source: onlyOnce(), end: "unknown|0|the file can not be read||", class: nethost.NotSent},
		{name: "a redirect, whose content can be opened again", path: "/moved", source: textSource("abc", 3, "", "", nil), end: "ok|200||stored|", class: nethost.Answered},
		{name: "healthy", source: textSource("abc", 3, "", "", nil), end: "ok|200||stored|", class: nethost.Answered},
		{name: "open fails", source: nethost.Source{Size: 3, Open: func() (io.ReadCloser, error) { return nil, errors.New("gone") }}, end: "failed|0|the file can not be read||", class: nethost.NotSent},
		{name: "open refuses", source: nethost.Source{Size: 3, Open: func() (io.ReadCloser, error) { return nil, nethost.Refused("the file was removed") }}, end: "failed|0|the file was removed||", class: nethost.NotSent},
		{name: "reading fails in the middle", source: nethost.Source{Size: 100, Open: func() (io.ReadCloser, error) { return &failingReader{text: "abc", err: errors.New("bad disk")}, nil }}, end: "unknown|0|the file can not be read||", class: nethost.NotSent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: map[string]nethost.Source{"f": tt.source}}})
			path := tt.path
			if path == "" {
				path = "/x"
			}
			r.send("Upload", path, "f")
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			if obs := r.observed(); len(obs) != 1 || obs[0].Class != tt.class {
				t.Errorf("observations: %+v, want class %d", obs, tt.class)
			}
		})
	}
}

// TestFormFields sends the form with both engines, because its list and its records cross the border of the engine.
func TestFormFields(t *testing.T) {
	for name, engine := range engines() {
		t.Run(name, func(t *testing.T) {
			got := make(chan map[string]string, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				form := map[string]string{}
				mr, err := r.MultipartReader()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				for {
					p, err := mr.NextPart()
					if err != nil {
						break
					}
					b, _ := io.ReadAll(p)
					form[p.FormName()] = string(b)
				}
				got <- form
				fmt.Fprint(w, "stored")
			}))
			t.Cleanup(srv.Close)
			r := newRig(t, rigConfig{engine: engine, url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: map[string]nethost.Source{"f": textSource("file", 4, "text/x", "a.txt", nil)}}})
			r.send("Form", "/x", "value", "f", "", "", int32(0))
			if end := r.end(); end != "ok|200||stored|" {
				t.Errorf("end: %q", end)
			}
			form := <-got
			if form["field"] != "value" || form["file"] != "file" {
				t.Errorf("form: %v", form)
			}
		})
	}
}

// rawForm is what a server has received as a form.
type rawForm struct {
	contentType string
	length      int64
	chunked     bool
	raw         []byte
	parts       []rawPart
}

type rawPart struct {
	name, file, mime, value string
	headers                 int
}

func parseForm(t *testing.T, contentType string, raw []byte) []rawPart {
	t.Helper()
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("content type %q: %v", contentType, err)
	}
	var res []rawPart
	mr := multipart.NewReader(bytes.NewReader(raw), params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			return res
		}
		if err != nil {
			t.Fatalf("part: %v\n%q", err, raw)
		}
		v, _ := io.ReadAll(p)
		res = append(res, rawPart{name: p.FormName(), file: p.FileName(), mime: p.Header.Get("Content-Type"), value: string(v), headers: len(p.Header)})
	}
}

func TestMultipart(t *testing.T) {
	var mu sync.Mutex
	var last *rawForm
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		last = &rawForm{contentType: r.Header.Get("Content-Type"), length: r.ContentLength, chunked: len(r.TransferEncoding) > 0, raw: b}
		mu.Unlock()
		fmt.Fprint(w, "stored")
	}))
	t.Cleanup(srv.Close)
	take := func() *rawForm {
		mu.Lock()
		defer mu.Unlock()
		l := last
		last = nil
		return l
	}

	known := textSource("file content", 12, "text/x", "stored.txt", nil)
	unknown := textSource("file content", -1, "", "", nil)
	sources := map[string]nethost.Source{"known": known, "unknown": unknown, "badmime": textSource("x", 1, "a\nb", "", nil)}

	for _, tt := range []struct {
		name  string
		fn    string
		args  []any
		end   string
		exact bool
		parts []rawPart
	}{
		{name: "known sizes have an exact length", fn: "Form", args: []any{"/x", "v", "known", "", "", int32(0)}, end: "ok|200||stored|", exact: true,
			parts: []rawPart{{name: "field", value: "v", headers: 1}, {name: "file", file: "stored.txt", mime: "text/x", value: "file content", headers: 2}}},
		{name: "the name and type of the program win", fn: "Form", args: []any{"/x", "v", "known", "own.bin", "application/x-own", int32(0)}, end: "ok|200||stored|", exact: true,
			parts: []rawPart{{name: "field", value: "v", headers: 1}, {name: "file", file: "own.bin", mime: "application/x-own", value: "file content", headers: 2}}},
		{name: "an unknown size is chunked with a default type", fn: "Form", args: []any{"/x", "v", "unknown", "", "", int32(0)}, end: "ok|200||stored|",
			parts: []rawPart{{name: "field", value: "v", headers: 1}, {name: "file", mime: "application/octet-stream", value: "file content", headers: 2}}},
		{name: "the host sets the type of a form", fn: "FormTyped", args: []any{"/x", "v", "multipart/form-data"}, end: "ok|200||stored|", exact: true,
			parts: []rawPart{{name: "field", value: "v", headers: 1}}},
		{name: "a quote stays in the name", fn: "OnePart", args: []any{"/x", `a"b`, "v", "", "", "", "", ""}, end: "ok|200||stored|", exact: true,
			parts: []rawPart{{name: `a"b`, value: "v", headers: 1}}},
		{name: "a line break in a file name can not add a header", fn: "OnePart", args: []any{"/x", "n", "v", "", "a\r\nX-Evil: 1", "", "", ""}, end: "ok|200||stored|", exact: true,
			parts: []rawPart{{name: "n", file: "a%0D%0AX-Evil: 1", value: "v", headers: 1}}},
		{name: "a line break in a field name can not add a header", fn: "OnePart", args: []any{"/x", "n\r\nX-Evil: 1", "v", "", "", "", "", ""}, end: "ok|200||stored|", exact: true,
			parts: []rawPart{{name: "n%0D%0AX-Evil: 1", value: "v", headers: 1}}},

		{name: "a part without a name", fn: "OnePart", args: []any{"/x", "", "v", "", "", "", "", ""}, end: "failed|0|the parts of the form are not valid||"},
		{name: "a part with a value and a source", fn: "OnePart", args: []any{"/x", "n", "v", "known", "", "", "", ""}, end: "failed|0|the parts of the form are not valid||"},
		{name: "a type of the program with a line break", fn: "OnePart", args: []any{"/x", "n", "v", "", "", "a\nb", "", ""}, end: "failed|0|the parts of the form are not valid||"},
		{name: "a type of the host with a line break", fn: "OnePart", args: []any{"/x", "n", "", "badmime", "", "", "", ""}, end: "failed|0|the parts of the form are not valid||"},
		{name: "parts and a body", fn: "OnePart", args: []any{"/x", "n", "v", "", "", "", "text", ""}, end: "failed|0|a request with parts can not have a body or a source||"},
		{name: "parts and a source", fn: "OnePart", args: []any{"/x", "n", "v", "", "", "", "", "known"}, end: "failed|0|a request with parts can not have a body or a source||"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: sources}})
			r.send(tt.fn, tt.args...)
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			form := take()
			if tt.parts == nil {
				if form != nil {
					t.Errorf("nothing must arrive: %q", form.raw)
				}
				return
			}
			if form == nil {
				t.Fatal("no form has arrived")
			}
			if !strings.HasPrefix(form.contentType, "multipart/form-data; boundary=") {
				t.Errorf("content type: %q", form.contentType)
			}
			if tt.exact {
				if form.chunked || form.length != int64(len(form.raw)) {
					t.Errorf("length %d chunked %v, body has %d", form.length, form.chunked, len(form.raw))
				}
			} else if !form.chunked {
				t.Errorf("an unknown size must be chunked: %d", form.length)
			}
			parts := parseForm(t, form.contentType, form.raw)
			if fmt.Sprint(parts) != fmt.Sprint(tt.parts) {
				t.Errorf("parts:\n got %+v\nwant %+v", parts, tt.parts)
			}
		})
	}
}

// TestMultipartOpensOneFileAtATime reads a form of many files, and an unreadable file ends it after the first one.
func TestMultipartOpensOneFileAtATime(t *testing.T) {
	var open, peak atomic.Int32
	source := func(text string) nethost.Source {
		return nethost.Source{Size: int64(len(text)), Open: func() (io.ReadCloser, error) {
			n := open.Add(1)
			for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
			}
			return closeCounter{io.NopCloser(strings.NewReader(text)), &open}, nil
		}}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		drain(r)
		fmt.Fprint(w, "stored")
	}))
	t.Cleanup(srv.Close)
	r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: map[string]nethost.Source{"a": source("aaa"), "bad": {Size: 3, Open: func() (io.ReadCloser, error) { return nil, errors.New("gone") }}}}})
	r.send("Form", "/x", "v", "a", "", "", int32(0))
	if got := r.end(); got != "ok|200||stored|" {
		t.Error(got)
	}
	if peak.Load() != 1 || open.Load() != 0 {
		t.Errorf("peak %d open %d", peak.Load(), open.Load())
	}
	r.send("Form", "/x", "v", "bad", "", "", int32(0))
	got := r.end()
	if got != "unknown|0|the file can not be read||" {
		t.Errorf("an unreadable file must end the form: %q", got)
	}
}

type closeCounter struct {
	io.ReadCloser
	open *atomic.Int32
}

func (c closeCounter) Close() error {
	c.open.Add(-1)
	return c.ReadCloser.Close()
}

// pacedReader delivers a byte in each read until the time is over.
type pacedReader struct {
	until time.Time
}

func (p *pacedReader) Read(b []byte) (int, error) {
	if time.Now().After(p.until) {
		return 0, io.EOF
	}
	b[0] = 'x'
	return 1, nil
}

func (p *pacedReader) Close() error { return nil }

// blockingSource delivers one byte and then waits until it is closed.
type blockingSource struct {
	once   sync.Once
	closed chan struct{}
}

func (b *blockingSource) Read(p []byte) (int, error) {
	var first bool
	b.once.Do(func() { first = true })
	if first {
		p[0] = 'x'
		return 1, nil
	}
	<-b.closed
	return 0, io.ErrClosedPipe
}

func (b *blockingSource) Close() error {
	select {
	case <-b.closed:
	default:
		close(b.closed)
	}
	return nil
}

func TestTransferLimits(t *testing.T) {
	mux := http.NewServeMux()
	silent := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "x")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}
	mux.HandleFunc("/api/silent", silent)
	mux.HandleFunc("/api/hang", func(_ http.ResponseWriter, r *http.Request) {
		// a server only notices the end of the client after it has read the body
		drain(r)
		<-r.Context().Done()
	})
	mux.HandleFunc("/api/slow", func(w http.ResponseWriter, _ *http.Request) {
		// the slow answer of a server is what this row is about, it synchronizes nothing
		time.Sleep(120 * time.Millisecond)
		fmt.Fprint(w, "x")
	})
	mux.HandleFunc("/api/flood", func(w http.ResponseWriter, r *http.Request) {
		for r.Context().Err() == nil {
			if _, err := fmt.Fprint(w, "x"); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	})
	mux.HandleFunc("/api/big", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		fmt.Fprint(w, strings.Repeat("x", 100))
	})
	mux.HandleFunc("/api/drain", func(w http.ResponseWriter, r *http.Request) {
		drain(r)
		fmt.Fprint(w, "stored")
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	file := textSource("content", 7, "", "", nil)
	for _, tt := range []struct {
		name string
		opts nethost.Options
		fn   string
		args []any
		end  string
		src  nethost.Source
	}{
		{name: "the program timeout ends a text request", opts: nethost.Options{StreamStall: time.Minute}, fn: "Timed", args: []any{"/hang", int32(30)}, end: "unknown|0|the request has timed out after it was sent||"},
		{name: "the program timeout does not bind a download", opts: nethost.Options{StreamStall: time.Minute}, fn: "Transfer", args: []any{"GET", "/slow", "", "", "s", "", int32(30)}, end: "ok|200|||"},
		{name: "the program timeout does not bind an upload", opts: nethost.Options{StreamStall: time.Minute}, fn: "Transfer", args: []any{"POST", "/slow", "", "f", "", "", int32(30)}, end: "ok|200||x|", src: file},
		{name: "the program timeout binds a form of values", opts: nethost.Options{StreamStall: time.Minute}, fn: "Form", args: []any{"/hang", "v", "", "", "", int32(30)}, end: "unknown|0|the request has timed out after it was sent||"},
		{name: "a stalled answer of a download ends", opts: nethost.Options{StreamStall: 60 * time.Millisecond}, fn: "Download", args: []any{"/silent", "s"}, end: "unknown|200|the transfer has stalled||"},
		{name: "a server without an answer stalls a transfer", opts: nethost.Options{StreamStall: 60 * time.Millisecond}, fn: "Download", args: []any{"/hang", "s"}, end: "unknown|0|the transfer has stalled||"},
		{name: "a stalled source ends an upload", opts: nethost.Options{StreamStall: 60 * time.Millisecond}, fn: "Upload", args: []any{"/drain", "f"}, end: "unknown|0|the transfer has stalled||",
			src: nethost.Source{Size: -1, Open: func() (io.ReadCloser, error) { return &blockingSource{closed: make(chan struct{})}, nil }}},
		{name: "the whole time of a transfer ends", opts: nethost.Options{StreamStall: time.Minute, MaxStream: 80 * time.Millisecond}, fn: "Download", args: []any{"/flood", "s"}, end: "unknown|200|the request has timed out after it was sent||"},
		{name: "the text of an upload stays limited", opts: nethost.Options{StreamStall: time.Minute, MaxBody: 10}, fn: "Upload", args: []any{"/big", "f"}, end: "unknown|200|the answer is too large||", src: file},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.AllowPrivateNetworks = true
			sink := &memSink{}
			src := tt.src
			if src.Open == nil {
				src = file
			}
			r := newRig(t, rigConfig{url: srv.URL, opts: tt.opts, streams: &fakeStreams{sources: map[string]nethost.Source{"f": src}, sinks: map[string]nethost.Sink{"s": sink}}})
			r.send(tt.fn, tt.args...)
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			obs := r.observed()
			if len(obs) != 1 {
				t.Fatalf("observations: %+v", obs)
			}
			if strings.HasPrefix(tt.end, "unknown") && !strings.Contains(tt.end, "too large") && obs[0].Class != nethost.Network {
				t.Errorf("class: %+v", obs[0])
			}
		})
	}
}

// TestProgressKeepsATransferAlive moves bytes for longer than the stall limit, in each direction. The limit is so high
// that a busy machine does not stall a transfer, which moves a byte in each of its steps.
func TestProgressKeepsATransferAlive(t *testing.T) {
	const stall = 250 * time.Millisecond
	mux := http.NewServeMux()
	mux.HandleFunc("/api/drip", func(w http.ResponseWriter, r *http.Request) {
		until := time.Now().Add(stall + stall/2)
		for time.Now().Before(until) {
			if _, err := fmt.Fprint(w, "x"); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	})
	mux.HandleFunc("/api/drain", func(w http.ResponseWriter, r *http.Request) {
		drain(r)
		fmt.Fprint(w, "stored")
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	for _, tt := range []struct {
		name string
		fn   string
		args []any
		end  string
	}{
		{name: "download", fn: "Download", args: []any{"/drip", "s"}, end: "ok|200|||"},
		{name: "upload", fn: "Upload", args: []any{"/drain", "f"}, end: "ok|200||stored|"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			begin := time.Now()
			src := nethost.Source{Size: -1, Open: func() (io.ReadCloser, error) { return &pacedReader{until: time.Now().Add(stall + stall/2)}, nil }}
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true, StreamStall: stall}, streams: &fakeStreams{sources: map[string]nethost.Source{"f": src}, sinks: map[string]nethost.Sink{"s": &memSink{}}}})
			r.send(tt.fn, tt.args...)
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			if time.Since(begin) <= stall {
				t.Errorf("the transfer was too short to prove anything: %v", time.Since(begin))
			}
		})
	}
}

// heapProbe records the largest growth of the live heap during a transfer.
type heapProbe struct {
	mu   sync.Mutex
	base uint64
	peak uint64
	seen int64
	next int64
}

const heapStep = 8 << 20

// start reads the live heap, from which the growth is measured.
func (p *heapProbe) start() {
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	p.base, p.next = ms.HeapAlloc, heapStep
}

func (p *heapProbe) moved(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seen += int64(n)
	if p.seen < p.next {
		return
	}
	p.next += heapStep
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	if ms.HeapAlloc > p.base {
		p.peak = max(p.peak, ms.HeapAlloc-p.base)
	}
}

// probeSink measures in each write. With retain, it holds all content in memory, which is what the host must not do.
type probeSink struct {
	probe  *heapProbe
	retain bool
	kept   [][]byte
	total  int64
}

func (s *probeSink) Write(p []byte) (int, error) {
	s.total += int64(len(p))
	if s.retain {
		s.kept = append(s.kept, bytes.Clone(p))
	}
	s.probe.moved(len(p))
	return len(p), nil
}

func (s *probeSink) Commit(string) (string, error) { return "", nil }
func (s *probeSink) Abort()                        {}

// probeSource delivers size bytes and measures in each read.
type probeSource struct {
	probe  *heapProbe
	left   int64
	retain bool
	kept   [][]byte
}

func (s *probeSource) Read(p []byte) (int, error) {
	if s.left <= 0 {
		return 0, io.EOF
	}
	n := int(min(int64(len(p)), s.left))
	clear(p[:n])
	s.left -= int64(n)
	if s.retain {
		s.kept = append(s.kept, bytes.Clone(p[:n]))
	}
	s.probe.moved(n)
	return n, nil
}

func (s *probeSource) Close() error { return nil }

// TestTransferNeedsNoMemoryOfItsSize moves 64 MiB through the host in each direction and reads the growth of the live
// heap at each 8 MiB. The rows with a part, which keeps everything in memory, show that the probe sees it.
func TestTransferNeedsNoMemoryOfItsSize(t *testing.T) {
	const size = 64 << 20
	const limit = 16 << 20
	chunk := make([]byte, 32<<10)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			n, _ := io.Copy(io.Discard, r.Body)
			fmt.Fprint(w, n)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(size))
		for left := size; left > 0; left -= len(chunk) {
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	for _, tt := range []struct {
		name   string
		upload bool
		retain bool
		over   bool
	}{
		{name: "download"},
		{name: "upload", upload: true},
		{name: "a sink, which keeps all, is seen", retain: true, over: true},
		{name: "a source, which keeps all, is seen", upload: true, retain: true, over: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			probe := &heapProbe{}
			sink := &probeSink{probe: probe, retain: tt.retain && !tt.upload}
			src := nethost.Source{Size: size, Open: func() (io.ReadCloser, error) {
				return &probeSource{probe: probe, left: size, retain: tt.retain}, nil
			}}
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: map[string]nethost.Source{"f": src}, sinks: map[string]nethost.Sink{"s": sink}}})
			// the rig itself must not count as growth of the transfer
			probe.start()

			if tt.upload {
				r.send("Upload", "/up", "f")
				if got, want := r.end(), fmt.Sprintf("ok|200||%d|", size); got != want {
					t.Fatalf("got %q, want %q", got, want)
				}
			} else {
				r.send("Download", "/down", "s")
				if got := r.end(); got != "ok|200|||" {
					t.Fatal(got)
				}
				if sink.total != size {
					t.Fatalf("the sink got %d bytes", sink.total)
				}
			}
			t.Logf("largest growth of the live heap: %.2f MiB", float64(probe.peak)/(1<<20))
			switch {
			case tt.over && probe.peak < limit:
				t.Errorf("the probe does not see what is kept: %d", probe.peak)
			case !tt.over && probe.peak >= limit:
				t.Errorf("the transfer needed %d bytes of live heap", probe.peak)
			}
		})
	}
}

// TestRedirectKeepsTheTransferAddress shows that a transfer does not leave its origin either.
func TestRedirectKeepsTheTransferAddress(t *testing.T) {
	var elsewhere atomic.Int32
	other := httptest.NewServer(countingHandler(&elsewhere, func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(other.Close)
	target, _ := url.Parse(other.URL)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.String(), http.StatusTemporaryRedirect)
	}))
	t.Cleanup(srv.Close)
	// the client of Go opens the content for a redirect before it decides, so the host must also close it then
	var open atomic.Int32
	src := nethost.Source{Size: 1, Open: func() (io.ReadCloser, error) {
		open.Add(1)
		return closeCounter{io.NopCloser(strings.NewReader("x")), &open}, nil
	}}
	sink := &memSink{}
	r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sources: map[string]nethost.Source{"f": src}, sinks: map[string]nethost.Sink{"s": sink}}})
	r.send("Transfer", "POST", "/x", "", "f", "s", "", int32(0))
	got := r.end()
	if !strings.HasPrefix(got, "rejected|307|") {
		t.Errorf("end: %q", got)
	}
	if elsewhere.Load() != 0 {
		t.Errorf("elsewhere %d", elsewhere.Load())
	}
	deadline := time.Now().Add(20 * time.Second)
	for open.Load() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("%d opened contents were not closed", open.Load())
		}
		time.Sleep(time.Millisecond)
	}
	if _, aborts, _ := sink.state(); aborts != 1 {
		t.Errorf("aborts: %d", aborts)
	}
}
