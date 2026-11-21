// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"slices"
	"sync"
)

// maxCachedLogIndexes bounds the log files, whose index is kept in memory.
const maxCachedLogIndexes = 16

// logPrefixLen is the length of the beginning of a line, which contains the time and the level.
const logPrefixLen = 256

// logLine is the position and the level of a complete line of a log file.
type logLine struct {
	off   int64
	n     int32
	level slog.Level
}

// logIndex indexes the complete lines of a log file up to size.
type logIndex struct {
	info  os.FileInfo
	size  int64
	lines []logLine
	used  uint64
}

// logIndexCache keeps the line index of recently read log files, so that a page of a log reads only its own lines
// and the part of the file, which has been written since the last read, instead of the whole file.
type logIndexCache struct {
	mu    sync.Mutex
	files map[string]*logIndex
	tick  uint64
}

// forget drops the index of the given file, e.g. because it has been deleted.
func (c *logIndexCache) forget(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.files, path)
}

// lines returns the index of the complete lines of the given open file.
func (c *logIndexCache) lines(f *os.File, path string) ([]logLine, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	idx := c.files[path]
	if idx == nil || idx.info == nil || !os.SameFile(idx.info, fi) || idx.size > fi.Size() {
		idx = &logIndex{info: fi}
		c.files[path] = idx
		c.evict()
	}

	if idx.size < fi.Size() {
		if err := idx.extend(f, fi.Size()); err != nil {
			delete(c.files, path)
			return nil, err
		}
	}

	c.tick++
	idx.used = c.tick

	// lines are only appended, so that the returned slice stays valid
	return idx.lines[:len(idx.lines):len(idx.lines)], nil
}

// evict drops the least recently used indexes beyond [maxCachedLogIndexes].
func (c *logIndexCache) evict() {
	for len(c.files) > maxCachedLogIndexes {
		var oldest string
		var used uint64
		first := true
		for path, idx := range c.files {
			if first || idx.used < used {
				oldest, used, first = path, idx.used, false
			}
		}

		delete(c.files, oldest)
	}
}

// extend indexes the complete lines between the indexed size and the given size. A partially written last line is
// indexed by a later call.
func (idx *logIndex) extend(f *os.File, size int64) error {
	br := bufio.NewReaderSize(io.NewSectionReader(f, idx.size, size-idx.size), 256*1024)
	off := idx.size
	prefix := make([]byte, 0, logPrefixLen)
	for {
		start := off
		prefix = prefix[:0]
		var n int64
		complete := false
		for {
			chunk, err := br.ReadSlice('\n')
			if rest := logPrefixLen - len(prefix); rest > 0 {
				prefix = append(prefix, chunk[:min(len(chunk), rest)]...)
			}

			n += int64(len(chunk))
			if err == nil {
				complete = true
				break
			}

			if err == bufio.ErrBufferFull {
				continue
			}

			if err == io.EOF {
				break
			}

			return err
		}

		if !complete {
			return nil
		}

		off += n
		idx.size = off

		level, ok := parseLogLevel(prefix)
		if !ok {
			// not written by a sink, e.g. a broken line: decode it as a whole
			buf := make([]byte, n)
			if _, err := f.ReadAt(buf, start); err != nil {
				return err
			}

			var e fileEntry
			if err := json.Unmarshal(buf, &e); err != nil {
				continue
			}

			level = e.Level
		}

		idx.lines = append(idx.lines, logLine{off: start, n: int32(min(n, 1<<31-1)), level: level})
	}
}

var levelKey = []byte(`,"level":`)

// parseLogLevel reads the level from the beginning of a line, which a sink has written as a fileEntry with the
// fields time, level, msg and attrs in this order.
func parseLogLevel(prefix []byte) (slog.Level, bool) {
	if !bytes.HasPrefix(prefix, []byte(`{"time":`)) {
		return 0, false
	}

	i := bytes.Index(prefix, levelKey)
	if i < 0 {
		return 0, false
	}

	rest := prefix[i+len(levelKey):]
	if len(rest) < 2 || rest[0] != '"' {
		return 0, false
	}

	end := bytes.IndexByte(rest[1:], '"')
	if end < 0 {
		return 0, false
	}

	var level slog.Level
	if err := level.UnmarshalText(rest[1 : 1+end]); err != nil {
		return 0, false
	}

	return level, true
}

// pageIndexed returns the requested page newest first and reads only the lines of the page.
func pageIndexed(f *os.File, lines []logLine, q LogQuery) LogPage {
	limit := min(max(q.Limit, 1), MaxLogPageSize)
	if q.Limit <= 0 {
		limit = MaxLogPageSize
	}
	offset := max(q.Offset, 0)

	total := 0
	for _, l := range lines {
		if l.level >= q.MinLevel {
			total++
		}
	}

	from := max(total-offset-limit, 0)
	to := total - offset
	page := LogPage{Total: total, Offset: offset}
	if to <= 0 {
		return page
	}

	var buf []byte
	i := 0
	for _, l := range lines {
		if l.level < q.MinLevel {
			continue
		}

		if i >= from {
			if cap(buf) < int(l.n) {
				buf = make([]byte, l.n)
			}
			buf = buf[:l.n]

			var e fileEntry
			if _, err := f.ReadAt(buf, l.off); err == nil && json.Unmarshal(buf, &e) == nil {
				page.Entries = append(page.Entries, e.intoEntry())
			}
		}

		i++
		if i >= to {
			break
		}
	}

	// newest first, also if concurrent writers interleaved the file order within the page
	slices.Reverse(page.Entries)
	slices.SortStableFunc(page.Entries, func(a, b LogEntry) int {
		return b.Time.Compare(a.Time)
	})

	return page
}
