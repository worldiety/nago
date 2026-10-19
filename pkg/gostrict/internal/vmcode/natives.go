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
	"math"
	"math/big"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// native is a function of the standard library or of a package of the subset, which the VM implements in Go. It
// finds its arguments at the beginning of the windows of the banks, and writes its results over them.
type native struct {
	name string
	fn   func(m *Machine, n []uint64, s []string, r []any)
	// inline emits the function as instructions, e.g. the length of a list
	inline func(g *gen, dst layout, args []layout)
}

type nativeFn = func(m *Machine, n []uint64, s []string, r []any)

// native returns the implementation of a function of a package, which is not in the subset, by its full name, or
// nil.
func (c *compiler) native(key string, targs []*tinfo, pos token.Pos) *native {
	k := key
	for _, t := range targs {
		k += "," + t.String()
	}
	if nat, ok := c.natives[k]; ok {
		return nat
	}
	var nat *native
	if fn, ok := simpleNatives[key]; ok {
		nat = &native{name: key, fn: fn}
	} else if mk, ok := genericNatives[key]; ok {
		nat = mk(c, targs, pos)
	} else if f, ok := mathFuncs[key]; ok {
		nat = &native{name: key, fn: func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(f(ffrom(n[0]))) }}
	} else if f, ok := mathFuncs2[key]; ok {
		nat = &native{name: key, fn: func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(f(ffrom(n[0]), ffrom(n[1]))) }}
	} else {
		return nil
	}
	c.natives[k] = nat
	return nat
}

