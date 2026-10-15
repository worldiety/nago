// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

const (
	stdPrefix = "go.wdy.de/nago/pkg/gostrict/std/"
	listPath  = stdPrefix + "list"
	dictPath  = stdPrefix + "dict"
	textPath  = stdPrefix + "text"
	numPath   = stdPrefix + "num"
	jsPath    = stdPrefix + "js"
	msgPath   = stdPrefix + "msg"
	optPath   = stdPrefix + "opt"
)

// stdAllowed contains the packages of the standard library, which the subset allows, with the functions, which
// behave the same in Go and in JavaScript and never panic. Constants of these packages are always allowed.
var stdAllowed = map[string]map[string]bool{
	"math": set("Abs", "Acos", "Asin", "Atan", "Atan2", "Cbrt", "Ceil", "Copysign", "Cos", "Exp", "Floor", "Hypot",
		"Inf", "IsInf", "IsNaN", "Log", "Log10", "Log2", "Max", "Min", "Mod", "NaN", "Pow", "Round", "Signbit", "Sin",
		"Sqrt", "Tan", "Trunc"),
	"strings": set("Contains", "HasPrefix", "HasSuffix", "ReplaceAll", "TrimPrefix", "TrimSpace", "TrimSuffix"),
	"errors":  set("New"),
	"iter":    set("Seq", "Seq2"),
}

func set(names ...string) map[string]bool {
	m := map[string]bool{}
	for _, n := range names {
		m[n] = true
	}
	return m
}

type hint struct {
	why, fix string
}

// importHints explain, what replaces a package of the standard library.
var importHints = map[string]hint{
	"fmt": {
		"fmt formats with reflection, which the subset does not have",
		"build strings with + and text.FromInt(n) or text.FromFloat(f, decimals), join a list with text.Join(l, sep), return values instead of printing them",
	},
	"strconv": {
		"its functions work with int, which the subset replaces by int32",
		`use text.FromInt(n), text.FromFloat(f, decimals) and text.ParseInt(s) from "go.wdy.de/nago/pkg/gostrict/std/text"`,
	},
	"sort": {
		"sort works on slices, which the subset replaces by list.List",
		`use list.SortFunc(l, func(a, b T) int { ... }) from "go.wdy.de/nago/pkg/gostrict/std/list"`,
	},
	"slices": {
		"the subset replaces slices by list.List",
		`use the methods of list.List and list.SortFunc from "go.wdy.de/nago/pkg/gostrict/std/list"`,
	},
	"maps": {
		"the subset replaces maps by dict.Map",
		`use dict.Map from "go.wdy.de/nago/pkg/gostrict/std/dict"`,
	},
	"unicode/utf8": {
		"Go counts the bytes of UTF-8, JavaScript the units of UTF-16",
		`use text.RuneCount, text.Chars or text.CodePoints from "go.wdy.de/nago/pkg/gostrict/std/text"`,
	},
	"unicode": {
		"the tables of Unicode differ between Go and JavaScript",
		`compare code points from text.CodePoints, or use text.UpperASCII and text.LowerASCII`,
	},
	"bytes": {
		"the subset has no bytes",
		`use strings and the functions of "go.wdy.de/nago/pkg/gostrict/std/text"`,
	},
	"time": {
		"time is not deterministic and differs between the targets",
		"pass the time as a float64 parameter, e.g. the elapsed seconds",
	},
	"math/rand": {
		"random numbers differ between Go and JavaScript",
		"implement a small generator with uint32 arithmetic, e.g. xorshift32, and pass the seed as a parameter",
	},
	"math/rand/v2": {
		"random numbers differ between Go and JavaScript",
		"implement a small generator with uint32 arithmetic, e.g. xorshift32, and pass the seed as a parameter",
	},
	"sync": {
		"the subset runs single threaded and has no goroutines",
		"remove the locks",
	},
}

// stdFuncHints explain, what replaces a function of an allowed package.
var stdFuncHints = map[string]hint{
	"strings.Split":     {"it returns a slice", `use text.Split(s, sep), which returns a list.List[string]`},
	"strings.SplitN":    {"it returns a slice", `use text.Split(s, sep) and join the rest with text.Join`},
	"strings.Fields":    {"it returns a slice", `use text.Split(strings.TrimSpace(s), " ") and skip empty parts, which splits only at spaces; to split at any white space, walk text.Chars(s).Values() and collect the parts`},
	"strings.Join":      {"it takes a slice", `use text.Join(l, sep) with a list.List[string]`},
	"strings.Index":     {"it returns a byte index, which differs in JavaScript", `use text.Index(s, sub), which returns the index of the character`},
	"strings.LastIndex": {"it returns a byte index, which differs in JavaScript", `use text.Index on parts of text.Split`},
	"strings.IndexByte": {"it returns a byte index, which differs in JavaScript", `use text.Index(s, sub)`},
	"strings.IndexRune": {"it returns a byte index, which differs in JavaScript", `use text.Index(s, text.FromCodePoint(r))`},
	"strings.ToUpper":   {"the case mapping of Unicode differs between Go and JavaScript, e.g. for ß", `use text.UpperASCII(s)`},
	"strings.ToLower":   {"the case mapping of Unicode differs between Go and JavaScript", `use text.LowerASCII(s)`},
	"strings.EqualFold": {"the case folding of Unicode differs between Go and JavaScript", `compare text.LowerASCII(a) == text.LowerASCII(b)`},
	"strings.Builder":   {"a builder works on bytes", `concatenate with + or collect the parts in a list.List[string] and use text.Join`},
	"strings.Trim":      {"it is not part of the subset yet", `use strings.TrimPrefix, strings.TrimSuffix or strings.TrimSpace`},
	"strings.Compare":   {"it returns an int, which the subset replaces by int32", `use text.Compare(a, b)`},
	"strings.Count":     {"it returns an int, which the subset replaces by int32", `use text.Count(s, sub)`},
	"strings.Repeat":    {"it panics for a negative count", `use text.Repeat(s, n), which returns "" for n <= 0`},
	"errors.Is":         {"errors are compared by identity, which the subset does not support", "return a distinct result, e.g. a bool or an int code, instead of comparing errors"},
	"errors.As":         {"it needs type assertions", "return a distinct result instead of inspecting the error"},
	"math.Float64bits":  {"it needs uint64", "the subset has no 64 bit integers"},
}
