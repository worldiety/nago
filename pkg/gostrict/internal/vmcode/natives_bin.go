// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"errors"
	"go/token"
	"maps"
	"math"
	"math/bits"

	"go.wdy.de/nago/pkg/gostrict/std/bin/core"
)

// A bin.Bytes is a []byte in its slot, which never changes, a bin.Reader a *core.Reader and a bin.Buffer a *core.Buffer.
// The natives use the same implementation as the native code of package bin, against which the tests compare them.

func bytesOf(v any) []byte {
	b, _ := v.([]byte)
	return b
}

func init() {
	maps.Copy(simpleNatives, binNatives)
	genericNatives[binPath+".Decode"] = binDecode
}

// binDecode is bin.Decode: the function of the program may grow the banks, so that the results go through the offsets
// of the frame instead of its windows.
func binDecode(c *compiler, targs []*tinfo, pos token.Pos) *native {
	el := targs[0].l
	return &native{name: "bin.Decode", reenters: true, fn: func(m *Machine, n []uint64, s []string, r []any) {
		atN, atS, atR := cap(m.n)-cap(n), cap(m.s)-cap(s), cap(m.r)-cap(r)
		rd := core.NewReader(bytesOf(r[0]))
		res := m.callFunc(r[1].(*Closure), func(_ []uint64, _ []string, r []any) { r[0] = rd })
		copy(m.n[atN:atN+int(el.n)], m.n[res.n:])
		copy(m.s[atS:atS+int(el.s)], m.s[res.s:])
		copy(m.r[atR:atR+int(el.r)], m.r[res.r:])
		var failure any
		if reason := rd.Close(); reason != "" {
			failure = errors.New(reason)
		}
		m.r[atR+int(el.r)] = failure
	}}
}

