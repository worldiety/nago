// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"testing"

	nui "go.wdy.de/nago/presentation/ui"
)

// TestEnumsB0 compares the constants of the enums of batch b0 with the constants of nago.
func TestEnumsB0(t *testing.T) {
	enums := []struct {
		name      string
		got, want int64
	}{
		{"PositionDefault", int64(PositionDefault), int64(nui.PositionDefault)},
		{"PositionOffset", int64(PositionOffset), int64(nui.PositionOffset)},
		{"PositionAbsolute", int64(PositionAbsolute), int64(nui.PositionAbsolute)},
		{"PositionFixed", int64(PositionFixed), int64(nui.PositionFixed)},
		{"PositionSticky", int64(PositionSticky), int64(nui.PositionSticky)},
		{"PositionRelative", int64(PositionRelative), int64(nui.PositionRelative)},
		{"AnimateNone", int64(AnimateNone), int64(nui.AnimateNone)},
		{"AnimateTransition", int64(AnimateTransition), int64(nui.AnimateTransition)},
		{"AnimatePulse", int64(AnimatePulse), int64(nui.AnimatePulse)},
		{"AnimateBounce", int64(AnimateBounce), int64(nui.AnimateBounce)},
		{"AnimatePing", int64(AnimatePing), int64(nui.AnimatePing)},
		{"AnimateSpin", int64(AnimateSpin), int64(nui.AnimateSpin)},
		{"OutlineSolid", int64(OutlineSolid), int64(nui.OutlineSolid)},
		{"OutlineDotted", int64(OutlineDotted), int64(nui.OutlineDotted)},
		{"StackLayoutAuto", int64(StackLayoutAuto), int64(nui.StackLayoutAuto)},
		{"StackLayoutVertical", int64(StackLayoutVertical), int64(nui.StackLayoutVertical)},
		{"StackLayoutHorizontal", int64(StackLayoutHorizontal), int64(nui.StackLayoutHorizontal)},
		{"StackLayoutWrap", int64(StackLayoutWrap), int64(nui.StackLayoutWrap)},
		{"StackLayoutFlex", int64(StackLayoutFlex), int64(nui.StackLayoutFlex)},
	}
	for _, e := range enums {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d", e.name, e.got, e.want)
		}
	}
}

