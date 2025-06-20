// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

//protonc:embed below

type BinaryWriter struct {
	writer *bytes.Buffer
	tmp    [32]byte
}

func NewBinaryWriter(writer *bytes.Buffer) *BinaryWriter {
	return &BinaryWriter{
		writer: writer,
	}
}

func (w *BinaryWriter) write(p []byte) error {
	_, err := w.writer.Write(p)
	return err
}

func (w *BinaryWriter) writeBool(b bool) {

	if b {
		w.tmp[0] = 1
	} else {
		w.tmp[0] = 0
	}

	w.write(w.tmp[0:1])
}

func (w *BinaryWriter) writeVarint(i int64) error {
	n := binary.PutVarint(w.tmp[:], i)
	w.write(w.tmp[0:n])
	return nil
}

func (w *BinaryWriter) writeUvarint(i uint64) error {
	n := binary.PutUvarint(w.tmp[:], i)
	w.write(w.tmp[0:n])
	return nil
}

func (w *BinaryWriter) writeByte(b byte) error {
	return w.writer.WriteByte(b)
}

func (w *BinaryWriter) writeFieldHeader(shape shape, id fieldId) error {
	return w.writeByte(fieldHeader{
		shape:   shape,
		fieldId: id,
	}.asValue())
}

func (w *BinaryWriter) writeTypeHeader(shape shape, id typeId) error {
	if err := w.writeFieldHeader(shape, 0); err != nil {
		return err
	}

	return w.writeUvarint(uint64(id))
}

func (w *BinaryWriter) writeSlice(s []byte) error {
	n := len(s)
	if err := w.writeUvarint(uint64(n)); err != nil {
		return err
	}
	return w.write(s)
}

func (w *BinaryWriter) writeFloat64(v float64) error {
	binary.LittleEndian.PutUint64(w.tmp[:8], math.Float64bits(v))
	return w.write(w.tmp[:8])
}

type BinaryReader struct {
	reader *bytes.Buffer
	tmp    [32]byte
}

func NewBinaryReader(reader *bytes.Buffer) *BinaryReader {
	return &BinaryReader{
		reader: reader,
	}
}

func (r *BinaryReader) read(b []byte) error {
	n, err := r.reader.Read(b)
	if err != nil {
		return err
	}

	if n != len(b) {
		return fmt.Errorf("short read")
	}

	return nil
}

func (r *BinaryReader) readByte() (byte, error) {
	return r.reader.ReadByte()
}

func (r *BinaryReader) readFieldHeader() (fieldHeader, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return fieldHeader{}, err
	}

	return parseFieldHeader(b), nil
}

func (r *BinaryReader) readTypeHeader() (shape, typeId, error) {
	h, err := r.readFieldHeader()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read type header: %w", err)
	}

	if h.isField() {
		return 0, 0, fmt.Errorf("nprotoc: expected a type header but got a field header")
	}

	tid, err := r.readUvarint()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read type id header: %w", err)
	}

	return h.shape, typeId(tid), nil
}

func (r *BinaryReader) readUvarint() (uint64, error) {
	return binary.ReadUvarint(r.reader)
}

func (r *BinaryReader) readVarint() (int64, error) {
	return binary.ReadVarint(r.reader)
}

func (r *BinaryReader) readFloat64() (float64, error) {
	if err := r.read(r.tmp[:8]); err != nil {
		return 0.0, err
	}

	value := binary.LittleEndian.Uint64(r.tmp[:8])

	return math.Float64frombits(value), nil
}

type shape uint8

func (s shape) String() string {
	switch s {
	case f32:
		return "f32"
	case f64:
		return "f64"
	case envelope:
		return "envelope"
	case uvarint:
		return "uvarint"
	case varint:
		return "varint"
	case byteSlice:
		return "byteSlice"
	case record:
		return "record"
	case array:
		return "array"
	case xobjectAsArray:
		return "xobjectAsArray"
	case xbool:
		return "xbool"
	case xmap:
		return "xmap"
	}

	panic(fmt.Sprintf("unknown shape: %d", s))
}

const (
	envelope = shape(iota)
	uvarint
	varint
	byteSlice
	record
	f32
	f64
	array
	xobjectAsArray
	xbool
	xmap
)

type fieldId uint

type fieldHeader struct {
	shape   shape
	fieldId fieldId
}

func (f fieldHeader) isField() bool {
	return f.fieldId != 0
}

func (f fieldHeader) asValue() uint8 {
	return uint8(((int(f.shape)) << 5) | ((int(f.fieldId)) & 0b00011111))
}

func parseFieldHeader(value uint8) fieldHeader {
	return fieldHeader{
		shape:   shape((value >> 5) & 0b00000111),
		fieldId: fieldId(value & 0b00011111),
	}
}

type typeId uint

type typeHeader struct {
	shape   shape
	fieldId fieldId
	typeId  typeId
}

func (f typeHeader) isType() bool {
	return f.fieldId == 0
}

func parseTypeHeader(value uint8) fieldHeader {
	return fieldHeader{
		shape:   shape((value >> 5) & 0b00000111),
		fieldId: fieldId(value & 0b00011111),
	}
}
