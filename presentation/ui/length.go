package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"strconv"
	"strings"
)

var Full = Relative(1)

// ViewportHeight is a magic value which sets the intrinsic size of an Element to be the smallest available viewport
// height. This is useful, if you have to center a component vertically on screen. Note, that scrollbars may
// or truncated views may appear, if contained view is larger than the view ports height.
const ViewportHeight = "100dvh"

func Absolute(v core.DP) Length {
	return Length(fmt.Sprintf("%vdp", v))
}

// Relative sizes must be interpreted according to the parents intrinsic size. E.g. setting 1 to height or width
// will not cause a visible change, as long as the parent has wrap content semantics. Thus, the parent
// must either have its own intrinsic size or its parent must force a specific size for it also.
func Relative(v core.Weight) Length {
	return Length(fmt.Sprintf("%v%%", v*100))
}

// Length is a very variable type, supporting a variety of declaration types:
//   - absolute units in dp e.g. 42dp
//   - relative units in rem e.g. 0.75rem
//   - relative units in percent e.g. 42%
//   - auto or default is the empty string
//   - current viewport height is 100dvh
//   - calc operator with constants, e.g. calc(100dvh - 2rem)
//   - experimental: everything else is interpreted as a variable name, like for [Color]
//
// Important: everything else is undefined, especially if using other units and calc quirks which are just
// implementation dependent e.g. because passed through directly into CSS.
type Length string

func (l Length) ora() ora.Length {
	return ora.Length(l)
}

func (l Length) Negate() Length {
	if strings.HasPrefix(string(l), "-") {
		return l[1:]
	}

	if l == "" {
		return ""
	}

	if l[0] >= '0' && l[0] <= '9' {
		return "-" + l
	}

	panic("usage error: you cannot negate a non-number length")
}

func (l Length) Mul(s float64) Length {
	if l == "" {
		return ""
	}

	var sb strings.Builder
	var ext Length
	for i, r := range l {
		if r >= '0' && r <= '9' || r == '-' || r == '.' {
			sb.WriteRune(r)
		} else {
			ext = l[i:]
			break
		}
	}

	v, err := strconv.ParseFloat(sb.String(), 64)
	if err != nil {
		panic("usage error: you cannot multiplicate a non-number length")
	}

	return Length(fmt.Sprintf("%.3f%s", v*s, ext))
}

// The following Length sizes are common for the ORA design system and will automatically adjust to the root elements font size.
// It is similar to the effect of Androids SP unit, however its factor is by default at 16, because we just use the CSS semantics.
const (
	// L0 relates to 0px which has usually a different meaning than Auto.
	L0 Length = "0px"
	// L1 relates to hairline which is always 1dp.
	L1 Length = "1px"
	// L2 relates to about 2dp at default font scale.
	L2 Length = "0.125rem"
	// L4 relates to about 4dp at default font scale.
	L4 Length = "0.25rem"
	// L8 relates to about 8dp at default font scale.
	L8 Length = "0.5rem"
	//L12 relates to about 12dp at default font scale.
	L12 Length = "0.75rem"
	// L14 corresponds to 14dp at default font scale.
	L14 Length = "0.875rem"
	// L16 relates to about 16dp at default font scale.
	L16 Length = "1rem"
	// L20 relates to about 20dp at default font scale.
	L20 Length = "1.25rem"
	//L24 relates to about 24dp at default font scale.
	L24 Length = "1.5rem"
	//L32 relates to about 32dp at default font scale.
	L32 Length = "2rem"
	//L40 relates to about 40dp at default font scale.
	L40 Length = "2.5rem"
	//L44 relates to about 44dp at default font scale.
	L44 Length = "2.75rem"
	//L48 relates to about 48dp at default font scale.
	L48 Length = "3rem"
	//L80 relates to about 80dp at default font scale.
	L80 Length = "5rem"
	//L120 relates to about 120dp at default font scale.
	L120 Length = "7.5rem"
	//L144 relates to about 144dp at default font scale.
	L144 Length = "9rem"
	//L160 relates to about 160dp at default font scale.
	L160 Length = "10rem"
	//L200 relates to about 200dp at default font scale.
	L200 Length = "12.5rem"
	//L256 relates to about 256dp at default font scale.
	L256 Length = "16rem"
	//L320 relates to about 320dp at default font scale.
	L320 Length = "20rem"
	//L400 relates to about 400dp at default font scale.
	L400 Length = "25rem"
	//L480 relates to about 480dp at default font scale.
	L480 Length = "30rem"
)

const Auto = ""