func fbits(f float64) uint64 { return math.Float64bits(f) }
func ffrom(u uint64) float64 { return math.Float64frombits(u) }
func i32(u uint64) int32     { return int32(u) }
func fromI32(v int32) uint64 { return uint64(int64(v)) }
func fromBool(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

// A big.Int of std/big is a *math/big.Int in its slot, which never changes. The functions follow the native
// implementation of std/big, against which TestBig checks them.

// maxPow10 is big.MaxPow10.
const maxPow10 = 4096

var (
	bigZero = new(big.Int)
	digits  = regexp.MustCompile(`^[+-]?[0-9]+$`)
)

func bigOf(v any) *big.Int {
	if b, ok := v.(*big.Int); ok {
		return b
	}
	return bigZero
}

func bigString(v any) string {
	return bigOf(v).String()
}

func parseBig(s string) (*big.Int, bool) {
	if !digits.MatchString(s) {
		return bigZero, false
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return bigZero, false
	}
	return v, true
}

// bigParse returns the integer of the neutral value of a big.Int, 0 for an invalid one.
func bigParse(s string) any {
	v, _ := parseBig(s)
	return v
}

var mathFuncs = map[string]func(float64) float64{
	"math.Abs": math.Abs, "math.Acos": math.Acos, "math.Asin": math.Asin, "math.Atan": math.Atan, "math.Cbrt": math.Cbrt,
	"math.Ceil": math.Ceil, "math.Cos": math.Cos, "math.Exp": math.Exp, "math.Floor": math.Floor, "math.Log": math.Log,
	"math.Log10": math.Log10, "math.Log2": math.Log2, "math.Round": math.Round, "math.Sin": math.Sin,
	"math.Sqrt": math.Sqrt, "math.Tan": math.Tan, "math.Trunc": math.Trunc,
}

var mathFuncs2 = map[string]func(float64, float64) float64{
	"math.Atan2": math.Atan2, "math.Copysign": math.Copysign, "math.Hypot": math.Hypot, "math.Max": math.Max,
	"math.Min": math.Min, "math.Mod": math.Mod, "math.Pow": math.Pow,
}

func newStrings(vs []string) *list {
	return &list{s: vs, len: int32(len(vs))}
}

var simpleNatives = map[string]nativeFn{
	"math.Inf":   func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(math.Inf(int(i32(n[0])))) },
	"math.NaN":   func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(math.NaN()) },
	"math.IsNaN": func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(math.IsNaN(ffrom(n[0]))) },
	"math.IsInf": func(m *Machine, n []uint64, s []string, r []any) {
		n[0] = fromBool(math.IsInf(ffrom(n[0]), int(i32(n[1]))))
	},
	"math.Signbit": func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(math.Signbit(ffrom(n[0]))) },

	"strings.Contains":   func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(strings.Contains(s[0], s[1])) },
	"strings.HasPrefix":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(strings.HasPrefix(s[0], s[1])) },
	"strings.HasSuffix":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(strings.HasSuffix(s[0], s[1])) },
	"strings.ReplaceAll": func(m *Machine, n []uint64, s []string, r []any) { s[0] = strings.ReplaceAll(s[0], s[1], s[2]) },
	"strings.TrimPrefix": func(m *Machine, n []uint64, s []string, r []any) { s[0] = strings.TrimPrefix(s[0], s[1]) },
	"strings.TrimSuffix": func(m *Machine, n []uint64, s []string, r []any) { s[0] = strings.TrimSuffix(s[0], s[1]) },
	"strings.TrimSpace":  func(m *Machine, n []uint64, s []string, r []any) { s[0] = strings.TrimSpace(s[0]) },

	"errors.New":  func(m *Machine, n []uint64, s []string, r []any) { r[0] = errors.New(s[0]) },
	"error.Error": func(m *Machine, n []uint64, s []string, r []any) { s[0] = r[0].(error).Error() },

	"min.int32":   func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(min(i32(n[0]), i32(n[1]))) },
	"max.int32":   func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(max(i32(n[0]), i32(n[1]))) },
	"min.uint32":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(min(uint32(n[0]), uint32(n[1]))) },
	"max.uint32":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(max(uint32(n[0]), uint32(n[1]))) },
	"min.float64": func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(min(ffrom(n[0]), ffrom(n[1]))) },
	"max.float64": func(m *Machine, n []uint64, s []string, r []any) { n[0] = fbits(max(ffrom(n[0]), ffrom(n[1]))) },
	"min.string":  func(m *Machine, n []uint64, s []string, r []any) { s[0] = min(s[0], s[1]) },
	"max.string":  func(m *Machine, n []uint64, s []string, r []any) { s[0] = max(s[0], s[1]) },

	textPath + ".RuneCount":     func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(text.RuneCount(s[0])) },
	textPath + ".ByteLen":       func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(text.ByteLen(s[0])) },
	textPath + ".FromCodePoint": func(m *Machine, n []uint64, s []string, r []any) { s[0] = text.FromCodePoint(i32(n[0])) },
	textPath + ".Compare":       func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(text.Compare(s[0], s[1])) },
	textPath + ".Count":         func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(text.Count(s[0], s[1])) },
	textPath + ".Repeat":        func(m *Machine, n []uint64, s []string, r []any) { s[0] = text.Repeat(s[0], i32(n[0])) },
	textPath + ".Index":         func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(text.Index(s[0], s[1])) },
	textPath + ".FromInt":       func(m *Machine, n []uint64, s []string, r []any) { s[0] = text.FromInt(i32(n[0])) },
	textPath + ".FromFloat":     func(m *Machine, n []uint64, s []string, r []any) { s[0] = text.FromFloat(ffrom(n[0]), i32(n[1])) },
	textPath + ".UpperASCII":    func(m *Machine, n []uint64, s []string, r []any) { s[0] = text.UpperASCII(s[0]) },
	textPath + ".LowerASCII":    func(m *Machine, n []uint64, s []string, r []any) { s[0] = text.LowerASCII(s[0]) },
	textPath + ".ParseInt": func(m *Machine, n []uint64, s []string, r []any) {
		v, err := text.ParseInt(s[0])
		n[0], r[0] = fromI32(v), err
	},
	textPath + ".Split": func(m *Machine, n []uint64, s []string, r []any) { r[0] = newStrings(strings.Split(s[0], s[1])) },
	textPath + ".Join": func(m *Machine, n []uint64, s []string, r []any) {
		l := r[0].(*list)
		s[0] = strings.Join(l.s[:l.len], s[0])
	},
	textPath + ".Chars": func(m *Machine, n []uint64, s []string, r []any) {
		res := make([]string, 0, utf8.RuneCountInString(s[0]))
		for _, c := range s[0] {
			res = append(res, string(c))
		}
		r[0] = newStrings(res)
	},
	textPath + ".CodePoints": func(m *Machine, n []uint64, s []string, r []any) {
		l := &list{}
		for _, c := range s[0] {
			l.n = append(l.n, fromI32(c))
			l.len++
		}
		r[0] = l
	},

	numPath + ".Trunc32":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(num.Trunc32(ffrom(n[0]))) },
	numPath + ".TruncU32": func(m *Machine, n []uint64, s []string, r []any) { n[0] = uint64(num.TruncU32(ffrom(n[0]))) },
	numPath + ".Round32":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(num.Round32(ffrom(n[0]))) },

	bigPath + ".Zero":       func(m *Machine, n []uint64, s []string, r []any) { r[0] = bigZero },
	bigPath + ".FromInt32":  func(m *Machine, n []uint64, s []string, r []any) { r[0] = big.NewInt(int64(i32(n[0]))) },
	bigPath + ".FromUint32": func(m *Machine, n []uint64, s []string, r []any) { r[0] = new(big.Int).SetUint64(uint64(uint32(n[0]))) },
	bigPath + ".Parse": func(m *Machine, n []uint64, s []string, r []any) {
		v, ok := parseBig(s[0])
		r[0], n[0] = v, fromBool(ok)
	},
	bigPath + ".String": func(m *Machine, n []uint64, s []string, r []any) { s[0] = bigString(r[0]) },
	bigPath + ".Add":    func(m *Machine, n []uint64, s []string, r []any) { r[0] = new(big.Int).Add(bigOf(r[0]), bigOf(r[1])) },
	bigPath + ".Sub":    func(m *Machine, n []uint64, s []string, r []any) { r[0] = new(big.Int).Sub(bigOf(r[0]), bigOf(r[1])) },
	bigPath + ".Mul":    func(m *Machine, n []uint64, s []string, r []any) { r[0] = new(big.Int).Mul(bigOf(r[0]), bigOf(r[1])) },
	bigPath + ".Quo": func(m *Machine, n []uint64, s []string, r []any) {
		if bigOf(r[1]).Sign() == 0 {
			r[0], n[0] = bigZero, 0
			return
		}
		r[0], n[0] = new(big.Int).Quo(bigOf(r[0]), bigOf(r[1])), 1
	},
	bigPath + ".Rem": func(m *Machine, n []uint64, s []string, r []any) {
		if bigOf(r[1]).Sign() == 0 {
			r[0], n[0] = bigZero, 0
			return
		}
		r[0], n[0] = new(big.Int).Rem(bigOf(r[0]), bigOf(r[1])), 1
	},
	bigPath + ".QuoRem": func(m *Machine, n []uint64, s []string, r []any) {
		if bigOf(r[1]).Sign() == 0 {
			r[0], r[1], n[0] = bigZero, bigZero, 0
			return
		}
		q, rem := new(big.Int).QuoRem(bigOf(r[0]), bigOf(r[1]), new(big.Int))
		r[0], r[1], n[0] = q, rem, 1
	},
	bigPath + ".Neg":  func(m *Machine, n []uint64, s []string, r []any) { r[0] = new(big.Int).Neg(bigOf(r[0])) },
	bigPath + ".Abs":  func(m *Machine, n []uint64, s []string, r []any) { r[0] = new(big.Int).Abs(bigOf(r[0])) },
	bigPath + ".Cmp":  func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(int32(bigOf(r[0]).Cmp(bigOf(r[1])))) },
	bigPath + ".Sign": func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(int32(bigOf(r[0]).Sign())) },
	bigPath + ".Pow10": func(m *Machine, n []uint64, s []string, r []any) {
		r[0] = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(min(max(i32(n[0]), 0), maxPow10))), nil)
	},
	bigPath + ".Int32": func(m *Machine, n []uint64, s []string, r []any) {
		v := bigOf(r[0])
		if !v.IsInt64() || v.Int64() < math.MinInt32 || v.Int64() > math.MaxInt32 {
			n[0], n[1] = 0, 0
			return
		}
		n[0], n[1] = fromI32(int32(v.Int64())), 1
	},
	bigPath + ".Float64": func(m *Machine, n []uint64, s []string, r []any) {
		f, _ := new(big.Float).SetInt(bigOf(r[0])).Float64()
		n[0] = fbits(f)
	},

	optPath + ".None": func(m *Machine, n []uint64, s []string, r []any) { r[0] = nil },
	listPath + ".New": func(m *Machine, n []uint64, s []string, r []any) { r[0] = &list{} },
	msgPath + ".New":  func(m *Machine, n []uint64, s []string, r []any) { r[0] = &list{} },
}

