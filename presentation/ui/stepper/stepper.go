// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package stepper

import (
	"fmt"
	"strconv"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Style int

const (
	Auto Style = iota
	Simple
	Full
)

// TStepper is a composite component (Stepper).
// It visually represents a sequence of steps in a process, highlighting
// completed, current, and upcoming steps with distinct colors and styles.
// Each step can display a label, and the layout adapts between simple
// or full-sized step representations.
type TStepper struct {
	style          Style     // visual style of the stepper (e.g., horizontal or vertical)
	colorDone      ui.Color  // color for completed steps (default: lightest main color)
	colorCurrent   ui.Color  // color for the active/current step (default: accent)
	colorFuture    ui.Color  // color for upcoming steps (default: darkest main color)
	index          int       // current step index
	steps          []TStep   // list of steps making up the stepper
	simpleStepText string    // label format for simplified step display
	fullCircleSize ui.Length // size of the step circle in full layout
	fullStepWidth  ui.Length // width allocated to each step in full layout
}

func Stepper(steps ...TStep) TStepper {
	return TStepper{
		style:          Auto,
		colorDone:      ui.ColorIcons,
		colorCurrent:   ui.ColorAccent,
		colorFuture:    ui.ColorIconsMuted,
		steps:          steps,
		simpleStepText: "Schritt %d von %d",
		fullCircleSize: ui.L32,
		fullStepWidth:  ui.L120,
	}
}

// FullCircleSize sets the diameter of the step circles in full layout mode.
func (c TStepper) FullCircleSize(length ui.Length) TStepper {
	c.fullCircleSize = length
	return c
}

// FullStepWidth sets the width allocated to each step in full layout mode.
func (c TStepper) FullStepWidth(length ui.Length) TStepper {
	c.fullStepWidth = length
	return c
}

// StepText sets a different localized and parameterized (simple) step text, like "Schritt %d von %d".
// An empty string will omit the step text entirely.
func (c TStepper) StepText(pattern string) TStepper {
	c.simpleStepText = pattern
	return c
}

func (c TStepper) Index(idx int) TStepper {
	c.index = idx
	return c
}

func (c TStepper) Style(style Style) TStepper {
	c.style = style
	return c
}

func (c TStepper) Render(ctx core.RenderContext) core.RenderNode {
	switch c.style {
	case Simple:
		return c.renderSimple(ctx)
	case Full:
		return c.renderFull(ctx)
	case Auto:
		fallthrough
	default:
		width := ctx.Window().Info().Width
		requiredLen := core.DP(len(c.steps)) * c.fullStepWidth.Estimate()
		if width < requiredLen {
			return c.renderSimple(ctx)
		}

		return c.renderFull(ctx)
	}

}

func (c TStepper) renderFull(ctx core.RenderContext) core.RenderNode {
	var cells []ui.TGridCell

	var cellWidths []ui.Length
	for idx, step := range c.steps {
		lastEntry := idx == len(c.steps)-1

		var colorCircle ui.Color
		var colorCircleBackground ui.Color
		var colorLine ui.Color
		var colorCircleText ui.Color
		switch {
		case idx == c.index:
			// current
			colorCircle = c.colorCurrent
			colorLine = c.colorFuture
		case idx < c.index:
			// done
			colorLine = c.colorDone
			colorCircle = c.colorDone
			colorCircleText = ui.ColorBackground
			colorCircleBackground = c.colorDone
		default:
			// future
			colorLine = c.colorFuture
			colorCircle = c.colorFuture
		}

		cellWidth := c.fullStepWidth

		var line core.View
		// we need the calc here, because CSS uses 100% of the container width
		lineWidth := ui.Length(fmt.Sprintf("calc(100%% - %v)", c.fullCircleSize))

		if lastEntry {
			lastHasNoText := step.supportingText == "" && step.headline == ""
			if lastHasNoText {
				lineWidth = ""
				line = nil
				cellWidth = c.fullCircleSize
			} else {
				line = ui.VStack().Frame(ui.Frame{Width: lineWidth})
			}
		} else {
			line = ui.VStack().
				Frame(ui.Frame{Width: lineWidth}).
				Border(ui.Border{TopWidth: ui.L1, TopColor: colorLine})

		}

		cellWidths = append(cellWidths, cellWidth)

		cells = append(cells, ui.GridCell(
			ui.VStack(
				ui.HStack(
					ui.VStack(ui.Text(strconv.Itoa(idx+1)).Color(colorCircleText)).
						BackgroundColor(colorCircleBackground).
						Border(ui.Border{}.Color(colorCircle).Width(ui.L1).Circle()).Frame(ui.Frame{}.Size(c.fullCircleSize, c.fullCircleSize)),

					line,
				).Frame(ui.Frame{Width: cellWidth}), // we need the explicit size here, because CSS cannot calculate the width of "nothing"
				ui.Text(step.headline),
				ui.Text(step.supportingText).Font(ui.Small),
			).Alignment(ui.TopLeading),
		))
	}

	return ui.Grid(cells...).
		Widths(cellWidths...).
		Rows(1).
		Render(ctx)
}

func (c TStepper) renderSimple(ctx core.RenderContext) core.RenderNode {
	var points []core.View

	if c.simpleStepText != "" {
		points = append(points, ui.Text(fmt.Sprintf("Schritt %d von %d", c.index+1, len(c.steps))).
			Padding(ui.Padding{Right: ui.L16}))
	}

	for idx := range c.steps {
		var colorCircle ui.Color
		var colorLine ui.Color
		switch {
		case idx == c.index:
			colorCircle = c.colorCurrent
			colorLine = c.colorFuture
		case idx < c.index:
			colorLine = c.colorDone
			colorCircle = c.colorDone
		default:
			colorLine = c.colorFuture
			colorCircle = c.colorFuture
		}

		points = append(points, ui.VStack().
			BackgroundColor(colorCircle).
			Border(ui.Border{}.Circle()).
			Frame(ui.Frame{Width: ui.L12, Height: ui.L12}),
		)

		if idx < len(c.steps)-1 {
			points = append(points, ui.VStack().
				Frame(ui.Frame{Width: ui.L12}).
				Border(ui.Border{TopWidth: ui.L1, TopColor: colorLine}))
		}
	}

	return ui.HStack(
		points...,
	).Render(ctx)
}

type TStep struct {
	headline       string
	supportingText string
}

func Step() TStep {
	return TStep{}
}

func (c TStep) Headline(headline string) TStep {
	c.headline = headline
	return c
}

func (c TStep) SupportingText(text string) TStep {
	c.supportingText = text
	return c
}
