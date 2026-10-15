// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package ui creates the views of nago in a program, which runs on a server, e.g. in goja. It mirrors the package
// presentation/ui of nago: the same constructors, builder methods, types and constants, with int32 instead of int.
// Only the server implements its functions, so such a program does not run in a browser.
//
// A program renders a window by an exported function, which takes the window and returns the view:
//
//	func Render(wnd ui.Window) ui.View {
//		name := ui.StringStateOf(wnd, "name")
//		return ui.VStack(
//			ui.Text("Kunde").Font(ui.HeadlineSmall),
//			ui.TextField("Name", name.Get()).InputValue(name),
//			ui.PrimaryButton(func() { save(name.Get()) }).Title("Speichern"),
//		).Gap(ui.L16).Padding(ui.Padding{}.All(ui.L16))
//	}
//
// Each builder method returns its own type, also where nago returns a DecoredView, so that the calls chain. A view
// may take functions of the program, e.g. the action of a button or the function, which renders the cells of a
// sheet. The window calls them later, when the user clicks or the window renders, and then renders the window
// again. A state of the window, e.g. StringStateOf, keeps a value between two renders.
//
//go:build !gostrict

package ui

import (
	"testing"

	nui "go.wdy.de/nago/presentation/ui"
)

// TestValues compares the lengths and alignments of the template with nago.
func TestValues(t *testing.T) {
	lengths := map[string][2]string{
		"L0":             {string(L0), string(nui.L0)},
		"L1":             {string(L1), string(nui.L1)},
		"L2":             {string(L2), string(nui.L2)},
		"L4":             {string(L4), string(nui.L4)},
		"L8":             {string(L8), string(nui.L8)},
		"L12":            {string(L12), string(nui.L12)},
		"L14":            {string(L14), string(nui.L14)},
		"L16":            {string(L16), string(nui.L16)},
		"L20":            {string(L20), string(nui.L20)},
		"L24":            {string(L24), string(nui.L24)},
		"L32":            {string(L32), string(nui.L32)},
		"L40":            {string(L40), string(nui.L40)},
		"L44":            {string(L44), string(nui.L44)},
		"L48":            {string(L48), string(nui.L48)},
		"L60":            {string(L60), string(nui.L60)},
		"L64":            {string(L64), string(nui.L64)},
		"L80":            {string(L80), string(nui.L80)},
		"L96":            {string(L96), string(nui.L96)},
		"L120":           {string(L120), string(nui.L120)},
		"L144":           {string(L144), string(nui.L144)},
		"L160":           {string(L160), string(nui.L160)},
		"L200":           {string(L200), string(nui.L200)},
		"L256":           {string(L256), string(nui.L256)},
		"L320":           {string(L320), string(nui.L320)},
		"L400":           {string(L400), string(nui.L400)},
		"L480":           {string(L480), string(nui.L480)},
		"L560":           {string(L560), string(nui.L560)},
		"L880":           {string(L880), string(nui.L880)},
		"L1200":          {string(L1200), string(nui.L1200)},
		"L1600":          {string(L1600), string(nui.L1600)},
		"Auto":           {string(Auto), string(nui.Auto)},
		"ViewportHeight": {string(ViewportHeight), string(nui.ViewportHeight)},
		"Full":           {string(Full), string(nui.Full)},
	}
	for name, v := range lengths {
		if v[0] != v[1] {
			t.Errorf("%s is %q, nago has %q", name, v[0], v[1])
		}
	}
	alignments := map[string][2]uint32{
		"Center":         {uint32(Center), uint32(nui.Center)},
		"Top":            {uint32(Top), uint32(nui.Top)},
		"Bottom":         {uint32(Bottom), uint32(nui.Bottom)},
		"Leading":        {uint32(Leading), uint32(nui.Leading)},
		"Trailing":       {uint32(Trailing), uint32(nui.Trailing)},
		"TopLeading":     {uint32(TopLeading), uint32(nui.TopLeading)},
		"TopTrailing":    {uint32(TopTrailing), uint32(nui.TopTrailing)},
		"BottomLeading":  {uint32(BottomLeading), uint32(nui.BottomLeading)},
		"BottomTrailing": {uint32(BottomTrailing), uint32(nui.BottomTrailing)},
		"Stretch":        {uint32(Stretch), uint32(nui.Stretch)},
	}
	for name, v := range alignments {
		if v[0] != v[1] {
			t.Errorf("%s is %d, nago has %d", name, v[0], v[1])
		}
	}
	for name, v := range map[string][2]int64{
		"BorderStyleSolid":             {int64(BorderStyleSolid), int64(nui.BorderStyleSolid)},
		"BorderStyleDotted":            {int64(BorderStyleDotted), int64(nui.BorderStyleDotted)},
		"BorderStyleDashed":            {int64(BorderStyleDashed), int64(nui.BorderStyleDashed)},
		"ButtonStylePrimary":           {int64(ButtonStylePrimary), int64(nui.ButtonStylePrimary)},
		"ButtonStyleSecondary":         {int64(ButtonStyleSecondary), int64(nui.ButtonStyleSecondary)},
		"ButtonStyleTertiary":          {int64(ButtonStyleTertiary), int64(nui.ButtonStyleTertiary)},
		"ButtonStylePrimaryCritical":   {int64(ButtonStylePrimaryCritical), int64(nui.ButtonStylePrimaryCritical)},
		"ButtonStyleSecondaryCritical": {int64(ButtonStyleSecondaryCritical), int64(nui.ButtonStyleSecondaryCritical)},
		"ButtonStyleTertiaryCritical":  {int64(ButtonStyleTertiaryCritical), int64(nui.ButtonStyleTertiaryCritical)},
		"TextAlignInherit":             {int64(TextAlignInherit), int64(nui.TextAlignInherit)},
		"TextAlignStart":               {int64(TextAlignStart), int64(nui.TextAlignStart)},
		"TextAlignEnd":                 {int64(TextAlignEnd), int64(nui.TextAlignEnd)},
		"TextAlignCenter":              {int64(TextAlignCenter), int64(nui.TextAlignCenter)},
		"TextAlignJustify":             {int64(TextAlignJustify), int64(nui.TextAlignJustify)},
		"WordBreakInherit":             {int64(WordBreakInherit), int64(nui.WordBreakInherit)},
		"WordBreakNormal":              {int64(WordBreakNormal), int64(nui.WordBreakNormal)},
		"WordBreakBreakAll":            {int64(WordBreakBreakAll), int64(nui.WordBreakBreakAll)},
		"WordBreakKeepAll":             {int64(WordBreakKeepAll), int64(nui.WordBreakKeepAll)},
		"HyphensInherit":               {int64(HyphensInherit), int64(nui.HyphensInherit)},
		"HyphensNone":                  {int64(HyphensNone), int64(nui.HyphensNone)},
		"HyphensManual":                {int64(HyphensManual), int64(nui.HyphensManual)},
		"HyphensAuto":                  {int64(HyphensAuto), int64(nui.HyphensAuto)},
		"WhiteSpaceNormal":             {int64(WhiteSpaceNormal), int64(nui.WhiteSpaceNormal)},
		"WhiteSpaceNoWrap":             {int64(WhiteSpaceNoWrap), int64(nui.WhiteSpaceNoWrap)},
		"WhiteSpacePreWrap":            {int64(WhiteSpacePreWrap), int64(nui.WhiteSpacePreWrap)},
		"NormalFontStyle":              {int64(NormalFontStyle), int64(nui.NormalFontStyle)},
		"ItalicFontStyle":              {int64(ItalicFontStyle), int64(nui.ItalicFontStyle)},
		"BodyFontWeight":               {int64(BodyFontWeight), int64(nui.BodyFontWeight)},
		"HeadlineAndTitleFontWeight":   {int64(HeadlineAndTitleFontWeight), int64(nui.HeadlineAndTitleFontWeight)},
		"DisplayAndLabelFontWeight":    {int64(DisplayAndLabelFontWeight), int64(nui.DisplayAndLabelFontWeight)},
	} {
		if v[0] != v[1] {
			t.Errorf("%s is %d, nago has %d", name, v[0], v[1])
		}
	}
}