// genericNatives create the natives of generic functions and methods for their type arguments.
var genericNatives = map[string]func(c *compiler, targs []*tinfo, pos token.Pos) *native{
	listPath + ".Of": func(c *compiler, targs []*tinfo, pos token.Pos) *native {
		// the variadic argument is a new list already
		return &native{name: "list.Of", inline: func(g *gen, dst layout, args []layout) { g.move(dst, args[0], layout{r: 1}) }}
	},
	listPath + ".List.Len":   listLen,
	listPath + ".Frozen.Len": listLen,
	msgPath + ".Out.Len":     listLen,
	listPath + ".List.Get":   listGet,
	listPath + ".Frozen.Get": listGet,
	listPath + ".List.Set": func(c *compiler, targs []*tinfo, pos token.Pos) *native {
		el := targs[0].l
		return &native{name: "List.Set", fn: func(m *Machine, n []uint64, s []string, r []any) {
			l := r[0].(*list)
			i := i32(n[0])
			if i < 0 || i >= l.len {
				n[0] = 0
				return
			}
			l.set(int(i), el, n[1:], s, r[1:])
			n[0] = 1
		}}
	},
	listPath + ".List.Pop": func(c *compiler, targs []*tinfo, pos token.Pos) *native {
		el := targs[0].l
		zero := c.zeroSlots(targs[0], pos)
		return &native{name: "List.Pop", fn: func(m *Machine, n []uint64, s []string, r []any) {
			l := r[0].(*list)
			if l.len == 0 {
				zero.copyTo(n, s, r)
				n[el.n] = 0
				return
			}
			l.get(int(l.len-1), el, n, s, r)
			l.truncate(l.len-1, el)
			n[el.n] = 1
		}}
	},
	listPath + ".List.Delete": func(c *compiler, targs []*tinfo, pos token.Pos) *native {
		el := targs[0].l
		return &native{name: "List.Delete", fn: func(m *Machine, n []uint64, s []string, r []any) {
			l := r[0].(*list)
			i := i32(n[0])
			if i < 0 || i >= l.len {
				n[0] = 0
				return
			}
			l.delete(int(i), el)
			n[0] = 1
		}}
	},
	listPath + ".List.Clone": listClone,
	listPath + ".Make": func(c *compiler, targs []*tinfo, pos token.Pos) *native {
		el := targs[0].l
		zero := c.zeroSlots(targs[0], pos)
		return &native{name: "list.Make", fn: func(m *Machine, n []uint64, s []string, r []any) {
			l := &list{}
			for range max(i32(n[0]), 0) {
				l.add(el, zero.n, zero.s, zero.r)
			}
			r[0] = l
		}}
	},
	listPath + ".SortFunc": func(c *compiler, targs []*tinfo, pos token.Pos) *native {
		el := targs[0].l
		return &native{name: "list.SortFunc", fn: func(m *Machine, n []uint64, s []string, r []any) {
			l, fn := r[0].(*list), r[1].(*closure).fn
			idx := make([]int, l.len)
			for i := range idx {
				idx[i] = i
			}
			slices.SortStableFunc(idx, func(a, b int) int {
				res := m.callFunc(fn, func(n []uint64, s []string, r []any) {
					l.get(a, el, n, s, r)
					l.get(b, el, n[el.n:], s[el.s:], r[el.r:])
				})
				return int(int32(m.n[res.n]))
			})
			sorted := &list{len: l.len}
			for _, i := range idx {
				sorted.n = append(sorted.n, l.n[i*int(el.n):(i+1)*int(el.n)]...)
				sorted.s = append(sorted.s, l.s[i*int(el.s):(i+1)*int(el.s)]...)
				sorted.r = append(sorted.r, l.r[i*int(el.r):(i+1)*int(el.r)]...)
			}
			l.n, l.s, l.r = sorted.n, sorted.s, sorted.r
		}}
	},
	dictPath + ".New":        func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "New", t, pos) },
	dictPath + ".Map.Len":    func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "Len", t, pos) },
	dictPath + ".Map.Get":    func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "Get", t, pos) },
	dictPath + ".Map.Has":    func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "Has", t, pos) },
	dictPath + ".Map.Set":    func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "Set", t, pos) },
	dictPath + ".Map.Delete": func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "Delete", t, pos) },
	dictPath + ".Map.$keys":  func(c *compiler, t []*tinfo, pos token.Pos) *native { return dictNatives(c, "$keys", t, pos) },
	optPath + ".Some":        func(c *compiler, t []*tinfo, pos token.Pos) *native { return optNatives(c, "Some", t, pos) },
	optPath + ".Option.Get":  func(c *compiler, t []*tinfo, pos token.Pos) *native { return optNatives(c, "Get", t, pos) },
	optPath + ".Option.Or":   func(c *compiler, t []*tinfo, pos token.Pos) *native { return optNatives(c, "Or", t, pos) },
	optPath + ".Option.IsSome": func(c *compiler, t []*tinfo, pos token.Pos) *native {
		return optNatives(c, "IsSome", t, pos)
	},
	listPath + ".Freeze": listClone,
}

