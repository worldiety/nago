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
)

// Content stores all relevant information with
// regard to a selectable Content.
type Content struct {
	ID       string
	Image    []byte
	Icon     core.SVG
	Headline string
	Text     string
}

// TSwitcher is a composite component (Switcher).
// It combines the given Content elements within a modifiable
// ui.TVStack. Only one Content element can be selected at a time
// whose index is then stored in the corresponding state selectedIdx of the TSwitcher.
// The corresponding text and the corresponding image will be faded in
// after selection, since an opacity mechanism is used.
type TSwitcher struct {
	id                  string
	selectedIdx         *core.State[int] // will hold the selected TContent from the TSwitcher
	group               []Content
	padding             Padding
	frame               Frame
	border              Border
	accessibilityLabel  string
	visible             bool
	backgroundColor     Color
	iconBackgroundColor Color
}

// Switcher represents a user interface which spans a visible area containing
// icons, a text and an image. Depending on which icon is selected and
// which related Content is associated the user will see another text and image.
// Only one Content can be selected at a time.
func Switcher(
	selectedIdx *core.State[int],
	contents ...Content,
) TSwitcher {

	minHeight := calculateHeight(contents)

	return TSwitcher{
		selectedIdx: selectedIdx,
		group:       contents,
		// initial default values for border, frame, visible, height and backgroundColor
		border:              Border{}.Radius(L32),
		frame:               Frame{MinHeight: minHeight, Height: Auto, MaxWidth: "100%", Width: Auto},
		visible:             true,
		backgroundColor:     M2,
		iconBackgroundColor: M4,
	}
}

// calculateHeight calculates the height in order to define
// a suitable frame for the given Content elements.
func calculateHeight(contents []Content) Length {
	var res int
	for i := 0; i < len(contents); i++ {
		res += 5 // ui.L80 == 5rem
	}
	res += 16 // ui.L256 == 16rem
	return Length(fmt.Sprintf("%drem", res))
}

// BackgroundColor sets the BackgroundColor of the TSwitcher.
func (t TSwitcher) BackgroundColor(color Color) TSwitcher {
	t.backgroundColor = color
	return t
}

// IconBackgroundColor sets the IconBackgroundColor of the TSwitcher.
func (t TSwitcher) IconBackgroundColor(color Color) TSwitcher {
	t.iconBackgroundColor = color
	return t
}

func (t TSwitcher) WithFrame(fn func(Frame) Frame) DecoredView {
	t.frame = fn(t.frame)
	return t
}

// Frame sets the Frame of the TSwitcher.
func (t TSwitcher) Frame(frame Frame) DecoredView {
	t.frame = frame
	return t
}

// Border sets the Border of the TSwitcher.
func (t TSwitcher) Border(border Border) DecoredView {
	t.border = border
	return t
}

// Visible determines the visibility of the TSwitcher.
func (t TSwitcher) Visible(visible bool) DecoredView {
	t.visible = visible
	return t
}

// AccessibilityLabel sets the AccessibilityLabel of the TSwitcher.
func (t TSwitcher) AccessibilityLabel(label string) DecoredView {
	t.accessibilityLabel = label
	return t
}

// Padding sets the Padding of the TSwitcher.
func (t TSwitcher) Padding(padding Padding) DecoredView {
	t.padding = padding
	return t
}

// ID sets the ID of the TSwitcher.
func (t TSwitcher) ID(id string) TSwitcher {
	t.id = id
	return t
}

// makeResultViews creates a []core.View of the Content elements
// of TSwitcher and returns it.
func (t TSwitcher) makeResultViews(sizeclass core.WindowSizeClass) []core.View {
	if len(t.group) == 0 {
		return nil
	}
	var resultViews []core.View
	for i, cont := range t.group {

		selected := t.selectedIdx.Get() == i
		var opacity float64

		if selected {
			opacity = 1.0
		}

		resultViews = append(resultViews,
			t.resultView(opacity, cont, sizeclass),
		)
	}
	return resultViews
}

// resultView creates a view for the current Content.
func (t TSwitcher) resultView(opacity float64, cont Content, sizeclass core.WindowSizeClass) core.View {
	if sizeclass >= core.SizeClassLarge {
		return t.resultViewLarge(opacity, cont, sizeclass)
	}
	return t.resultViewSmall(opacity, cont, sizeclass)
}

func (t TSwitcher) resultViewSmall(opacity float64, cont Content, sizeClass core.WindowSizeClass) core.View {
	return VStack(

		VStack(
			t.horizontalIconView(),
			contentTexts(cont, opacity),
		).
			Alignment(TopLeading).
			Gap(L20).
			Padding(t.padding).
			Frame(Frame{}.FullWidth()),

		VStack(
			t.img(cont,
				Border{BottomRightRadius: t.border.BottomRightRadius, BottomLeftRadius: t.border.BottomLeftRadius},
				sizeClass,
			),
		).
			Alignment(Bottom).
			Frame(Frame{}.Size(Full, Full)),
	).
		ID(t.id).
		Alignment(TopLeading).
		BackgroundColor(t.backgroundColor).
		Opacity(opacity).
		Animation(AnimateTransition).
		Visible(t.visible).
		AccessibilityLabel(t.accessibilityLabel).
		Border(t.border).
		Frame(t.frame)
}

