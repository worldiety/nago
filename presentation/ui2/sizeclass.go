package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ViewWithSizeClass struct {
	SizeClass ora.WindowSizeClass
	View      core.View
}

func SizeClass(class ora.WindowSizeClass, view core.View) ViewWithSizeClass {
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
