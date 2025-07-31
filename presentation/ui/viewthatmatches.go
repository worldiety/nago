// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

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

type TViewThatMatches struct {
	wnd     core.Window
	matches []ViewWithSizeClass
}

// ViewThatMatches returns the best logical match for the given view with size class matcher.
func ViewThatMatches(wnd core.Window, matches ...ViewWithSizeClass) TViewThatMatches {
	return TViewThatMatches{
		wnd:     wnd,
		matches: matches,
	}
}

func (t TViewThatMatches) Render(ctx core.RenderContext) core.RenderNode {
	if len(t.matches) == 0 {
		panic("you must provide at least a single matcher")
	}

	class := t.wnd.Info().SizeClass
	if !class.Valid() {
		slog.Error("frontend has not submitted a window size class, assuming sm")
		class = core.SizeClassSmall
	}

	var best ViewWithSizeClass
	for _, match := range t.matches {
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
		best = t.matches[0]
		for _, match := range t.matches {
			if math.Abs(float64(class.Ordinal()-match.SizeClass.Ordinal())) < math.Abs(float64(match.SizeClass.Ordinal()-best.SizeClass.Ordinal())) {
				best = match
			}
		}
	}

	if best.View == nil {
		panic("unreachable")
	}

	return best.View().Render(ctx)
}
