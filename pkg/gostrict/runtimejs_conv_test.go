// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

// goja computes with strconv and math of Go, so the bits in goja prove nothing about a browser, whose Number and
// parseFloat may round a text of more than 20 significant digits by its first 20 (L-N1-6). So the new bodies of
// runtime.js use Number( and parseFloat( only on the way for at most 15 significant digits. checkConversions
// finds the uses in the source text of the bodies and of the declarations, which they use, and runs the source in a
// runtime, whose Number and parseFloat record the line, which calls them, and the input of ParseFloat:
//   - a reference without a call, e.g. an alias, is a violation, because it hides the calls;
//   - a call during ParseFloat of an input with more than 15 significant digits, or during Slice, GroupDigits, Min
//     or Max, which have no such way, is a violation;
//   - a call in the text, which no input with at most 15 significant digits reaches, is a violation, because
//     nothing shows, that it lies on that way.
//
// It does not see a conversion without these names, e.g. +s or s * 1.

// n1Bodies are the functions of runtime.js, which N1 adds or changes.
var n1Bodies = []string{"$text_ParseFloat", "$text_Slice", "$text_GroupDigits", "$math_Min", "$math_Max"}

// stripJS replaces the comments and the contents of the strings, templates and regular expressions of src by
// spaces, keeping the lines, so that only the code is left.
func stripJS(src string) string {
	s := &jsStripper{src: src, out: []byte(src)}
	s.code(0, false)
	return string(s.out)
}

type jsStripper struct {
	src string
	out []byte
}

func (s *jsStripper) blank(i, j int) {
	for k := i; k < j && k < len(s.out); k++ {
		if s.out[k] != '\n' {
			s.out[k] = ' '
		}
	}
}