// TestColorsB0 compares the color constants with the constants of nago.
func TestColorsB0(t *testing.T) {
	colors := []struct {
		name      string
		got, want Color
	}{
		{"M0", M0, Color(nui.M0)},
		{"M1", M1, Color(nui.M1)},
		{"M2", M2, Color(nui.M2)},
		{"M3", M3, Color(nui.M3)},
		{"M4", M4, Color(nui.M4)},
		{"M5", M5, Color(nui.M5)},
		{"M6", M6, Color(nui.M6)},
		{"M7", M7, Color(nui.M7)},
		{"M8", M8, Color(nui.M8)},
		{"M9", M9, Color(nui.M9)},
		{"A0", A0, Color(nui.A0)},
		{"A1", A1, Color(nui.A1)},
		{"I0", I0, Color(nui.I0)},
		{"I1", I1, Color(nui.I1)},
		{"SE0", SE0, Color(nui.SE0)},
		{"SW0", SW0, Color(nui.SW0)},
		{"SG0", SG0, Color(nui.SG0)},
		{"SV0", SV0, Color(nui.SV0)},
		{"SI0", SI0, Color(nui.SI0)},
		{"ST0", ST0, Color(nui.ST0)},
		{"ColorCardBody", ColorCardBody, Color(nui.ColorCardBody)},
		{"ColorCardTop", ColorCardTop, Color(nui.ColorCardTop)},
		{"ColorCardFooter", ColorCardFooter, Color(nui.ColorCardFooter)},
		{"ColorContainer", ColorContainer, Color(nui.ColorContainer)},
		{"ColorBackground", ColorBackground, Color(nui.ColorBackground)},
		{"ColorAccent", ColorAccent, Color(nui.ColorAccent)},
		{"ColorInputBorder", ColorInputBorder, Color(nui.ColorInputBorder)},
		{"ColorText", ColorText, Color(nui.ColorText)},
		{"ColorLine", ColorLine, Color(nui.ColorLine)},
		{"ColorError", ColorError, Color(nui.ColorError)},
		{"ColorBlack", ColorBlack, Color(nui.ColorBlack)},
		{"ColorWhite", ColorWhite, Color(nui.ColorWhite)},
		{"ColorIcons", ColorIcons, Color(nui.ColorIcons)},
		{"ColorIconsMuted", ColorIconsMuted, Color(nui.ColorIconsMuted)},
		{"ColorInteractive", ColorInteractive, Color(nui.ColorInteractive)},
		{"ColorSemanticGood", ColorSemanticGood, Color(nui.ColorSemanticGood)},
		{"ColorSemanticWarn", ColorSemanticWarn, Color(nui.ColorSemanticWarn)},
		{"ColorSemanticError", ColorSemanticError, Color(nui.ColorSemanticError)},
		{"ColorBannerErrorBackground", ColorBannerErrorBackground, Color(nui.ColorBannerErrorBackground)},
		{"ColorBannerErrorText", ColorBannerErrorText, Color(nui.ColorBannerErrorText)},
		{"ColorBannerInfoBackground", ColorBannerInfoBackground, Color(nui.ColorBannerInfoBackground)},
		{"ColorBannerInfoText", ColorBannerInfoText, Color(nui.ColorBannerInfoText)},
		{"ColorBannerWarningBackground", ColorBannerWarningBackground, Color(nui.ColorBannerWarningBackground)},
		{"ColorBannerWarningText", ColorBannerWarningText, Color(nui.ColorBannerWarningText)},
		{"ColorBannerSuccessBackground", ColorBannerSuccessBackground, Color(nui.ColorBannerSuccessBackground)},
		{"ColorBannerSuccessText", ColorBannerSuccessText, Color(nui.ColorBannerSuccessText)},
	}
	for _, c := range colors {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// TestFontsB0 compares the font presets with the presets of nago, converted by fontFrom.
func TestFontsB0(t *testing.T) {
	fonts := []struct {
		name      string
		got, want Font
	}{
		{"DisplayMedium", DisplayMedium, fontFrom(nui.DisplayMedium)},
		{"DisplaySmall", DisplaySmall, fontFrom(nui.DisplaySmall)},
		{"HeadlineLarge", HeadlineLarge, fontFrom(nui.HeadlineLarge)},
		{"HeadlineMedium", HeadlineMedium, fontFrom(nui.HeadlineMedium)},
		{"TitleLarge", TitleLarge, fontFrom(nui.TitleLarge)},
		{"TitleMedium", TitleMedium, fontFrom(nui.TitleMedium)},
		{"TitleSmall", TitleSmall, fontFrom(nui.TitleSmall)},
		{"BodyLarge", BodyLarge, fontFrom(nui.BodyLarge)},
		{"BodySmall", BodySmall, fontFrom(nui.BodySmall)},
		{"LabelLarge", LabelLarge, fontFrom(nui.LabelLarge)},
		{"LabelMedium", LabelMedium, fontFrom(nui.LabelMedium)},
		{"LabelSmall", LabelSmall, fontFrom(nui.LabelSmall)},
		{"MonoLarge", MonoLarge, fontFrom(nui.MonoLarge)},
		{"MonoMedium", MonoMedium, fontFrom(nui.MonoMedium)},
		{"MonoSmall", MonoSmall, fontFrom(nui.MonoSmall)},
		{"MonoBoldLarge", MonoBoldLarge, fontFrom(nui.MonoBoldLarge)},
		{"MonoBoldMedium", MonoBoldMedium, fontFrom(nui.MonoBoldMedium)},
		{"MonoBoldSmall", MonoBoldSmall, fontFrom(nui.MonoBoldSmall)},
		{"MonoItalicLarge", MonoItalicLarge, fontFrom(nui.MonoItalicLarge)},
		{"MonoItalicMedium", MonoItalicMedium, fontFrom(nui.MonoItalicMedium)},
		{"MonoItalicSmall", MonoItalicSmall, fontFrom(nui.MonoItalicSmall)},
		{"Title", Title, fontFrom(nui.Title)},
		{"SubTitle", SubTitle, fontFrom(nui.SubTitle)},
		{"Large", Large, fontFrom(nui.Large)},
		{"Small", Small, fontFrom(nui.Small)},
		{"Monospace", Monospace, fontFrom(nui.Monospace)},
	}
	for _, f := range fonts {
		if f.got != f.want {
			t.Errorf("%s = %+v, want %+v", f.name, f.got, f.want)
		}
	}
}