func listLen(c *compiler, targs []*tinfo, pos token.Pos) *native {
	return c.listLen()
}

func (c *compiler) listLen() *native {
	return &native{name: "List.Len", inline: func(g *gen, dst layout, args []layout) { g.emit(opLLen, dst.n, args[0].r, 0) }}
}

func listGet(c *compiler, targs []*tinfo, pos token.Pos) *native {
	el := targs[0].l
	zero := c.zeroSlots(targs[0], pos)
	return &native{name: "List.Get", fn: func(m *Machine, n []uint64, s []string, r []any) {
		l := r[0].(*list)
		i := i32(n[0])
		if i < 0 || i >= l.len {
			zero.copyTo(n, s, r)
			n[el.n] = 0
			return
		}
		l.get(int(i), el, n, s, r)
		n[el.n] = 1
	}}
}

func listClone(c *compiler, targs []*tinfo, pos token.Pos) *native {
	return &native{name: "List.Clone", fn: func(m *Machine, n []uint64, s []string, r []any) {
		l := r[0].(*list)
		r[0] = &list{n: cloneSlice(l.n), s: cloneSlice(l.s), r: cloneSlice(l.r), len: l.len}
	}}
}

func cloneSlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return append([]T(nil), s...)
}

// slots are the slots of a constant value, e.g. a zero value.
type slots struct {
	n []uint64
	s []string
	r []any
}

