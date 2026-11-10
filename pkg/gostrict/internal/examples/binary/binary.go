// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package binary reads and writes a binary format like a recorder of an industrial plant.
package binary

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/bin"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// Header is the head of a recording.
type Header struct {
	Magic   uint32
	Version uint32
	Flags   uint32
	Serial  int64
	Scale   float64
	Name    string
	Offset  int32
}

var magic = bin.Of(0x57, 0x4b, 0x44, 0x41)

// Encode returns a recording with n samples of 12 bits.
func Encode(version uint32, serial int64, scale float64, name string, n int32) (string, error) {
	b := bin.NewBuffer(16)
	b.AppendBytes(magic)
	b.AppendU16LE(version)
	b.AppendBits(5, 3)
	b.AppendBits(1, 5)
	b.AppendU64BE(uint64(serial))
	b.AppendF32BE(scale)
	b.AppendF64LE(scale)
	b.AppendText(name)
	b.AppendU8(0)
	for i := range n {
		b.AppendBits(uint64(i*37-100), 12)
	}
	b.AppendU32BE(0)
	if !b.SetU16BE(4, version+1) {
		return "", errors.New("the buffer is too small")
	}
	data, err := b.Freeze()
	if err != nil {
		return "", err
	}
	return data.Hex(), nil
}

// Decode reads a recording.
func Decode(hex string, nameLen, samples int32) (Header, int32, error) {
	data, ok := bin.FromHex(hex)
	if !ok {
		return Header{}, 0, errors.New("not hexadecimal")
	}
	sum := list.New[int32]()
	h, err := bin.Decode(data, func(r bin.Reader) Header {
		m := r.Bytes(4)
		if !m.Equal(magic) {
			r.Fail("wrong magic")
		}
		res := Header{Magic: bin.Swap32(0x01020304), Version: r.U16BE()}
		res.Flags = uint32(r.Bits(3))<<8 | uint32(r.Bits(5))
		res.Serial = r.I64BE()
		res.Scale = r.F32BE() + r.F64LE()
		res.Name = r.TextZ(nameLen)
		for range samples {
			sum.Append(bin.SignExtend32(uint32(r.Bits(12)), 12))
		}
		r.Align()
		res.Offset = r.Pos()
		return res
	})
	if err != nil {
		return Header{}, 0, err
	}
	var total int32
	for _, v := range sum.All() {
		total += v
	}
	return h, total, nil
}

// Shared returns what a concurrent function may capture.
func Text(s string) (string, int32, string, string, bool, string, bool) {
	b := bin.FromText(s)
	cut := b.Slice(1, b.Len()-1)
	l1, ok := bin.FromLatin1("Grüße")
	if !ok {
		return "", 0, "", "", false, "", false
	}
	b64 := b.Base64()
	back, ok2 := bin.FromBase64(b64)
	if !ok2 {
		return "", 0, "", "", false, "", false
	}
	return b.Text(), b.Len(), cut.Hex(), l1.Latin1(), ok, b64, back.Equal(b)
}

// Invalid returns the replacement of invalid UTF-8.
func Invalid() (string, int32, int32, bool, bool, uint32, uint32) {
	b := bin.Of(0x61, 0xff, 0xfe, 0x62, 0xe2, 0x82)
	sub := b.Index(bin.Of(0xfe, 0x62))
	missing := b.Index(bin.Of(0x01))
	v := b.GetOr(100, 7)
	w := uint32(0)
	if i := int32(3); i < b.Len() {
		w = b.GetOr(i, 0)
	}
	return b.Text(), sub, missing, b.HasPrefix(bin.Of(0x61)), b.HasSuffix(bin.Of(0x61)), v, w
}

// Bits exercises the functions of the bits.
func Bits(x uint32, y uint64, k uint32) (uint32, uint32, uint64, uint64, uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint64, uint32) {
	return bin.RotL32(x, k), bin.RotR32(x, k), bin.RotL64(y, k), bin.RotR64(y, k),
		bin.PopCount32(x), bin.PopCount64(y), bin.LeadingZeros32(x), bin.LeadingZeros64(y), bin.TrailingZeros32(x), bin.TrailingZeros64(y),
		bin.Swap32(x), bin.Swap64(y), bin.Swap16(x)
}

// Floats exercises the bits of floats and the sign extension.
func Floats(f float64, u uint32, n uint32) (uint32, float64, uint64, float64, int32, int64) {
	return bin.Float32Bits(f), bin.Float32FromBits(u), bin.Float64Bits(f), bin.Float64FromBits(bin.Float64Bits(f)), bin.SignExtend32(u, n), bin.SignExtend64(uint64(u), n)
}

// Overrun reads beyond the end.
func Overrun(n int32) (uint32, string) {
	data := bin.Of(1, 2, 3)
	v, err := bin.Decode(data, func(r bin.Reader) uint32 {
		r.Skip(n)
		return r.U32LE()
	})
	if err != nil {
		return 0, err.Error()
	}
	return v, ""
}

// LittleBits reads and writes the lowest bits first.
func LittleBits(v uint64, n uint32) (uint64, string, error) {
	b := bin.NewBuffer(0)
	b.AppendBitsLE(v, n)
	b.AppendBitsLE(5, 3)
	data, err := b.Freeze()
	if err != nil {
		return 0, "", err
	}
	got, err := bin.Decode(data, func(r bin.Reader) uint64 {
		return r.BitsLE(n)
	})
	if err != nil {
		return 0, "", err
	}
	return got, data.Base64(), nil
}

// Records writes n records of a timestamp, a channel and a measured value and sums them up again.
func Records(n int32) (int64, error) {
	b := bin.NewBuffer(n * 14)
	for i := range n {
		b.AppendU64BE(uint64(i) * 1000)
		b.AppendU16LE(uint32(i) & 7)
		b.AppendF32BE(float64(i) / 2)
	}
	data, err := b.Freeze()
	if err != nil {
		return 0, err
	}
	sum, err := bin.Decode(data, func(r bin.Reader) int64 {
		var s int64
		for range n {
			s += r.I64BE()
			s += int64(r.U16LE())
			s += int64(bin.Float32Bits(r.F32BE()) >> 31)
		}
		return s
	})
	if err != nil {
		return 0, err
	}
	return sum, nil
}
