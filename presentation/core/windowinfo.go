package core

import "log/slog"

// DP is Density-independent pixels: an abstract unit that is based on the physical density of the screen.
// These units are relative to a 160 dpi (dots per inch) screen, on which 1 dp is roughly equal to 1 px.
// When running on a higher density screen, the number of pixels used to draw 1 dp is scaled up by a factor
// appropriate for the screen's dpi.
//
// Likewise, when on a lower-density screen, the number of pixels used for 1 dp is scaled down.
// The ratio of dps to pixels changes with the screen density, but not necessarily in direct proportion.
// Using dp units instead of px units is a solution to making the view dimensions in your layout
// resize properly for different screen densities. It provides consistency for the real-world sizes of
// your UI elements across different devices.
// Source: https://developer.android.com/guide/topics/resources/more-resources.html#Dimension
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type DP float64

// Density describes the scale factor of physical pixels to screen pixels normalized to a 160dpi screen.
// This is identical to the Android specification. On a 160dpi screen, this factor is 1. Note, that
// this may also be used to optimize accessibility which makes everything equally larger. There is also the
// concept of SP, but that is usually implemented at the frontend interpreter anyway.
type Density float64

// Weight is between 0-1 and can be understood as 1 = 100%, however implementations must normalize the total
// of all weights and recalculate the effective percentage.
type Weight float64

type WindowInfo struct {
	Width       DP
	Height      DP
	Density     Density
	SizeClass   WindowSizeClass
	ColorScheme ColorScheme
}

// WindowSizeClass represents media break points of the screen which an ora application is shown.
// The definition of a size class is disjunct and for all possible sizes, exact one size class will match.
// See also https://developer.android.com/develop/ui/views/layout/window-size-classes and
// https://tailwindcss.com/docs/responsive-design.
type WindowSizeClass string

func (w WindowSizeClass) Ordinal() int {
	switch w {
	case SizeClassSmall:
		return 1
	case SizeClassMedium:
		return 2
	case SizeClassLarge:
		return 3
	case SizeClassXL:
		return 4
	case SizeClass2XL:
		return 5
	default:
		return 0
	}
}

func (w WindowSizeClass) Valid() bool {
	return w.Ordinal() != 0
}

const (
	// SizeClassSmall are devices below 640 dp screen width.
	SizeClassSmall WindowSizeClass = "sm"
	// SizeClassMedium are devices below 768dp screen width.
	SizeClassMedium WindowSizeClass = "md"
	// SizeClassLarge are devices below 1024dp screen width.
	SizeClassLarge WindowSizeClass = "lg"
	// SizeClassXL are devices below 1280dp screen width.
	SizeClassXL WindowSizeClass = "xl"
	// SizeClass2XL are devices below 1536dp screen width.
	SizeClass2XL WindowSizeClass = "2xl"
)

type ViewWithSizeClass struct {
	SizeClass WindowSizeClass
	View      View
}

func SizeClass(class WindowSizeClass, view View) ViewWithSizeClass {
	return ViewWithSizeClass{
		SizeClass: class,
		View:      view,
	}
}

// ViewThatMatches returns the best logical match for the given view with size class matcher.
func ViewThatMatches(wnd Window, matches ...ViewWithSizeClass) View {
	if len(matches) == 0 {
		panic("you must provide at least a single matcher")
	}

	class := wnd.Info().SizeClass
	if !class.Valid() {
		slog.Error("frontend has not submitted a window size class, assuming sm")
		class = SizeClassSmall
	}

	var best ViewWithSizeClass
	for _, match := range matches {
		if match.SizeClass.Ordinal() > best.SizeClass.Ordinal() && match.SizeClass.Ordinal() <= class.Ordinal() {
			best = match
		}
	}

	if best.SizeClass == "" && best.View == nil {
		panic("unreachable")
	}

	if best.View == nil {
		panic("you must not provide an empty view in match")
	}

	return best.View
}
