// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"container/list"
	"fmt"
	"hash/maphash"
	"os"
	"runtime"
	"sync"
)

// FilePool manages a bounded pool of open file descriptors with LRU eviction.
// All file I/O in an engine should be routed through the pool to prevent
// unbounded accumulation of open file handles (e.g. during a full replay across
// millions of segment files).
//
// A single FilePool is owned by a [DB] and shared by every engine instance it
// opens, so the process holds one bounded set of descriptors regardless of how
// many engines are active. Engines receive the shared pool through their
// [EngineFactory] and must never close it: its lifecycle belongs to the DB.
//
// The pool is sharded by path hash into independent segments, each with its own
// mutex and LRU. Operations on different paths (e.g. different columns of a time
// series, or different segment files) therefore contend on different mutexes and
// scale across CPUs, rather than serializing on a single process-wide lock.
//
// FilePool is safe for concurrent use.
type FilePool struct {
	seed   maphash.Seed
	shards []*poolShard
	mask   uint64
}

// poolShard is one independent LRU segment of the pool.
type poolShard struct {
	mu      sync.Mutex
	maxOpen int
	files   map[string]*list.Element
	lru     list.List // front = most recently used, back = eviction candidate
}

type poolEntry struct {
	path string
	file *os.File
	// refs counts the reads, writes and stats running on file. An entry, which leaves the pool while refs > 0, is
	// marked evicted, and the last of these operations closes the file, so that no operation finds its handle closed by
	// another path of the same shard. The pool may therefore hold more open files than maxOpen for as long as such
	// operations run, at most one per operation.
	refs    int
	evicted bool
}

// NewFilePool creates a FilePool that keeps at most maxOpen files open
// simultaneously. When the limit is reached, the least recently used handle
// is closed before a new one is opened. The limit is distributed evenly across
// internal shards.
func NewFilePool(maxOpen int) *FilePool {
	if maxOpen <= 0 {
		maxOpen = 1024
	}

	// Choose a power-of-two shard count scaled to the machine, capped so each
	// shard still keeps a reasonable share of the descriptor budget.
	n := nextPow2(runtime.GOMAXPROCS(0) * 4)
	if n < 1 {
		n = 1
	}
	for n > 1 && maxOpen/n < 8 {
		n /= 2
	}

	perShard := maxOpen / n
	if perShard < 1 {
		perShard = 1
	}

	p := &FilePool{
		seed:   maphash.MakeSeed(),
		shards: make([]*poolShard, n),
		mask:   uint64(n - 1),
	}
	for i := range p.shards {
		p.shards[i] = &poolShard{
			maxOpen: perShard,
			files:   make(map[string]*list.Element),
		}
	}
	return p
}

func nextPow2(v int) int {
	n := 1
	for n < v {
		n <<= 1
	}
	return n
}

// shardFor returns the shard owning path. A given path always maps to the same
// shard so its single open handle is never duplicated across shards.
func (p *FilePool) shardFor(path string) *poolShard {
	var h maphash.Hash
	h.SetSeed(p.seed)
	_, _ = h.WriteString(path)
	return p.shards[h.Sum64()&p.mask]
}

// acquire returns the entry of path, cached or freshly opened, with a reference for one operation, which release
// ends. flags are passed to os.OpenFile (e.g. os.O_RDWR or os.O_RDWR|os.O_CREATE). Caller must hold s.mu.
func (s *poolShard) acquire(path string, flags int) (*poolEntry, error) {
	if elem, ok := s.files[path]; ok {
		s.lru.MoveToFront(elem)
		e := elem.Value.(*poolEntry)
		e.refs++
		return e, nil
	}

	// evict least recently used handles until we are below the limit
	for s.lru.Len() >= s.maxOpen {
		back := s.lru.Back()
		if back == nil {
			break
		}
		_ = s.drop(back)
	}

	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, err
	}

	entry := &poolEntry{path: path, file: f, refs: 1}
	elem := s.lru.PushFront(entry)
	s.files[path] = elem
	return entry, nil
}

// drop removes elem from the pool and closes its file, or leaves the close to the last running operation on it.
// Caller must hold s.mu.
func (s *poolShard) drop(elem *list.Element) error {
	e := elem.Value.(*poolEntry)
	delete(s.files, e.path)
	s.lru.Remove(elem)
	if e.refs > 0 {
		e.evicted = true
		return nil
	}
	return e.file.Close()
}

// release ends the operation of acquire and closes the file of an entry, which left the pool meanwhile, after its last
// operation. Caller must hold s.mu.
func (s *poolShard) release(e *poolEntry) {
	e.refs--
	if e.refs == 0 && e.evicted {
		_ = e.file.Close()
	}
}

// with runs op on the file of path, which the pool does not close until op returns.
func (p *FilePool) with(path string, flags int, what string, op func(f *os.File) error) error {
	s := p.shardFor(path)
	s.mu.Lock()
	e, err := s.acquire(path, flags)
	s.mu.Unlock()
	if err != nil {
		return fmt.Errorf("filepool: open %s for %s: %w", path, what, err)
	}
	defer func() {
		s.mu.Lock()
		s.release(e)
		s.mu.Unlock()
	}()
	return op(e.file)
}

// ReadAt reads len(buf) bytes from the file at path starting at byte offset off.
// The file is opened on first access and kept in the pool for reuse.
func (p *FilePool) ReadAt(path string, buf []byte, off int64) (int, error) {
	var n int
	err := p.with(path, os.O_RDWR, "read", func(f *os.File) error {
		var err error
		n, err = f.ReadAt(buf, off)
		return err
	})
	return n, err
}

// WriteAt writes len(buf) bytes to the file at path starting at byte offset off.
// The file is created if it does not exist.
func (p *FilePool) WriteAt(path string, buf []byte, off int64) (int, error) {
	var n int
	err := p.with(path, os.O_RDWR|os.O_CREATE, "write", func(f *os.File) error {
		var err error
		n, err = f.WriteAt(buf, off)
		return err
	})
	return n, err
}

// Stat returns the FileInfo for the file at path.
// Uses the pooled handle (fstat) if available, otherwise falls back to os.Stat.
func (p *FilePool) Stat(path string) (os.FileInfo, error) {
	s := p.shardFor(path)
	s.mu.Lock()
	elem, ok := s.files[path]
	if !ok {
		s.mu.Unlock()
		return os.Stat(path)
	}
	s.lru.MoveToFront(elem)
	e := elem.Value.(*poolEntry)
	e.refs++
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.release(e)
		s.mu.Unlock()
	}()
	return e.file.Stat()
}

// Evict closes and removes the file at path from the pool; a read or write running on it closes it when it ends.
// This must be called before renaming or deleting a file so that stale
// handles do not remain in the pool.
func (p *FilePool) Evict(path string) {
	s := p.shardFor(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, ok := s.files[path]; ok {
		_ = s.drop(elem)
	}
}

// Close closes all open file handles in the pool.
func (p *FilePool) Close() error {
	var firstErr error
	for _, s := range p.shards {
		s.mu.Lock()
		for _, elem := range s.files {
			if err := s.drop(elem); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		s.mu.Unlock()
	}
	return firstErr
}
