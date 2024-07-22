package ora

import (
	"fmt"
)

// Color specifies either a hex color like #rrggbb or #rrggbbaa or an internal custom color name.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Color string

func (c Color) WithAlpha(a int8) Color {
	if len(c) == 8 {
		c = c[:len(c)-2]
	}

	return Color(fmt.Sprintf("%s%02x", string(c), a))
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ColorScheme string

const (
	Light ColorScheme = "light"
	Dark  ColorScheme = "dark"
)

// A ColorSet marks a simple struct with public color fields (like Colors) to be a set of colors.
// It returns its unique namespace and has a Default behavior, as a fallback.
// Even though this looks quite cumbersome, for just defining some custom colors, it will play out its strength,
// when designing custom views with complex color sets. If a component requires 10 additional color values and
// you combine 10 different components, you already have to manage and define 100 unstructured color values
// at configuration time. Therefore, we have namespaces and the type safety.
type ColorSet interface {
	// Default returns an initialized color set of the same type as self but with sensible default values set.
	Default(scheme ColorScheme) ColorSet
	// Namespace must be unique within an entire application. "ora" is reserved.
	Namespace() NamespaceName
}

// Colors defines a themes color set. See also https://wiki.worldiety.net/books/design-system-ora/page/farbsystem.
type Colors struct {

	// P0 defines the primary source color and is used for a harmonic overall color expression.
	P0 Color `json:"p0"`
	// P1 is usually used for the background.
	P1 Color `json:"p1"`
	// P2 is usually used for the background of a first level container.
	P2 Color `json:"p2"`
	// P3 is usually used for a card bottom.
	P3 Color `json:"p3"`
	// P4 is usually used for a card body.
	P4 Color `json:"p4"`
	// P5 may be used as a line or dot color.
	P5 Color `json:"p5"`
	// P6 is for hovered containers.
	P6 Color `json:"p6"`
	// P7 is for Text or muted icons.
	P7 Color `json:"p7"`
	// P8 is for normal Text or Icons.
	P8 Color `json:"p8"`
	// P9 is for card tops.
	P9 Color `json:"p9"`

	// S0 is the secondary source color and represents non-area color accents.
	S0 Color `json:"s0"`
	// S1 is for progressbars, charts, some accented headlines or borders.
	S1 Color `json:"s1"`
	// S2 is still tbd.
	S2 Color `json:"s2"
	// S3 is still tbd.`
	S3 Color `json:"s3"`
	// S4 is still tbd.
	S4 Color `json:"s4"`
	// S5 is still tbd.
	S5 Color `json:"s5"`
	// S6 is still tbd.
	S6 Color `json:"s6"`

	// T0 is the tertiary source color and represents a non-area interactive element color accent.
	T0 Color `json:"t0"`
	// T1 is used for buttons.
	T1 Color `json:"t1"`
	// T2 is used for hovered buttons.
	T2 Color `json:"t2"`
	// T3 is used for pressed buttons.
	T3 Color `json:"t3"`

	// Error describes a negative or a destructive intention. In Western Europe usually red. Use it, when the
	// user cannot continue normally and has to fix the problem first.
	Error Color `json:"clE"`
	// Warning describes a critical condition. In Western Europe usually yellow. Use it to warn on situations,
	// which may result in a future error condition.
	Warning Color `json:"clW"`
	// Good describes a positive condition or a confirming intention. In Western Europe usually green.
	// Use it to symbolize something which has been successfully applied.
	Good Color `json:"clG"`
	// Informative shall be used to highlight something, which just changed. E.g. a newly added component or
	// a recommendation from a system. Do not use it to highlight text. In Western Europe usually blue.
	Informative Color `json:"clI"`
	// Disabled defines an otherwise interactive colored area color.
	Disabled Color `json:"clD"`
	// DisabledText defines the color of text which has been rendered on a disabled color.
	DisabledText Color `json:"clT"`
}

func DefaultColors(scheme ColorScheme, primary, secondary, tertiary Color) Colors {
	switch scheme {
	case Dark:
		return darkColors(primary, secondary, tertiary)
	default:
		return lightColors(primary, secondary, tertiary)
	}
}

func darkColors(primary, secondary, tertiary Color) Colors {
	var c Colors
	p := mustParseHSL(string(primary))
	c.P0 = primary
	c.P1 = p.Brightness(10).RGBHex()
	c.P2 = p.Brightness(12).RGBHex()
	c.P3 = p.Brightness(22).RGBHex()
	c.P4 = p.Brightness(17).RGBHex()
	c.P5 = p.Brightness(30).RGBHex()
	c.P6 = p.Brightness(22).RGBHex()
	c.P7 = p.Brightness(60).RGBHex()
	c.P8 = p.Brightness(90).RGBHex()
	c.P9 = p.Brightness(14).RGBHex()

	s := mustParseHSL(string(secondary))
	c.S0 = secondary
	c.S1 = secondary
	c.S2 = secondary.WithAlpha(50)
	c.S3 = secondary.WithAlpha(25)
	c.S4 = s.Brightness(80).RGBHex()
	c.S5 = s.Brightness(80).RGBHex().WithAlpha(50)
	c.S6 = s.Brightness(80).RGBHex().WithAlpha(50)

	_ = mustParseHSL(string(tertiary))
	c.T0 = tertiary
	c.T1 = tertiary
	c.T2 = tertiary.WithAlpha(10)
	c.T3 = tertiary.WithAlpha(25)

	c.Error = "#F47954"
	c.Warning = "#F7A823"
	c.Informative = "#17428C"
	c.Good = "#2BCA73"
	c.Disabled = "#E2E2E2"
	c.DisabledText = "#848484"

	return c
}

func lightColors(primary, secondary, tertiary Color) Colors {
	var c Colors
	p := mustParseHSL(string(primary))
	c.P0 = primary
	c.P1 = p.Brightness(98).RGBHex()
	c.P2 = p.Brightness(96).RGBHex()
	c.P3 = p.Brightness(90).RGBHex()
	c.P4 = p.Brightness(94).RGBHex()
	c.P5 = p.Brightness(70).RGBHex()
	c.P6 = p.Brightness(90).RGBHex()
	c.P7 = p.Brightness(60).RGBHex()
	c.P8 = p.Brightness(10).RGBHex()
	c.P9 = p.Brightness(92).RGBHex()

	s := mustParseHSL(string(secondary))
	c.S0 = secondary
	c.S1 = secondary
	c.S2 = secondary.WithAlpha(50)
	c.S3 = secondary.WithAlpha(25)
	c.S4 = s.Brightness(80).RGBHex()
	c.S5 = s.Brightness(80).RGBHex().WithAlpha(50)
	c.S6 = s.Brightness(80).RGBHex().WithAlpha(50)

	_ = mustParseHSL(string(tertiary))
	c.T0 = tertiary
	c.T1 = tertiary
	c.T2 = tertiary.WithAlpha(10)
	c.T3 = tertiary.WithAlpha(25)

	c.Error = "#F47954"
	c.Warning = "#F7A823"
	c.Informative = "#17428C"
	c.Good = "#2BCA73"
	c.Disabled = "#E2E2E2"
	c.DisabledText = "#848484"

	return c
}

func (c Colors) Default(scheme ColorScheme) ColorSet {
	var p, s, t Color
	p, s, t = "#8462DA", "#A6A5C2", "#EF8A97"
	return DefaultColors(scheme, p, s, t)
}

func (c Colors) Namespace() NamespaceName {
	return "ora"
}
