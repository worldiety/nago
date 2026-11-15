// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"container/list"
	"errors"
	"fmt"
	"hash/maphash"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
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
// scale across CPUs, rather than serializing on a single process-wide lock. A
// shard does not hold its mutex while a file is opened, so that a slow open(2)
// does not hold up the other paths of the shard.
//
// Besides ReadAt and WriteAt, the pool reads and writes whole files, truncates,
// renames, removes and syncs them, so that an engine routes every access to a
// file through it. [FilePool.Stats] counts what all of them do.
//
// FilePool is safe for concurrent use.
type FilePool struct {
	seed   maphash.Seed
	shards []*poolShard
	mask   uint64
}

// FilePoolStats counts the work of a [FilePool] since it was created. Every
// counter only grows.
type FilePoolStats struct {
	// Opens is the number of files the pool has opened: a file it keeps open
	// counts once, a file it opens for a single operation (reading or writing a
	// whole file, syncing a directory) once per operation.
	Opens int64
	// ReadBytes is the number of bytes read from files through the pool.
	ReadBytes int64
	// WrittenBytes is the number of bytes written to files through the pool.
	WrittenBytes int64
}

// poolShard is one independent LRU segment of the pool.
type poolShard struct {
	mu      sync.Mutex
	maxOpen int
	files   map[string]*list.Element
	lru     list.List // front = most recently used, back = eviction candidate
	// gen counts the renames and removals of paths of this shard. A file, which acquire opened without holding mu, is
	// only kept when gen did not change meanwhile, since it may be the file of another name by then.
	gen uint64

	opens, readBytes, writtenBytes atomic.Int64
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
// ends. flags are passed to os.OpenFile (e.g. os.O_RDWR or os.O_RDWR|os.O_CREATE). Caller must hold s.mu; acquire
// releases it while it opens a file, so that a slow open does not hold up the other paths of the shard.
func (s *poolShard) acquire(path string, flags int) (*poolEntry, error) {
	for {
		if elem, ok := s.files[path]; ok {
			s.lru.MoveToFront(elem)
			e := elem.Value.(*poolEntry)
			e.refs++
			return e, nil
		}

		gen := s.gen
		s.mu.Unlock()
		f, err := os.OpenFile(path, flags, 0644)
		s.mu.Lock()
		if err != nil {
			return nil, err
		}
		s.opens.Add(1)
		if _, ok := s.files[path]; ok || s.gen != gen {
			// another operation opened path meanwhile, whose handle is taken, or a path of the shard was renamed or
			// removed, after which f may be the file of another name
			_ = f.Close()
			continue
		}

		// evict least recently used handles until we are below the limit
		for s.lru.Len() >= s.maxOpen {
			back := s.lru.Back()
			if back == nil {
				break
			}
			_ = s.drop(back)
		}
		entry := &poolEntry{path: path, file: f, refs: 1}
		elem := s.lru.PushFront(entry)
		s.files[path] = elem
		return entry, nil
	}
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
	p.shardFor(path).readBytes.Add(int64(n))
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
	p.shardFor(path).writtenBytes.Add(int64(n))
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

// Stats returns what the pool has done since it was created.
func (p *FilePool) Stats() FilePoolStats {
	var st FilePoolStats
	for _, s := range p.shards {
		st.Opens += s.opens.Load()
		st.ReadBytes += s.readBytes.Load()
		st.WrittenBytes += s.writtenBytes.Load()
	}
	return st
}

// ReadFile reads the whole file at path, through its handle in the pool or, if it has none, through a handle of its
// own, which it closes afterwards, so that reading a file once does not take a place in the pool. A missing file
// reports an error for which os.IsNotExist is true.
func (p *FilePool) ReadFile(path string) ([]byte, error) {
	s := p.shardFor(path)
	s.mu.Lock()
	elem, ok := s.files[path]
	if ok {
		s.lru.MoveToFront(elem)
		e := elem.Value.(*poolEntry)
		e.refs++
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			s.release(e)
			s.mu.Unlock()
		}()
		return p.readAll(s, e.file)
	}
	s.mu.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	s.opens.Add(1)
	defer f.Close()
	return p.readAll(s, f)
}

func (p *FilePool) readAll(s *poolShard, f *os.File) ([]byte, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, fi.Size())
	n, err := f.ReadAt(buf, 0)
	s.readBytes.Add(int64(n))
	if err == io.EOF {
		err = nil
	}
	return buf[:n], err
}

