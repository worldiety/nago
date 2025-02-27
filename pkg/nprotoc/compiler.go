package nprotoc

import (
	"bytes"
	_ "embed"
	"fmt"
	"iter"
	"slices"
	"strings"
)

type Compiler struct {
	marshals []string
	declr    map[Typename]Declaration
	buf      bytes.Buffer
	indent   int
}

func NewCompiler(declarations map[Typename]Declaration) *Compiler {
	return &Compiler{
		declr: declarations,
	}
}

func (c *Compiler) sortedDecl() iter.Seq2[Typename, Declaration] {
	type tupel struct {
		name Typename
		decl Declaration
	}

	var tmp []tupel
	for typename, declaration := range c.declr {
		tmp = append(tmp, tupel{typename, declaration})
	}

	slices.SortFunc(tmp, func(a, b tupel) int {
		aid, aok := a.decl.(IdentityTypeDeclaration)
		bid, bok := b.decl.(IdentityTypeDeclaration)
		if !aok && !bok {
			return strings.Compare(string(a.name), string(b.name))
		}

		if !aok && bok {
			return -1
		}

		if aok && !bok {
			return 1
		}

		return aid.ID() - bid.ID()
	})

	return func(yield func(Typename, Declaration) bool) {
		for _, t := range tmp {
			if !yield(t.name, t.decl) {
				return
			}
		}
	}
}

func (c *Compiler) i() {
	for range c.indent {
		c.buf.WriteString("\t")
	}
}

func (c *Compiler) p(str ...string) {
	for _, s := range str {
		c.buf.WriteString(s)
	}
}

func (c *Compiler) pn(str string) {
	c.i()
	c.buf.WriteString(str)
	c.buf.WriteString("\n")
}

func (c *Compiler) pf(str string, args ...any) {
	c.i()
	c.buf.WriteString(fmt.Sprintf(str, args...))
}

func (c *Compiler) inc() {
	c.indent++
}

func (c *Compiler) dec() {
	c.indent--
}

func trim(s string) string {
	const below = "//protonc:embed below"
	idx := strings.Index(s, below)
	if idx == -1 {
		return s
	}

	return s[idx+len(below):]
}

func (c *Compiler) implements(typename Typename) []Typename {
	var res []Typename
	for declName, decl := range c.sortedDecl() {
		if e, ok := decl.(Enum); ok {
			if slices.Contains(e.Variants, typename) {
				res = append(res, declName)
			}
		}
	}

	return res
}

func linify(buf []byte) []byte {
	var tmp bytes.Buffer
	for i, line := range strings.Split(string(buf), "\n") {
		tmp.WriteString(fmt.Sprintf("%3d: %s\n", i, line))
	}

	return tmp.Bytes()
}

func (c *Compiler) shapeOf(t Typename) (shape, error) {
	d, ok := c.declr[t]
	if !ok {
		return 0, fmt.Errorf("no declaration for type %s", t)
	}

	switch d.(type) {
	case Enum:
		return xobjectAsArray, nil
	case Uint:
		return uvarint, nil
	case Record:
		return record, nil
	case String:
		return byteSlice, nil
	case Array:
		return array, nil
	case Bool:
		return uvarint, nil
	case Map:
		return array, nil
	case Float64:
		return f64, nil
	default:
		return 0, fmt.Errorf("unknown shape type %s", t)
	}
}

func (c *Compiler) isPrimitive(t Typename) bool {
	return c.isString(t) || c.isFloat(t) || c.isBool(t) || c.isInt(t)
}

func (c *Compiler) isString(t Typename) bool {
	d, ok := c.declr[t]
	if !ok {
		return false
	}

	_, ok = d.(String)
	return ok
}

func (c *Compiler) isFloat(t Typename) bool {
	d, ok := c.declr[t]
	if !ok {
		return false
	}

	_, ok = d.(Float64)
	return ok
}

func (c *Compiler) isInt(t Typename) bool {
	d, ok := c.declr[t]
	if !ok {
		return false
	}

	_, ok = d.(Uint)
	return ok
}

func (c *Compiler) isBool(t Typename) bool {
	d, ok := c.declr[t]
	if !ok {
		return false
	}

	_, ok = d.(Bool)
	return ok
}
