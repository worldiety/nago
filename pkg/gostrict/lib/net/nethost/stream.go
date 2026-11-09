// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nethost

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// errStalled is the cause of the end of a transfer, which has moved no byte for too long.
var errStalled = errors.New("the transfer has stalled")

// upload is the content of a transfer.
type upload struct {
	open func() (io.ReadCloser, error)
	// size is the exact size or -1
	size        int64
	contentType string
	// replaceType makes contentType win over a header of the program, because a form needs its boundary.
	replaceType bool
}

// watchdog ends a transfer, which has moved no byte in any direction for a while. A timer per progress would cost much
// more than the atomic store, so the timer only compares the time of the last progress when it fires.
type watchdog struct {
	stall  time.Duration
	begin  time.Time
	last   atomic.Int64 // nanoseconds since begin
	expire func()

	mu      sync.Mutex
	timer   *time.Timer
	stopped bool
}

func startWatchdog(stall time.Duration, expire func()) *watchdog {
	w := &watchdog{stall: stall, begin: time.Now(), expire: expire}
	w.arm(stall)
	return w
}

func (w *watchdog) arm(after time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.stopped {
		w.timer = time.AfterFunc(after, w.check)
	}
}

func (w *watchdog) check() {
	idle := time.Since(w.begin) - time.Duration(w.last.Load())
	if idle >= w.stall {
		w.expire()
		return
	}
	w.arm(w.stall - idle)
}

func (w *watchdog) progress() {
	w.last.Store(int64(time.Since(w.begin)))
}

func (w *watchdog) stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stopped = true
	w.timer.Stop()
}

// progressReader reports each read byte to the watchdog. It also hides any faster path of its reader, so that all
// bytes pass through here.
type progressReader struct {
	r   io.Reader
	dog *watchdog
}

func (p progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.dog.progress()
	}
	return n, err
}

// uploadBody is the content, which the transport reads. A transport closes it on its own schedule, even during a read
// of another goroutine, which then fails without being a fault of the source.
type uploadBody struct {
	rc     io.ReadCloser
	x      *exchange
	closed atomic.Bool
}

func (b *uploadBody) Read(p []byte) (int, error) {
	n, err := b.rc.Read(p)
	if n > 0 && b.x.dog != nil {
		b.x.dog.progress()
	}
	if err != nil && err != io.EOF && !b.closed.Load() {
		b.x.fault.CompareAndSwap(nil, &err)
	}
	return n, err
}

// Close closes the source once, because a redirect or the end of the request may reach it after the transport.
func (b *uploadBody) Close() error {
	if b.closed.Swap(true) {
		return nil
	}
	return b.rc.Close()
}

// sinkWriter keeps the error of the sink apart from the error of the network.
type sinkWriter struct {
	sink Sink
	err  error
}

func (w *sinkWriter) Write(p []byte) (int, error) {
	n, err := w.sink.Write(p)
	if err == nil && n < len(p) {
		err = io.ErrShortWrite
	}
	if err != nil {
		w.err = err
	}
	return n, err
}

// multipartBody is a form, whose boundaries and headers are known before it is sent. So its files can be a chain of
// readers without a goroutine and it knows its length, if the sizes of its files are known.
type multipartBody struct {
	segments    []segment
	contentType string
	size        int64
}

// segment is either a fixed text or the content of a source.
type segment struct {
	text     string
	source   Source
	isSource bool
}

// newMultipart returns the form. It is false, if a type of a source can not be a header.
func newMultipart(parts list.List[net.Part], sources []Source) (*multipartBody, bool) {
	m := &multipartBody{}
	var buf bytes.Buffer
	cut := func() {
		if buf.Len() > 0 {
			m.segments = append(m.segments, segment{text: buf.String()})
			buf.Reset()
		}
	}
	w := multipart.NewWriter(&buf)
	for i, p := range parts.All() {
		fileName, mime := p.FileName, p.Mime
		if p.Source != "" {
			if fileName == "" {
				fileName = sources[i].Name
			}
			if mime == "" {
				mime = sources[i].Mime
			}
			if mime == "" {
				mime = "application/octet-stream"
			}
		}
		if !validMime(mime) {
			return nil, false
		}

		disposition := `form-data; name="` + quoteField(p.Name) + `"`
		if fileName != "" {
			disposition += `; filename="` + quoteField(fileName) + `"`
		}
		header := textproto.MIMEHeader{"Content-Disposition": {disposition}}
		if mime != "" {
			header.Set("Content-Type", mime)
		}
		// the writer of a form writes to the memory here, which never fails
		part, _ := w.CreatePart(header)
		if p.Source != "" {
			cut()
			m.segments = append(m.segments, segment{source: sources[i], isSource: true})
			continue
		}
		_, _ = io.WriteString(part, p.Value)
	}
	_ = w.Close()
	cut()

	m.contentType = w.FormDataContentType()
	for _, seg := range m.segments {
		switch {
		case !seg.isSource:
			m.size += int64(len(seg.text))
		case seg.source.Size < 0:
			m.size = -1
			return m, true
		default:
			m.size += seg.source.Size
		}
	}
	return m, true
}

func (m *multipartBody) upload() *upload {
	return &upload{open: m.open, size: m.size, contentType: m.contentType, replaceType: true}
}

func (m *multipartBody) open() (io.ReadCloser, error) {
	return &segmentReader{segments: m.segments}, nil
}

// quoteField returns the text for a quoted parameter of a header. A control character can not appear there at all, so
// it becomes percent-encoded like in a browser, otherwise a name could add headers or end the part.
func quoteField(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, "%%%02X", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// segmentReader reads the segments of a form one after the other and opens a file only when its turn comes, so a form
// with many files holds one of them open at a time.
type segmentReader struct {
	mu       sync.Mutex
	segments []segment
	current  io.ReadCloser
	closed   bool
}

// next returns the reader of the current segment. It holds the lock only to switch, never during a read, because
// Close must not wait for a read.
func (s *segmentReader) next() (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, http.ErrBodyReadAfterClose
	}
	if s.current != nil {
		return s.current, nil
	}
	if len(s.segments) == 0 {
		return nil, io.EOF
	}
	seg := s.segments[0]
	s.segments = s.segments[1:]
	if seg.isSource {
		rc, err := seg.source.Open()
		if err != nil {
			return nil, err
		}
		s.current = rc
	} else {
		s.current = io.NopCloser(strings.NewReader(seg.text))
	}
	return s.current, nil
}

func (s *segmentReader) done(rc io.ReadCloser) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == rc {
		s.current = nil
		return rc.Close()
	}
	return nil
}

func (s *segmentReader) Read(p []byte) (int, error) {
	for {
		rc, err := s.next()
		if err != nil {
			return 0, err
		}
		n, err := rc.Read(p)
		if err != io.EOF {
			return n, err
		}
		if err := s.done(rc); err != nil {
			return n, err
		}
		if n > 0 {
			return n, nil
		}
	}
}

func (s *segmentReader) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if s.current != nil {
		rc := s.current
		s.current = nil
		return rc.Close()
	}
	return nil
}