// WriteFile creates or truncates the file at path and writes data, through a handle of its own; a handle of path in
// the pool is dropped first.
func (p *FilePool) WriteFile(path string, data []byte) error {
	p.Evict(path)
	s := p.shardFor(path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	s.opens.Add(1)
	n, err := f.Write(data)
	s.writtenBytes.Add(int64(n))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// Truncate changes the size of the file at path through its handle.
func (p *FilePool) Truncate(path string, size int64) error {
	return p.with(path, os.O_RDWR, "truncate", func(f *os.File) error { return f.Truncate(size) })
}

// SyncAll writes the files and directories at paths to the disk, so that their data and, for a directory, its entries
// survive a power failure. It syncs them side by side, with the handle of the pool where it has one. Where a sync of
// one file does not reach through the cache of the drive (macOS, where Go's File.Sync is F_FULLFSYNC), each one is an
// fsync(2) and one F_FULLFSYNC at the end flushes the cache of the drive for all of them, so that many files cost
// about as much as one.
func (p *FilePool) SyncAll(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	errs := make([]error, len(paths))
	var next atomic.Int64
	var wg sync.WaitGroup
	for range min(len(paths), 16) {
		wg.Go(func() {
			for i := int(next.Add(1) - 1); i < len(paths); i = int(next.Add(1) - 1) {
				errs[i] = p.syncOne(paths[i], syncData)
			}
		})
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return err
	}
	if syncBarrier {
		return p.syncOne(paths[len(paths)-1], syncFull)
	}
	return nil
}

func (p *FilePool) syncOne(path string, sync func(*os.File) error) error {
	s := p.shardFor(path)
	s.mu.Lock()
	if elem, ok := s.files[path]; ok {
		s.lru.MoveToFront(elem)
		e := elem.Value.(*poolEntry)
		e.refs++
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			s.release(e)
			s.mu.Unlock()
		}()
		return sync(e.file)
	}
	s.mu.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("filepool: sync %s: %w", path, err)
	}
	s.opens.Add(1)
	err = sync(f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// Rename renames oldPath to newPath, replacing a file at newPath. It holds the shards of both paths, so that no
// operation opens either of them meanwhile, and drops their handles. Call it instead of os.Rename.
func (p *FilePool) Rename(oldPath, newPath string) error {
	a, b := p.shardFor(oldPath), p.shardFor(newPath)
	first, second := a, b
	if p.shardIndex(b) < p.shardIndex(a) {
		first, second = b, a
	}
	first.mu.Lock()
	if second != first {
		second.mu.Lock()
	}
	for _, x := range []struct {
		s    *poolShard
		path string
	}{{a, oldPath}, {b, newPath}} {
		if elem, ok := x.s.files[x.path]; ok {
			_ = x.s.drop(elem)
		}
		x.s.gen++
	}
	err := os.Rename(oldPath, newPath)
	if second != first {
		second.mu.Unlock()
	}
	first.mu.Unlock()
	return err
}

// Remove removes the file at path and drops its handle. A missing file is no error. Call it instead of os.Remove, so
// that the pool does not keep the handle of a removed file.
func (p *FilePool) Remove(path string) error {
	s := p.shardFor(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, ok := s.files[path]; ok {
		_ = s.drop(elem)
	}
	s.gen++
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *FilePool) shardIndex(s *poolShard) int {
	for i, x := range p.shards {
		if x == s {
			return i
		}
	}
	return -1
}
