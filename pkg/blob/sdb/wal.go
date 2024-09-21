package sdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"sync"
	"sync/atomic"
)

type logEntry struct {
	kind uint8
	tx   uint64
	key  []byte
	val  []byte
}

func (e *logEntry) write(w *bytes.Buffer) {
	var tmp [12]byte

	// kind
	w.WriteByte(e.kind)

	//tx
	l := binary.PutUvarint(tmp[:], e.tx)
	w.Write(tmp[:l])

	//key len
	l = binary.PutUvarint(tmp[:], uint64(len(e.key)))
	w.Write(tmp[:l])

	// key bytes
	w.Write(e.key)

	if e.kind == walEntrySet {
		// val len
		l = binary.PutUvarint(tmp[:], uint64(len(e.val)))
		w.Write(tmp[:l])

		// val bytes
		w.Write(e.val)
	}
}

func (e *logEntry) read(r *bytes.Buffer) error {
	// kind
	t, err := r.ReadByte()
	if err != nil {
		return err
	}
	e.kind = t

	//tx
	tx, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	e.tx = tx

	//key len
	l, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}

	if cap(e.key) < int(l) {
		e.key = make([]byte, l)
	} else {
		e.key = e.key[:l]
	}

	// key data
	if n, err := r.Read(e.key); err != nil || n != int(l) {
		return fmt.Errorf("cannot read key: %w", err)
	}

	////

	if e.kind == walEntrySet {
		// val len
		l, err = binary.ReadUvarint(r)
		if err != nil {
			return err
		}

		if cap(e.val) < int(l) {
			e.val = make([]byte, l)
		} else {
			e.val = e.val[:l]
		}

		// val data
		if n, err := r.Read(e.val); err != nil || n != int(l) {
			return fmt.Errorf("cannot read val: %w", err)
		}
	}

	return nil
}

const (
	walEntrySet = iota + 1
	walEntryRem
)

// wal is our write ahead log. it is synchronized, so that we can re-use our buffer and become defacto
// gc-less.
type wal struct {
	f     *os.File
	tx    atomic.Uint64
	mutex sync.Mutex
	buf   *bytes.Buffer
}

func newWal(f *os.File, replay func()) *wal {
	return &wal{
		f:   f,
		buf: &bytes.Buffer{},
	}
}

func (w *wal) read(entry *logEntry) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	var header [8]byte
	if _, err := w.f.Read(header[:]); err != nil {
		return err
	}

	crc := binary.LittleEndian.Uint32(header[:4])
	plen := binary.LittleEndian.Uint32(header[4:])

	w.buf.Reset()
	w.buf.Grow(int(plen))
	buf := w.buf.Bytes()
	if _, err := w.f.Read(buf[:int(plen)]); err != nil {
		return err
	}

	checkCRC := crc32.ChecksumIEEE(buf[:int(plen)])
	//fmt.Printf("read checkCRC=%d crc=%d len=%d\n", checkCRC, crc, plen)

	if crc != checkCRC {
		return fmt.Errorf("invalid CRC")
	}

	return entry.read(bytes.NewBuffer(buf[:int(plen)]))

}

func (w *wal) write(entry logEntry) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	var entryLen [4]byte
	w.buf.Reset()
	w.buf.Write(entryLen[:]) // 4 byte crc32 of payload
	w.buf.Write(entryLen[:]) // 4 byte payload len
	entry.write(w.buf)       // payload

	buf := w.buf.Bytes()

	binary.LittleEndian.PutUint32(buf[:4], crc32.ChecksumIEEE(buf[8:]))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(buf)-8))

	//fmt.Printf("crc=%d len=%d\n", crc32.ChecksumIEEE(buf[8:]), len(buf)-8)

	return w.f.Write(buf)
}

func (w *wal) Close() error {
	if err := w.f.Sync(); err != nil {
		return err
	}

	return w.f.Close()
}
