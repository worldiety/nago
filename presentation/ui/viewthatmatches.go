package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
	"math"
)

type ViewWithSizeClass struct {
	SizeClass core.WindowSizeClass
	View      func() core.View
}

func SizeClass(class core.WindowSizeClass, view func() core.View) ViewWithSizeClass {
	return ViewWithSizeClass{
		SizeClass: class,
		View:      view,
	}
}

// ViewThatMatches returns the best logical match for the given view with size class matcher.
func ViewThatMatches(wnd core.Window, matches ...ViewWithSizeClass) core.View {
	if len(matches) == 0 {
		panic("you must provide at least a single matcher")
	}

	class := wnd.Info().SizeClass
	if !class.Valid() {
		slog.Error("frontend has not submitted a window size class, assuming sm")
		class = core.SizeClassSmall
	}

	var best ViewWithSizeClass
	for _, match := range matches {
		if match.View == nil {
			panic(fmt.Errorf("match branch %v contains nil view, which is not allowed", match.SizeClass))
		}

		if match.SizeClass.Ordinal() > best.SizeClass.Ordinal() && match.SizeClass.Ordinal() <= class.Ordinal() {
			best = match
		}
	}

	if best.SizeClass == 0 {
		// obviously, we have an undefined size class which has no real match
		// pick either the largest or smallest, whatever is nearer
		best = matches[0]
		for _, match := range matches {
			if math.Abs(float64(class.Ordinal()-match.SizeClass.Ordinal())) < math.Abs(float64(match.SizeClass.Ordinal()-best.SizeClass.Ordinal())) {
				best = match
			}
		}
	}

	if best.View == nil {
		panic("unreachable")
	}

	return best.View()
}
