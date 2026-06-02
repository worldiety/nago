package msgstore

import (
	"container/list"
	"fmt"
	"os"
	"sync"
)

// FilePool manages a bounded pool of open file descriptors with LRU eviction.
// All file I/O in the store is routed through the pool to prevent unbounded
// accumulation of open file handles (e.g. during a full replay across millions
// of segment files).
type FilePool struct {
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
// is closed before a new one is opened.
func NewFilePool(maxOpen int) *FilePool {
	if maxOpen <= 0 {
		maxOpen = 1024
	}
	return &FilePool{
		maxOpen: maxOpen,
		files:   make(map[string]*list.Element),
	}
}

// get returns a cached or freshly opened *os.File for path.
// flags are passed to os.OpenFile (e.g. os.O_RDWR or os.O_RDWR|os.O_CREATE).
// Caller must hold p.mu.
func (p *FilePool) get(path string, flags int) (*os.File, error) {
	if elem, ok := p.files[path]; ok {
		p.lru.MoveToFront(elem)
		return elem.Value.(*poolEntry).file, nil
	}

	// evict least recently used handles until we are below the limit
	for p.lru.Len() >= p.maxOpen {
		back := p.lru.Back()
		if back == nil {
			break
		}
		evict := back.Value.(*poolEntry)
		evict.file.Close()
		delete(p.files, evict.path)
		p.lru.Remove(back)
	}

	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, err
	}

	entry := &poolEntry{path: path, file: f}
	elem := p.lru.PushFront(entry)
	p.files[path] = elem
	return f, nil
}

// ReadAt reads len(buf) bytes from the file at path starting at byte offset off.
// The file is opened on first access and kept in the pool for reuse.
func (p *FilePool) ReadAt(path string, buf []byte, off int64) (int, error) {
	p.mu.Lock()
	f, err := p.get(path, os.O_RDWR)
	p.mu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("filepool: open %s for read: %w", path, err)
	}
	return f.ReadAt(buf, off)
}

// WriteAt writes len(buf) bytes to the file at path starting at byte offset off.
// The file is created if it does not exist.
func (p *FilePool) WriteAt(path string, buf []byte, off int64) (int, error) {
	p.mu.Lock()
	f, err := p.get(path, os.O_RDWR|os.O_CREATE)
	p.mu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("filepool: open %s for write: %w", path, err)
	}
	return f.WriteAt(buf, off)
}

// Stat returns the FileInfo for the file at path.
// Uses the pooled handle (fstat) if available, otherwise falls back to os.Stat.
func (p *FilePool) Stat(path string) (os.FileInfo, error) {
	p.mu.Lock()
	if elem, ok := p.files[path]; ok {
		p.lru.MoveToFront(elem)
		f := elem.Value.(*poolEntry).file
		p.mu.Unlock()
		return f.Stat()
	}
	p.mu.Unlock()
	return os.Stat(path)
}

// Evict closes and removes the file at path from the pool.
// This must be called before renaming or deleting a file so that stale
// handles do not remain in the pool.
func (p *FilePool) Evict(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if elem, ok := p.files[path]; ok {
		elem.Value.(*poolEntry).file.Close()
		delete(p.files, path)
		p.lru.Remove(elem)
	}
}

// Close closes all open file handles in the pool.
func (p *FilePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for _, elem := range p.files {
		if err := elem.Value.(*poolEntry).file.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.files = make(map[string]*list.Element)
	p.lru.Init()
	return firstErr
}