func (z slots) copyTo(n []uint64, s []string, r []any) {
	copy(n, z.n)
	copy(s, z.s)
	copy(r, z.r)
}

// zeroSlots returns the zero value of t. A reference other than a shared struct is nil, which a checked program
// never reads, e.g. in the zero value, which l.Get(i) returns for an index out of range.
func (c *compiler) zeroSlots(t *tinfo, pos token.Pos) slots {
	z := slots{n: make([]uint64, t.l.n), s: make([]string, t.l.s), r: make([]any, t.l.r)}
	switch t.kind {
	case kShared:
		z.r[0] = c.zeroObject(t, pos)
	case kStruct:
		c.zeroSlotsInto(t, z.r)
	}
	return z
}

// get copies element i into the windows.
func (l *list) get(i int, el layout, n []uint64, s []string, r []any) {
	copy(n[:el.n], l.n[i*int(el.n):])
	copy(s[:el.s], l.s[i*int(el.s):])
	copy(r[:el.r], l.r[i*int(el.r):])
}

func (l *list) set(i int, el layout, n []uint64, s []string, r []any) {
	copy(l.n[i*int(el.n):], n[:el.n])
	copy(l.s[i*int(el.s):], s[:el.s])
	copy(l.r[i*int(el.r):], r[:el.r])
}

func (l *list) add(el layout, n []uint64, s []string, r []any) {
	l.n = append(l.n, n[:el.n]...)
	l.s = append(l.s, s[:el.s]...)
	l.r = append(l.r, r[:el.r]...)
	l.len++
}

func (l *list) truncate(length int32, el layout) {
	clear(l.s[length*el.s:])
	clear(l.r[length*el.r:])
	l.n, l.s, l.r = l.n[:length*el.n], l.s[:length*el.s], l.r[:length*el.r]
	l.len = length
}

func (l *list) delete(i int, el layout) {
	cut := func(k int32) (int, int) { return i * int(k), (i + 1) * int(k) }
	a, b := cut(el.n)
	l.n = append(l.n[:a], l.n[b:]...)
	a, b = cut(el.s)
	l.s = append(l.s[:a], l.s[b:]...)
	clear(l.s[len(l.s) : len(l.s)+int(el.s)])
	a, b = cut(el.r)
	l.r = append(l.r[:a], l.r[b:]...)
	clear(l.r[len(l.r) : len(l.r)+int(el.r)])
	l.len--
}
