// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"encoding/binary"
	"fmt"
	"github.com/rogpeppe/go-internal/lockedfile"
	"go.wdy.de/nago/pkg/xbytes"
	"io"
	"iter"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func OpenFile(path string) (*lockedfile.File, error) {
	//return os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	return lockedfile.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
}

// WAL is our write ahead log. it is synchronized (resp. single thread),
// so that we can re-use our buffer and become defacto gc-less. We also write without any fixed
// block size, thus we cannot reserve holes for concurrent block allocation strategies.
type WAL struct {
	f              *lockedfile.File
	tx             atomic.Uint64
	readlock       sync.Mutex
	writelock      sync.Mutex
	buf            *xbytes.Buffer
	size           atomic.Int64
	readPos        int64
	maxPayloadSize int
	debugTx        atomic.Uint64
}

func OpenWAL(path string, replay func(entry *Node)) (*WAL, error) {
	f, err := OpenFile(path)
	if err != nil {
		return nil, err
	}

	return NewWAL(f, replay)
}

// NewWAL creates a new WAL instance based on the given file. Note, that you must not issue any read/write calls from
// the replay func.
func NewWAL(f *lockedfile.File, replay func(entry *Node)) (*WAL, error) {
	w := &WAL{
		f:              f,
		buf:            &xbytes.Buffer{},
		maxPayloadSize: 64 * 1024 * 1024, // set a reasonable oom limit for keys and values
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	slog.Info("tdb WAL file opened", "size", info.Size(), "file", f.Name())

	w.size.Store(info.Size())

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	count := 0
	maxSize := 0
	totalSize := int64(0)
	var lastTx uint64
	start := time.Now()
	for node, err := range w.All() {
		if err != nil {
			return nil, err
		}

		if lastTx > node.tx {
			return nil, fmt.Errorf("tx must be monotonic")
		}

		lastTx = node.tx
		count++
		if int(node.valLength) > maxSize {
			maxSize = int(node.valLength)
		}

		totalSize += int64(node.valLength)

		if replay != nil {
			replay(node)
		}
	}

	w.tx.Store(lastTx)

	slog.Info("tdb WAL replay complete", "entries", count, "max-value-size", maxSize, "total-value-size", totalSize, "avg-value-size", float64(totalSize)/float64(count), "duration", time.Since(start))

	return w, nil
}

func (w *WAL) read(entry *Node) error {
	startOfEntryInFile := w.readPos

	var header [1 + 4]byte // kind+size
	if _, err := w.f.Read(header[:]); err != nil {
		return err
	}

	plen := binary.LittleEndian.Uint32(header[1:])

	if int(plen) > w.maxPayloadSize {
		return fmt.Errorf("payload too large: %d > %d", plen, w.maxPayloadSize)
	}

	fullSize := int(plen) + 9 // kind + len + crc
	if cap(w.buf.Buf) < fullSize {
		w.buf.Buf = make([]byte, fullSize)
	}

	if len(w.buf.Buf) < fullSize {
		w.buf.Buf = w.buf.Buf[:fullSize]
	}

	w.buf.Reset()
	_, _ = w.buf.Write(header[:])
	w.buf.Reset()
	tmp := w.buf.Buf[len(header):fullSize]
	if n, err := w.f.Read(tmp); err != nil || n != (fullSize-len(header)) {
		if err != nil {
			return err
		}

		return io.EOF
	}

	if err := entry.read(w.buf); err != nil {
		w.readPos += int64(w.buf.Pos)
		return err
	}

	w.readPos += int64(w.buf.Pos)
	entry.f = w.f
	entry.valOffset += uint64(startOfEntryInFile)

	return nil
}

// write does not require that the filepointer is at the end, because we calculate the offset directly and use
// pwrite for appending.
func (w *WAL) write(entry *Node) (int, error) {
	w.writelock.Lock()
	defer w.writelock.Unlock()

	startOfNodeInFile := uint64(w.size.Load())
	w.buf.Reset()
	entry.write(w.buf)
	buf := w.buf.Buf[:w.buf.Pos]

	n, err := w.f.WriteAt(buf, w.size.Load())
	if err != nil {
		// we probably screwed things up, if we have run out of space, truncate to be consistent with next read/write
		if err := w.f.Truncate(w.size.Load()); err != nil {
			slog.Error("write to WAL failed and unable to truncate", "err", err)
		}

		return n, err
	}

	entry.f = w.f
	entry.valOffset += startOfNodeInFile
	w.size.Add(int64(n))

	return n, nil
}

// All iterates over all entries. You must not nest calls to All, because it will cause a deadlock. This is
// because we have a single file cursor. We may solve that, if we introduce fixed size blocks because we then
// can use concurrent pread calls for everything.
// You must neither keep the Node nor the associated data, because they are re-used throughout the iteration.
// However, writes are always accepted and appended concurrently while iterating and become visible as the last Node.
func (w *WAL) All() iter.Seq2[*Node, error] {
	return func(yield func(*Node, error) bool) {
		w.readlock.Lock()
		defer w.readlock.Unlock()

		_, err := w.f.Seek(0, io.SeekStart)
		if err != nil {
			if err := w.Close(); err != nil {
				slog.Error("failed to seek to start-of-file for All in WAL, closing failed", "suppressed", err)
			}
			yield(nil, err)
			return
		}

		w.readPos = 0

		var entry Node
		// always loop and check the consistency of the entire WAL
		for {
			err := w.read(&entry)
			if err != nil {
				if err == io.EOF {
					break
				}
				if !yield(nil, err) {
					break
				}
			}

			if !yield(&entry, nil) {
				break
			}
		}

		// we don't need to reset the filepointer, because write does not it

	}
}

func (w *WAL) SetWithTx(bucket, key, value []byte, tx uint64) (Value, error) {
	if w.debugTx.Load() > tx {
		panic("debug tx is smaller than tx to write")
	}
	w.debugTx.Store(tx)

	n := Node{
		kind:   setKeyValue,
		tx:     tx,
		bucket: bucket,
		key:    key,
		val:    value,
	}

	_, err := w.write(&n)
	return n.Value(), err
}

func (w *WAL) DeleteWithTx(bucket, key []byte, tx uint64) error {
	if w.debugTx.Load() > tx {
		panic("debug tx is smaller than tx to write")
	}
	w.debugTx.Store(tx)

	n := Node{
		kind:   removeKeyValue,
		tx:     tx,
		bucket: bucket,
		key:    key,
	}

	_, err := w.write(&n)
	return err
}

func (w *WAL) Close() error {
	if err := w.f.Sync(); err != nil {
		return err
	}

	return w.f.Close()
}