var binNatives = map[string]nativeFn{
	binPath + ".Empty": func(m *Machine, n []uint64, s []string, r []any) { r[0] = []byte(nil) },
	binPath + ".Of": func(m *Machine, n []uint64, s []string, r []any) {
		l := r[0].(*list)
		b := make([]byte, l.len)
		for i := range b {
			b[i] = byte(l.n[i])
		}
		r[0] = b
	},
	binPath + ".FromText": func(m *Machine, n []uint64, s []string, r []any) { r[0] = core.FromText(s[0]) },
	binPath + ".FromLatin1": func(m *Machine, n []uint64, s []string, r []any) {
		b, ok := core.FromLatin1(s[0])
		r[0], n[0] = b, fromBool(ok)
	},
	binPath + ".FromHex": func(m *Machine, n []uint64, s []string, r []any) {
		b, ok := core.FromHex(s[0])
		r[0], n[0] = b, fromBool(ok)
	},
	binPath + ".FromBase64": func(m *Machine, n []uint64, s []string, r []any) {
		b, ok := core.FromBase64(s[0])
		r[0], n[0] = b, fromBool(ok)
	},
	binPath + ".Bytes.Len": func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(len(bytesOf(r[0]))) },
	binPath + ".Bytes.Get": func(m *Machine, n []uint64, s []string, r []any) {
		b, i := bytesOf(r[0]), i32(n[0])
		if i < 0 || int(i) >= len(b) {
			n[0], n[1] = 0, 0
			return
		}
		n[0], n[1] = uint64(b[i]), 1
	},
	binPath + ".Bytes.GetOr": func(m *Machine, n []uint64, s []string, r []any) {
		b, i := bytesOf(r[0]), i32(n[0])
		if i < 0 || int(i) >= len(b) {
			n[0] = n[1]
			return
		}
		n[0] = uint64(b[i])
	},
	binPath + ".Bytes.Slice": func(m *Machine, n []uint64, s []string, r []any) {
		r[0] = core.Slice(bytesOf(r[0]), i32(n[0]), i32(n[1]))
	},
	binPath + ".Bytes.Index": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromI32(core.Index(bytesOf(r[0]), bytesOf(r[1])))
	},
	binPath + ".Bytes.Equal": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(string(bytesOf(r[0])) == string(bytesOf(r[1])))
	},
	binPath + ".Bytes.HasPrefix": func(m *Machine, n []uint64, s []string, r []any) {
		b, p := bytesOf(r[0]), bytesOf(r[1])
		n[0] = fromBool(len(p) <= len(b) && string(b[:len(p)]) == string(p))
	},
	binPath + ".Bytes.HasSuffix": func(m *Machine, n []uint64, s []string, r []any) {
		b, p := bytesOf(r[0]), bytesOf(r[1])
		n[0] = fromBool(len(p) <= len(b) && string(b[len(b)-len(p):]) == string(p))
	},
	binPath + ".Bytes.Text":       func(m *Machine, n []uint64, s []string, r []any) { s[0] = core.Text(bytesOf(r[0])) },
	binPath + ".Bytes.Latin1":     func(m *Machine, n []uint64, s []string, r []any) { s[0] = core.Latin1(bytesOf(r[0])) },
	binPath + ".Bytes.Hex":        func(m *Machine, n []uint64, s []string, r []any) { s[0] = core.Hex(bytesOf(r[0])) },
	binPath + ".Bytes.Base64":     func(m *Machine, n []uint64, s []string, r []any) { s[0] = core.Base64(bytesOf(r[0])) },
	binPath + ".Reader.U8":        func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U8()) },
	binPath + ".Reader.U16BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U16BE()) },
	binPath + ".Reader.U16LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U16LE()) },
	binPath + ".Reader.U24BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U24BE()) },
	binPath + ".Reader.U24LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U24LE()) },
	binPath + ".Reader.U32BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U32BE()) },
	binPath + ".Reader.U32LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).U32LE()) },
	binPath + ".Reader.I8":        func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).I8()) },
	binPath + ".Reader.I16BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).I16BE()) },
	binPath + ".Reader.I16LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).I16LE()) },
	binPath + ".Reader.I32BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).I32BE()) },
	binPath + ".Reader.I32LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).I32LE()) },
	binPath + ".Reader.U64BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = r[0].(*core.Reader).U64BE() },
	binPath + ".Reader.U64LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = r[0].(*core.Reader).U64LE() },
	binPath + ".Reader.I64BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).I64BE()) },
	binPath + ".Reader.I64LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(r[0].(*core.Reader).I64LE()) },
	binPath + ".Reader.F32BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(r[0].(*core.Reader).F32BE()) },
	binPath + ".Reader.F32LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(r[0].(*core.Reader).F32LE()) },
	binPath + ".Reader.F64BE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(r[0].(*core.Reader).F64BE()) },
	binPath + ".Reader.F64LE":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(r[0].(*core.Reader).F64LE()) },
	binPath + ".Reader.Bits":      func(m *Machine, n []uint64, s []string, r []any) { n[0] = r[0].(*core.Reader).Bits(uint32(n[0])) },
	binPath + ".Reader.BitsLE":    func(m *Machine, n []uint64, s []string, r []any) { n[0] = r[0].(*core.Reader).BitsLE(uint32(n[0])) },
	binPath + ".Reader.Align":     func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Reader).Align() },
	binPath + ".Reader.Skip":      func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Reader).Skip(i32(n[0])) },
	binPath + ".Reader.Seek":      func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Reader).Seek(i32(n[0])) },
	binPath + ".Reader.Pos":       func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).Pos()) },
	binPath + ".Reader.Remaining": func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Reader).Remaining()) },
	binPath + ".Reader.Bytes":     func(m *Machine, n []uint64, s []string, r []any) { r[0] = r[0].(*core.Reader).Bytes(i32(n[0])) },
	binPath + ".Reader.Rest":      func(m *Machine, n []uint64, s []string, r []any) { r[0] = r[0].(*core.Reader).Rest() },
	binPath + ".Reader.Text":      func(m *Machine, n []uint64, s []string, r []any) { s[0] = r[0].(*core.Reader).Text(i32(n[0])) },
	binPath + ".Reader.TextZ":     func(m *Machine, n []uint64, s []string, r []any) { s[0] = r[0].(*core.Reader).TextZ(i32(n[0])) },
	binPath + ".Reader.Latin1":    func(m *Machine, n []uint64, s []string, r []any) { s[0] = r[0].(*core.Reader).Latin1(i32(n[0]), false) },
	binPath + ".Reader.Latin1Z":   func(m *Machine, n []uint64, s []string, r []any) { s[0] = r[0].(*core.Reader).Latin1(i32(n[0]), true) },
	binPath + ".Reader.Fail":      func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Reader).Fail(s[0]) },
	binPath + ".Reader.Failed":    func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(r[0].(*core.Reader).Failed()) },
	binPath + ".NewBuffer":        func(m *Machine, n []uint64, s []string, r []any) { r[0] = core.NewBuffer(i32(n[0]), m.maxBuffer) },
	binPath + ".Buffer.Len":       func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*core.Buffer).Len()) },
	binPath + ".Buffer.Get": func(m *Machine, n []uint64, s []string, r []any) {
		b, i := r[0].(*core.Buffer).Bytes(), i32(n[0])
		if i < 0 || int(i) >= len(b) {
			n[0], n[1] = 0, 0
			return
		}
		n[0], n[1] = uint64(b[i]), 1
	},
	binPath + ".Buffer.GetOr": func(m *Machine, n []uint64, s []string, r []any) {
		b, i := r[0].(*core.Buffer).Bytes(), i32(n[0])
		if i < 0 || int(i) >= len(b) {
			n[0] = n[1]
			return
		}
		n[0] = uint64(b[i])
	},
	binPath + ".Buffer.Set": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).Set(i32(n[0]), uint32(n[1])))
	},
	binPath + ".Buffer.SetU16BE": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).SetU16BE(i32(n[0]), uint32(n[1])))
	},
	binPath + ".Buffer.SetU16LE": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).SetU16LE(i32(n[0]), uint32(n[1])))
	},
	binPath + ".Buffer.SetU32BE": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).SetU32BE(i32(n[0]), uint32(n[1])))
	},
	binPath + ".Buffer.SetU32LE": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).SetU32LE(i32(n[0]), uint32(n[1])))
	},
	binPath + ".Buffer.SetU64BE": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).SetU64BE(i32(n[0]), n[1]))
	},
	binPath + ".Buffer.SetU64LE": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(r[0].(*core.Buffer).SetU64LE(i32(n[0]), n[1]))
	},
	binPath + ".Buffer.AppendU8":    func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU8(uint32(n[0])) },
	binPath + ".Buffer.AppendU16BE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU16BE(uint32(n[0])) },
	binPath + ".Buffer.AppendU16LE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU16LE(uint32(n[0])) },
	binPath + ".Buffer.AppendU24BE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU24BE(uint32(n[0])) },
	binPath + ".Buffer.AppendU24LE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU24LE(uint32(n[0])) },
	binPath + ".Buffer.AppendU32BE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU32BE(uint32(n[0])) },
	binPath + ".Buffer.AppendU32LE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU32LE(uint32(n[0])) },
	binPath + ".Buffer.AppendU64BE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU64BE(n[0]) },
	binPath + ".Buffer.AppendU64LE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendU64LE(n[0]) },
	binPath + ".Buffer.AppendF32BE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendF32BE(ffrom(n[0])) },
	binPath + ".Buffer.AppendF32LE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendF32LE(ffrom(n[0])) },
	binPath + ".Buffer.AppendF64BE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendF64BE(ffrom(n[0])) },
	binPath + ".Buffer.AppendF64LE": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendF64LE(ffrom(n[0])) },
	binPath + ".Buffer.AppendBytes": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendBytes(bytesOf(r[1])) },
	binPath + ".Buffer.AppendText": func(m *Machine, n []uint64, s []string, r []any) {
		r[0].(*core.Buffer).AppendBytes(core.FromText(s[0]))
	},
	binPath + ".Buffer.AppendBits": func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).AppendBits(n[0], uint32(n[1])) },
	binPath + ".Buffer.AppendBitsLE": func(m *Machine, n []uint64, s []string, r []any) {
		r[0].(*core.Buffer).AppendBitsLE(n[0], uint32(n[1]))
	},
	binPath + ".Buffer.Reset":    func(m *Machine, n []uint64, s []string, r []any) { r[0].(*core.Buffer).Reset() },
	binPath + ".Buffer.Snapshot": func(m *Machine, n []uint64, s []string, r []any) { r[0] = r[0].(*core.Buffer).Snapshot() },
	binPath + ".Buffer.Freeze": func(m *Machine, n []uint64, s []string, r []any) {
		b, reason := r[0].(*core.Buffer).Freeze()
		if reason != "" {
			r[0], r[1] = []byte(nil), errors.New(reason)
			return
		}
		r[0], r[1] = b, nil
	},
	binPath + ".RotL32": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = uint64(bits.RotateLeft32(uint32(n[0]), int(uint32(n[1])&31)))
	},
	binPath + ".RotR32": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = uint64(bits.RotateLeft32(uint32(n[0]), -int(uint32(n[1])&31)))
	},
	binPath + ".RotL64": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = bits.RotateLeft64(n[0], int(uint32(n[1])&63))
	},
	binPath + ".RotR64": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = bits.RotateLeft64(n[0], -int(uint32(n[1])&63))
	},
	binPath + ".PopCount32":      func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.OnesCount32(uint32(n[0]))) },
	binPath + ".PopCount64":      func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.OnesCount64(n[0])) },
	binPath + ".LeadingZeros32":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.LeadingZeros32(uint32(n[0]))) },
	binPath + ".LeadingZeros64":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.LeadingZeros64(n[0])) },
	binPath + ".TrailingZeros32": func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.TrailingZeros32(uint32(n[0]))) },
	binPath + ".TrailingZeros64": func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.TrailingZeros64(n[0])) },
	binPath + ".Swap16":          func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.ReverseBytes16(uint16(n[0]))) },
	binPath + ".Swap32":          func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(bits.ReverseBytes32(uint32(n[0]))) },
	binPath + ".Swap64":          func(m *Machine, n []uint64, s []string, r []any) { n[0] = bits.ReverseBytes64(n[0]) },
	binPath + ".SignExtend32": func(m *Machine, n []uint64, s []string, r []any) {
		v, c := uint32(n[0]), uint32(n[1])
		switch {
		case c == 0:
			n[0] = 0
		case c >= 32:
			n[0] = fromI32(int32(v))
		default:
			n[0] = fromI32(int32(v<<(32-c)) >> (32 - c))
		}
	},
	binPath + ".SignExtend64": func(m *Machine, n []uint64, s []string, r []any) {
		v, c := n[0], uint32(n[1])
		switch {
		case c == 0:
			n[0] = 0
		case c >= 64:
		default:
			n[0] = uint64(int64(v<<(64-c)) >> (64 - c))
		}
	},
	binPath + ".Float32Bits": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = uint64(math.Float32bits(float32(ffrom(n[0]))))
	},
	binPath + ".Float32FromBits": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fbits(float64(math.Float32frombits(uint32(n[0]))))
	},
	binPath + ".Float64Bits":     func(m *Machine, n []uint64, s []string, r []any) {},
	binPath + ".Float64FromBits": func(m *Machine, n []uint64, s []string, r []any) {},
}
