package tdb

import (
	"errors"
	"go.wdy.de/nago/pkg/xbytes"
	"hash/crc32"
	"io"
	"os"
)

var InvalidNodeType = errors.New("invalid Node type")
var InvalidNodeCRC = errors.New("invalid Node crc")

type nodeKind uint8

const (
	setKeyValue nodeKind = iota + 1
	removeKeyValue
)

type Node struct {
	f         *os.File
	kind      nodeKind
	size      uint32
	crc       uint32
	tx        uint64
	bucket    []byte
	key       []byte
	valOffset uint64
	valLength uint32
	val       []byte
}

func (e *Node) Value() ValPtr {
	return ValPtr{
		f:      e.f,
		offset: int64(e.valOffset),
		len:    e.valLength,
	}
}

func (e *Node) write(w *xbytes.Buffer) {
	// we know, that xbytes.Buffer cannot fail for write
	_ = w.WriteByte(byte(e.kind))
	startOfLen := w.Pos
	_ = w.WriteUint32(0) // write dummy length
	_ = w.WriteUint32(0) // write dummy crc
	startOfPayload := w.Pos
	_, _ = w.WriteUvarint(e.tx)
	_, _ = w.WriteSlice(e.bucket)
	_, _ = w.WriteSlice(e.key)
	if e.kind == setKeyValue {
		e.valLength = uint32(len(e.val))
		_, _ = w.WriteSlice(e.val)
		e.valOffset = uint64(w.Pos - len(e.val))
	} else {
		e.valLength = 0
		e.valOffset = 0
	}

	endOfWrite := w.Pos
	e.crc = crc32.ChecksumIEEE(w.Buf[startOfPayload:w.Pos])

	// update length and crc
	w.Pos = startOfLen
	e.size = uint32(endOfWrite - startOfPayload)
	_ = w.WriteUint32(e.size)
	_ = w.WriteUint32(e.crc)

	w.Pos = endOfWrite
}

func (e *Node) read(r *xbytes.Buffer) error {
	kind, err := r.ReadByte()
	if err != nil {
		return err // this may be just EOF
	}

	e.kind = nodeKind(kind)
	if e.kind != setKeyValue && e.kind != removeKeyValue {
		return InvalidNodeType
	}

	e.size, _ = r.ReadUint32()
	e.crc, _ = r.ReadUint32()
	startOfPayload := r.Pos

	if startOfPayload+int(e.size) > len(r.Buf) {
		// well, we obviously are already screwed
		return io.ErrShortBuffer
	}

	crc := crc32.ChecksumIEEE(r.Buf[startOfPayload : startOfPayload+int(e.size)])
	if e.crc != crc {
		// don't try to parse a broken Node at all, probably all lengths are rubbish anyway
		return InvalidNodeCRC
	}

	e.tx, _ = r.ReadUvarint()
	e.bucket, _ = r.ReadInto(e.bucket)
	e.key, _ = r.ReadInto(e.key)
	if e.kind == setKeyValue {
		e.val, _ = r.ReadInto(e.val)
		e.valOffset = uint64(r.Pos - len(e.val))
		e.valLength = uint32(len(e.val))
	} else {
		e.valLength = 0
		e.valOffset = 0
	}

	return r.Err // all errors on the buffer are short-circuit
}