func isIdentByte(c byte) bool {
	return c == '_' || c == '$' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// code strips the code from i on and returns the index after the brace, which closes it, if inner, or the end.
func (s *jsStripper) code(i int, inner bool) int {
	src := s.src
	depth := 0
	prev, word := byte(0), "" // the last character of code, which is no white space, and the last word
	for i < len(src) {
		c := src[i]
		switch {
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			end := strings.IndexByte(src[i:], '\n')
			if end < 0 {
				end = len(src)
			} else {
				end += i
			}
			s.blank(i, end)
			i = end
			continue
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			end := len(src)
			if j := strings.Index(src[i+2:], "*/"); j >= 0 {
				end = i + 2 + j + 2
			}
			s.blank(i, end)
			i = end
			continue
		case c == '"' || c == '\'':
			j := i + 1
			for j < len(src) && src[j] != c && src[j] != '\n' {
				if src[j] == '\\' {
					j++
				}
				j++
			}
			s.blank(i+1, j)
			i, prev, word = j+1, 'a', ""
			continue
		case c == '`':
			i, prev, word = s.template(i+1), 'a', ""
			continue
		case c == '/' && regexStart(prev, word):
			j, class := i+1, false
			for j < len(src) && src[j] != '\n' {
				if src[j] == '\\' {
					j += 2
					continue
				}
				if src[j] == '[' {
					class = true
				} else if src[j] == ']' {
					class = false
				} else if src[j] == '/' && !class {
					break
				}
				j++
			}
			s.blank(i+1, j)
			i, prev, word = j+1, 'a', ""
			continue
		case c == '{':
			depth++
		case c == '}':
			if inner && depth == 0 {
				return i + 1
			}
			depth--
		}
		if isIdentByte(c) {
			j := i
			for j < len(src) && isIdentByte(src[j]) {
				j++
			}
			word, prev, i = src[i:j], src[j-1], j
			continue
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			prev, word = c, ""
		}
		i++
	}
	return i
}

// template strips a template literal from after its backquote on and returns the index after its end.
func (s *jsStripper) template(i int) int {
	src := s.src
	for i < len(src) {
		switch {
		case src[i] == '\\':
			s.blank(i, i+2)
			i += 2
		case src[i] == '`':
			return i + 1
		case src[i] == '$' && i+1 < len(src) && src[i+1] == '{':
			i = s.code(i+2, true)
		default:
			s.blank(i, i+1)
			i++
		}
	}
	return i
}

// regexStart tells, whether a slash after prev and word starts a regular expression instead of a division.
func regexStart(prev byte, word string) bool {
	switch word {
	case "return", "typeof", "case", "in", "of", "delete", "void", "throw", "new", "do", "else", "instanceof", "yield", "await":
		return true
	}
	if word != "" {
		return false
	}
	return prev == 0 || strings.IndexByte("(,=:[!&|?{};+-*%<>~^", prev) >= 0
}

// jsDeclStart finds a declaration of the runtime at the start of a line.
var jsDeclStart = regexp.MustCompile(`^(?:async\s+)?(?:function\*?|const|let|var|class)\s+(\$[A-Za-z0-9_$]*)`)

type jsDecl struct {
	name       string
	start, end int // the indexes of its first line and after its last one
}

// jsDecls returns the top level declarations of the stripped lines, each up to the next one.
func jsDecls(lines []string) map[string]jsDecl {
	ds := map[string]jsDecl{}
	last := ""
	for i, l := range lines {
		if m := jsDeclStart.FindStringSubmatch(l); m != nil {
			if last != "" {
				d := ds[last]
				d.end = i
				ds[last] = d
			}
			ds[m[1]] = jsDecl{name: m[1], start: i, end: len(lines)}
			last = m[1]
		}
	}
	return ds
}

var jsDollar = regexp.MustCompile(`\$[A-Za-z0-9_$]*`)

// jsClosure returns the declarations of the roots and of the ones, which they refer to.
func jsClosure(lines []string, ds map[string]jsDecl, roots []string) []jsDecl {
	seen := map[string]bool{}
	var res []jsDecl
	var visit func(string)
	visit = func(n string) {
		d, ok := ds[n]
		if !ok || seen[n] {
			return
		}
		seen[n] = true
		res = append(res, d)
		for _, l := range lines[d.start:d.end] {
			for _, m := range jsDollar.FindAllString(l, -1) {
				visit(m)
			}
		}
	}
	for _, r := range roots {
		visit(r)
	}
	return res
}

type jsUse struct {
	decl string
	line int // in runtime.js, from 1
	name string
	call bool
}

// jsConversions finds Number and parseFloat in the stripped lines of the declarations. A property of Number, e.g.
// Number.isInteger, is no use, Number.parseFloat is one.
func jsConversions(lines []string, decls []jsDecl) []jsUse {
	var res []jsUse
	for _, d := range decls {
		for i := d.start; i < d.end; i++ {
			l := lines[i]
			for _, name := range []string{"Number", "parseFloat"} {
				for k := 0; ; {
					j := strings.Index(l[k:], name)
					if j < 0 {
						break
					}
					j += k
					k = j + len(name)
					if j > 0 && isIdentByte(l[j-1]) || k < len(l) && isIdentByte(l[k]) {
						continue
					}
					rest := strings.TrimLeft(l[k:], " \t")
					if name == "Number" && strings.HasPrefix(rest, ".") {
						continue
					}
					res = append(res, jsUse{decl: d.name, line: i + 1, name: name, call: strings.HasPrefix(rest, "(")})
				}
			}
		}
	}
	return res
}

// jsGuardSrc replaces Number and parseFloat by functions, which call $n1Guard and then the original ones.
const jsGuardSrc = `(() => {
	const N = Number, P = parseFloat;
	const guarded = function parseFloat(x) { $n1Guard("parseFloat"); return P(x); };
	globalThis.Number = new Proxy(N, {
		apply(t, self, args) { $n1Guard("Number"); return Reflect.apply(t, self, args); },
		construct(t, args, nt) { $n1Guard("Number"); return Reflect.construct(t, args, nt); },
		get(t, k, r) { return k === "parseFloat" ? guarded : Reflect.get(t, k, r); },
	});
	globalThis.parseFloat = guarded;
})();`

type jsGuardCall struct {
	name  string // Number or parseFloat
	line  int    // the line of runtime.js, which called it, or 0
	input string // the call of the runtime, during which it happened
	long  bool   // whether the call of the runtime may not use them
}

// shortDecimals returns texts with at most 15 significant digits, many with small exponents.
func shortDecimals(n int) []string {
	rnd := rand.New(rand.NewPCG(20261105, 11))
	res := make([]string, n)
	for i := range res {
		d := []byte(strconv.Itoa(1 + rnd.IntN(9)))
		for range rnd.IntN(15) {
			d = append(d, byte('0'+rnd.IntN(10)))
		}
		s := string(d)
		if p := rnd.IntN(len(s) + 1); p < len(s) {
			s = s[:p] + "." + s[p:]
		}
		x := rnd.IntN(51) - 25
		if rnd.IntN(2) == 0 {
			x = rnd.IntN(660) - 340
		}
		res[i] = s + "e" + strconv.Itoa(x)
	}
	return res
}

// guardInputs are the inputs of ParseFloat, with which checkConversions runs the runtime.
func guardInputs() []string {
	ins := []string{"123456789012345", "1234567890123456", "999999999999999e-300", "0.1234567890123456e5",
		"123456789012345000000", "1.000000000000000000000001", "9007199254740993", "1.5", "-0", "1e400"}
	for _, r := range parseFloatRows() {
		ins = append(ins, r.in)
	}
	ins = append(ins, randomDecimals(2000)...)
	ins = append(ins, randomForms(500)...)
	return append(ins, shortDecimals(1000)...)
}

// runGuarded runs src with the guards and calls the new bodies.
func runGuarded(src string) (calls []jsGuardCall, long, short int, err error) {
	rt := goja.New()
	var cur jsGuardCall
	rt.Set("$n1Guard", func(name string) {
		c := cur
		c.name = name
		for _, f := range rt.CaptureCallStack(0, nil) {
			if p := f.Position(); p.Filename == "runtime.js" {
				c.line = p.Line
				break
			}
		}
		calls = append(calls, c)
	})
	if _, err := rt.RunString(jsGuardSrc); err != nil {
		return nil, 0, 0, err
	}
	if _, err := rt.RunScript("runtime.js", src); err != nil {
		return nil, 0, 0, fmt.Errorf("runtime.js does not run: %w", err)
	}
	fns := map[string]goja.Callable{}
	for _, n := range n1Bodies {
		v, err := rt.RunString(n)
		if err != nil {
			return nil, 0, 0, err
		}
		f, ok := goja.AssertFunction(v)
		if !ok {
			return nil, 0, 0, fmt.Errorf("%s is no function", n)
		}
		fns[n] = f
	}
	run := func(name string, long bool, args ...any) error {
		vs := make([]goja.Value, len(args))
		for i, a := range args {
			vs[i] = rt.ToValue(a)
		}
		cur = jsGuardCall{input: name + "(" + shortArgs(args) + ")", long: long}
		if _, err := fns[name](goja.Undefined(), vs...); err != nil {
			return fmt.Errorf("%s throws: %w", cur.input, err)
		}
		return nil
	}

	for _, s := range guardInputs() {
		// an input, which is not in the form, has no significant digits and may use them
		l := pfForm.MatchString(s) && sigDigits(s) > 15
		if l {
			long++
		} else {
			short++
		}
		if err := run("$text_ParseFloat", l, s); err != nil {
			return nil, 0, 0, err
		}
	}
	for _, r := range sliceRows {
		if err := run("$text_Slice", true, r.s, r.from, r.to); err != nil {
			return nil, 0, 0, err
		}
	}
	for _, r := range groupRows() {
		if err := run("$text_GroupDigits", true, r.s, r.decimalMark, r.groupMark); err != nil {
			return nil, 0, 0, err
		}
	}
	for _, r := range minMaxRows {
		if r.fn == "Max" || r.fn == "Min" {
			if err := run("$math_"+r.fn, true, r.a, r.b); err != nil {
				return nil, 0, 0, err
			}
		}
	}
	return calls, long, short, nil
}

// checkConversions returns the violations of L-N1-6 in the source of runtime.js and what it checked.
func checkConversions(src string) (violations []string, checked string) {
	lines := strings.Split(stripJS(src), "\n")
	ds := jsDecls(lines)
	for _, n := range n1Bodies {
		if _, ok := ds[n]; !ok {
			violations = append(violations, "runtime.js declares no "+n)
		}
	}
	if len(violations) > 0 {
		return violations, ""
	}
	decls := jsClosure(lines, ds, n1Bodies)
	uses := jsConversions(lines, decls)
	for _, u := range uses {
		if !u.call {
			violations = append(violations, fmt.Sprintf("runtime.js:%d: %s refers to %s without calling it, which hides its calls", u.line, u.decl, u.name))
		}
	}

	calls, long, short, err := runGuarded(src)
	if err != nil {
		return append(violations, err.Error()), ""
	}
	if long < 1000 || short < 1000 {
		violations = append(violations, fmt.Sprintf("the guarded run called ParseFloat with only %d long and %d short inputs", long, short))
	}
	reached := map[int]bool{}
	byLine := map[string][]jsGuardCall{}
	for _, c := range calls {
		if c.long {
			k := fmt.Sprintf("runtime.js:%d: %s(", c.line, c.name)
			byLine[k] = append(byLine[k], c)
		} else {
			reached[c.line] = true
		}
	}
	var keys []string
	for k := range byLine {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c := byLine[k]
		violations = append(violations, fmt.Sprintf("%s is called by %d calls without a way for at most 15 significant digits, e.g. %s", k, len(c), c[0].input))
	}
	for _, u := range uses {
		if u.call && !reached[u.line] {
			violations = append(violations, fmt.Sprintf("runtime.js:%d: %s( in %s is not reached by an input with at most 15 significant digits", u.line, u.name, u.decl))
		}
	}
	var names []string
	for _, d := range decls {
		names = append(names, d.name)
	}
	checked = fmt.Sprintf("declarations %v, uses %v, ParseFloat with %d long and %d short inputs, %d calls of Number or parseFloat",
		names, uses, long, short, len(calls))
	return violations, checked
}

// TestRuntimeConversions checks runtime.js (L-N1-6).
func TestRuntimeConversions(t *testing.T) {
	src, err := os.ReadFile("runtime.js")
	if err != nil {
		t.Fatal(err)
	}
	violations, checked := checkConversions(string(src))
	for _, v := range violations {
		t.Error(v)
	}
	t.Log(checked)
}

// withDecl returns src with the declaration of name replaced by code, or code added, if src has none.
func withDecl(src, name, code string) string {
	lines := strings.Split(src, "\n")
	d, ok := jsDecls(strings.Split(stripJS(src), "\n"))[name]
	if !ok {
		return src + "\n" + code + "\n"
	}
	res := append([]string{}, lines[:d.start]...)
	res = append(res, code)
	return strings.Join(append(res, lines[d.end:]...), "\n")
}

// fakeBodies are new bodies without Number and parseFloat, so that the counter checks do not depend on the ones of
// runtime.js.
var fakeBodies = map[string]string{
	"$text_ParseFloat":  "function $text_ParseFloat(s, p) {\n\treturn [0, null];\n}",
	"$text_Slice":       "function $text_Slice(s, from, to) {\n\treturn Array.from(s).slice($M.max(from, 0), $M.max(to, 0)).join(\"\");\n}",
	"$text_GroupDigits": "function $text_GroupDigits(s, d, g) {\n\treturn s;\n}",
	"$math_Min":         "const $math_Min = (x, y) => $M.min(x, y);",
	"$math_Max":         "const $math_Max = (x, y) => $M.max(x, y);",
}

func fakeRuntime(src string, name, code string) string {
	for _, n := range n1Bodies {
		c := fakeBodies[n]
		if n == name {
			c = code
		}
		src = withDecl(src, n, c)
	}
	return src
}

// TestRuntimeConversionsCounter shows, that checkConversions finds the uses, which L-N1-6 forbids, and lets pass
// those, which it allows: each row changes one body of runtime.js in the text, not in the file.
func TestRuntimeConversionsCounter(t *testing.T) {
	b, err := os.ReadFile("runtime.js")
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	digits := `const d = s.replace(/[eE].*$/, "").replace(/[^0-9]/g, "").replace(/^0+/, "");`
	rows := []struct {
		name, body, code string
		want             string // a part of a violation, or "" for none
	}{
		{"none", "", "", ""},
		{"naive", "$text_ParseFloat", "function $text_ParseFloat(s, p) {\n\treturn [Number(s), null];\n}", "is called by"},
		{"at most 15 digits", "$text_ParseFloat", "function $text_ParseFloat(s, p) {\n\t" + digits + "\n\tif (d.length <= 15) return [Number(s), null];\n\treturn [0, null];\n}", ""},
		{"at most 16 digits", "$text_ParseFloat", "function $text_ParseFloat(s, p) {\n\t" + digits + "\n\tif (d.length <= 16) return [Number(s), null];\n\treturn [0, null];\n}", "is called by"},
		{"parseFloat in a helper", "$text_ParseFloat", "function $n1Helper(s) {\n\treturn parseFloat(s);\n}\nfunction $text_ParseFloat(s, p) {\n\treturn [$n1Helper(s), null];\n}", "parseFloat( is called by"},
		{"alias", "$text_ParseFloat", "const $n1N = Number;\nfunction $text_ParseFloat(s, p) {\n\tif (s === \"never\") return [$n1N(s), null];\n\treturn [0, null];\n}", "without calling it"},
		{"not reached", "$text_ParseFloat", "function $text_ParseFloat(s, p) {\n\tif (s === \"never\") return [Number(s), null];\n\treturn [0, null];\n}", "is not reached"},
		{"in a template", "$text_ParseFloat", "function $text_ParseFloat(s, p) {\n\tconst m = `${Number(s)}`;\n\treturn [0, null];\n}", "is called by"},
		{"in a comment and a string", "$text_ParseFloat", "function $text_ParseFloat(s, p) {\n\t// Number(s) rounds by 20 digits\n\tconst why = \"parseFloat(s) and Number(s)\";\n\treturn [0, null];\n}", ""},
		{"in Slice", "$text_Slice", "function $text_Slice(s, from, to) {\n\treturn Array.from(s).slice(Number($M.max(from, 0)), $M.max(to, 0)).join(\"\");\n}", "is called by"},
		{"Number.parseFloat in GroupDigits", "$text_GroupDigits", "function $text_GroupDigits(s, d, g) {\n\treturn Number.parseFloat(s) > 0 ? s : s;\n}", "parseFloat( is called by"},
		{"in Max", "$math_Max", "const $math_Max = (x, y) => Number($M.max(x, y));", "is called by"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			violations, _ := checkConversions(fakeRuntime(src, r.body, r.code))
			found := false
			for _, v := range violations {
				if r.want != "" && strings.Contains(v, r.want) {
					found = true
				}
			}
			if r.want == "" && len(violations) > 0 || r.want != "" && !found {
				t.Errorf("want a violation with %q, got %q", r.want, violations)
			}
			t.Log(violations)
		})
	}

	// the same in the body of ParseFloat in runtime.js: a call on any input and one, which no input reaches
	lines := strings.Split(src, "\n")
	d, ok := jsDecls(strings.Split(stripJS(src), "\n"))["$text_ParseFloat"]
	if !ok {
		t.Fatal("runtime.js declares no $text_ParseFloat")
	}
	head := lines[d.start]
	i := strings.LastIndex(head, "{")
	if i < 0 {
		t.Fatalf("the body of $text_ParseFloat does not start on its first line: %s", head)
	}
	for _, c := range []struct{ code, want string }{
		{`Number("0.1000000000000000124900090270330110797659");`, "is called by"},
		{`if (typeof $n1Never !== "undefined") Number("1");`, "is not reached"},
	} {
		changed := append([]string{}, lines...)
		changed[d.start] = head[:i+1] + " " + c.code + head[i+1:]
		violations, _ := checkConversions(strings.Join(changed, "\n"))
		found := false
		for _, v := range violations {
			found = found || strings.Contains(v, c.want)
		}
		if !found {
			t.Errorf("%s in the body of $text_ParseFloat: want a violation with %q, got %q", c.code, c.want, violations)
		}
		t.Log(c.code, violations)
	}
}