func (t TSwitcher) resultViewLarge(opacity float64, cont Content, sizeClass core.WindowSizeClass) core.View {
	return VStack(
		HStack(
			HStack(
				VStack(
					t.verticalIconView(),
				).
					Alignment(TopLeading).
					Padding(Padding{Right: L40}),
				contentTexts(cont, opacity),
			).
				Alignment(BottomLeading).
				Padding(t.padding).
				Frame(Frame{}.FullWidth()),

			// image
			HStack(

				VStack(
					t.img(cont,
						Border{TopRightRadius: t.border.TopRightRadius, BottomRightRadius: t.border.BottomRightRadius},
						sizeClass,
					),
				).
					Opacity(opacity).
					Alignment(Trailing).
					Frame(Frame{}.FullWidth()),
			).
				Alignment(Center).
				Frame(Frame{}.FullWidth()),
		).
			Alignment(BottomLeading).
			Frame(t.frame),
	).
		ID(t.id).
		Alignment(BottomLeading).
		BackgroundColor(t.backgroundColor).
		Opacity(opacity).
		Animation(AnimateTransition).
		Visible(t.visible).
		AccessibilityLabel(t.accessibilityLabel).
		Border(t.border).
		WithFrame(func(frame Frame) Frame {
			f := frame
			f.MinHeight = calculateHeight(t.group) // set the min height at least to the calculated height
			// in order to make sure that all icons fit in the container
			return f
		})
}

// img returns the image of the given Content.
func (t TSwitcher) img(cont Content, border Border, sizeClass core.WindowSizeClass) core.View {
	var fr Frame
	if sizeClass >= core.SizeClassLarge {
		fr = Frame{MinHeight: calculateHeight(t.group)}.Size(Full, L560)
	} else {
		fr = Frame{}.Size(Full, L256)
	}

	return Image().
		Embed(cont.Image).
		ObjectFit(FitCover).
		Border(border).
		Frame(fr)
}

// contentTexts returns the headline and the text of the content.
func contentTexts(cont Content, opacity float64) core.View {
	return VStack(
		Text(cont.Headline).
			Font(Font{Size: L32, Weight: HeadlineAndTitleFontWeight}),
		Text(cont.Text),
	).
		Alignment(BottomLeading).
		Gap(L16).
		Opacity(opacity).
		Frame(Frame{}.FullWidth())
}

// verticalIconView returns a vertical icon view.
func (t TSwitcher) verticalIconView() core.View {
	return VStack(
		t.makeIconViews()...,
	).
		Alignment(BottomLeading).
		Gap(L12).
		BackgroundColor(t.iconBackgroundColor).
		Border(Border{}.Radius(L20)).
		Padding(Padding{}.All(L12))
}

// horizontalIconView returns a horizontal icon view.
func (t TSwitcher) horizontalIconView() core.View {
	return HStack(
		t.makeIconViews()...,
	).
		Alignment(BottomLeading).
		Gap(L12).
		BackgroundColor(t.iconBackgroundColor).
		Border(Border{}.Radius(L20)).
		Padding(Padding{}.All(L12))
}

// makeIconViews returns a slice of the resulting icon views.
func (t TSwitcher) makeIconViews() []core.View {
	if len(t.group) == 0 {
		return nil
	}
	var iconViews []core.View
	for i := range t.group {

		var bgColor Color
		var border Border
		var opacity float64

		selected := t.selectedIdx.Get() == i

		if selected {
			bgColor = M3
			border = Border{}.Radius(L20).Color(M8)
			opacity = 1.0
		}

		iconViews = append(iconViews,
			t.iconView(i, bgColor, selected, opacity, border),
		)
	}
	return iconViews
}

// iconView returns a single icon view
// with regard to the currentIdx.
func (t TSwitcher) iconView(
	currentIdx int,
	bgColor Color,
	selected bool,
	opacity float64,
	border Border,
) core.View {

	return VStack(
		IfElse(selected,
			VStack(
				ImageIcon(t.group[t.selectedIdx.Get()].Icon).
					Frame(Frame{}.Size(Full, Full)),
			).Alignment(BottomLeading).
				Opacity(opacity),

			VStack(
				ImageIcon(t.group[currentIdx].Icon).
					Frame(Frame{}.Size(Full, Full)),
			).Alignment(BottomLeading).
				Opacity(opacity+0.5),
		),
	).
		BackgroundColor(bgColor).
		Action(
			func() {
				t.selectedIdx.Set(currentIdx)
			},
		).
		Border(border).
		Padding(Padding{}.All(L12)).
		Frame(Frame{}.Size(L80, L80))
}

// Render builds and returns the UI representation of the TSwitcher.
func (t TSwitcher) Render(context core.RenderContext) core.RenderNode {
	views := t.makeResultViews(context.Window().Info().SizeClass)
	var cells []TGridCell
	for _, view := range views {
		cells = append(cells,
			GridCell(
				view,
			).
				RowStart(1).
				RowEnd(1).
				ColStart(1).
				ColEnd(1),
		)
	}

	return Grid(
		cells...,
	).
		Columns(1).
		Rows(1).
		Frame(t.frame).
		Render(context)
}
