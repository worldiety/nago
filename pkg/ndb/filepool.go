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

// get returns a cached or freshly opened *os.File for path.
// flags are passed to os.OpenFile (e.g. os.O_RDWR or os.O_RDWR|os.O_CREATE).
// Caller must hold s.mu.
func (s *poolShard) get(path string, flags int) (*os.File, error) {
	if elem, ok := s.files[path]; ok {
		s.lru.MoveToFront(elem)
		return elem.Value.(*poolEntry).file, nil
	}

	// evict least recently used handles until we are below the limit
	for s.lru.Len() >= s.maxOpen {
		back := s.lru.Back()
		if back == nil {
			break
		}
		evict := back.Value.(*poolEntry)
		evict.file.Close()
		delete(s.files, evict.path)
		s.lru.Remove(back)
	}

	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, err
	}

	entry := &poolEntry{path: path, file: f}
	elem := s.lru.PushFront(entry)
	s.files[path] = elem
	return f, nil
}

// ReadAt reads len(buf) bytes from the file at path starting at byte offset off.
// The file is opened on first access and kept in the pool for reuse.
func (p *FilePool) ReadAt(path string, buf []byte, off int64) (int, error) {
	s := p.shardFor(path)
	s.mu.Lock()
	f, err := s.get(path, os.O_RDWR)
	s.mu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("filepool: open %s for read: %w", path, err)
	}
	return f.ReadAt(buf, off)
}

// WriteAt writes len(buf) bytes to the file at path starting at byte offset off.
// The file is created if it does not exist.
func (p *FilePool) WriteAt(path string, buf []byte, off int64) (int, error) {
	s := p.shardFor(path)
	s.mu.Lock()
	f, err := s.get(path, os.O_RDWR|os.O_CREATE)
	s.mu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("filepool: open %s for write: %w", path, err)
	}
	return f.WriteAt(buf, off)
}

// Stat returns the FileInfo for the file at path.
// Uses the pooled handle (fstat) if available, otherwise falls back to os.Stat.
func (p *FilePool) Stat(path string) (os.FileInfo, error) {
	s := p.shardFor(path)
	s.mu.Lock()
	if elem, ok := s.files[path]; ok {
		s.lru.MoveToFront(elem)
		f := elem.Value.(*poolEntry).file
		s.mu.Unlock()
		return f.Stat()
	}
	s.mu.Unlock()
	return os.Stat(path)
}

// Evict closes and removes the file at path from the pool.
// This must be called before renaming or deleting a file so that stale
// handles do not remain in the pool.
func (p *FilePool) Evict(path string) {
	s := p.shardFor(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, ok := s.files[path]; ok {
		elem.Value.(*poolEntry).file.Close()
		delete(s.files, path)
		s.lru.Remove(elem)
	}
}

// Close closes all open file handles in the pool.
func (p *FilePool) Close() error {
	var firstErr error
	for _, s := range p.shards {
		s.mu.Lock()
		for _, elem := range s.files {
			if err := elem.Value.(*poolEntry).file.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		s.files = make(map[string]*list.Element)
		s.lru.Init()
		s.mu.Unlock()
	}
	return firstErr
}
