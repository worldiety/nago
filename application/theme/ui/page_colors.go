// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uitheme

import (
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/presentation/ui/tabs"
	"golang.org/x/text/language"
)

var (
	StrColorsTitle    = i18n.MustString("nago.theme.colors.title", i18n.Values{language.English: "Colors", language.German: "Farben"})
	StrColorsDesc     = i18n.MustString("nago.theme.colors.desc", i18n.Values{language.English: "Choose three base colors, from which all other colors are calculated. In the advanced mode, every calculated color can be adjusted.", language.German: "Wählen Sie drei Basisfarben, aus denen alle weiteren Farben berechnet werden. Im erweiterten Modus lässt sich jede berechnete Farbe einzeln anpassen."})
	StrAdvancedMode   = i18n.MustString("nago.theme.colors.advanced", i18n.Values{language.English: "Advanced mode", language.German: "Erweiterter Modus"})
	StrLight          = i18n.MustString("nago.theme.colors.light", i18n.Values{language.English: "Light mode", language.German: "Light-Mode"})
	StrDark           = i18n.MustString("nago.theme.colors.dark", i18n.Values{language.English: "Dark mode", language.German: "Dark-Mode"})
	StrBaseColors     = i18n.MustString("nago.theme.colors.base", i18n.Values{language.English: "Base colors", language.German: "Basisfarben"})
	StrMain           = i18n.MustString("nago.theme.colors.main", i18n.Values{language.English: "Main color", language.German: "Hauptfarbe"})
	StrMainDesc       = i18n.MustString("nago.theme.colors.main_desc", i18n.Values{language.English: "Backgrounds, surfaces and text are calculated from it.", language.German: "Daraus werden Hintergründe, Flächen und Text berechnet."})
	StrInteractive    = i18n.MustString("nago.theme.colors.interactive", i18n.Values{language.English: "Interactive color", language.German: "Interaktionsfarbe"})
	StrInteractiveDsc = i18n.MustString("nago.theme.colors.interactive_desc", i18n.Values{language.English: "Buttons, links and other interactive elements.", language.German: "Knöpfe, Links und andere interaktive Elemente."})
	StrAccent         = i18n.MustString("nago.theme.colors.accent", i18n.Values{language.English: "Accent color", language.German: "Akzentfarbe"})
	StrAccentDesc     = i18n.MustString("nago.theme.colors.accent_desc", i18n.Values{language.English: "Accents like charts, progress bars and highlighted headlines.", language.German: "Akzente wie Diagramme, Fortschrittsbalken und hervorgehobene Überschriften."})
	StrCopyFromLight  = i18n.MustString("nago.theme.colors.copy_from_light", i18n.Values{language.English: "Copy base colors of the light mode", language.German: "Basisfarben vom Light-Mode übernehmen"})
	StrCustomized     = i18n.MustString("nago.theme.colors.customized", i18n.Values{language.English: "customized", language.German: "angepasst"})
	StrRestoreColor   = i18n.MustString("nago.theme.colors.restore", i18n.Values{language.English: "Restore the calculated color", language.German: "Berechnete Farbe wiederherstellen"})
	StrDiscardCustom  = i18n.MustString("nago.theme.colors.discard_custom", i18n.Values{language.English: "Discard all adjustments", language.German: "Alle Anpassungen verwerfen"})
	StrPreview        = i18n.MustString("nago.theme.colors.preview", i18n.Values{language.English: "Preview", language.German: "Vorschau"})
	StrDiscard        = i18n.MustString("nago.theme.colors.discard", i18n.Values{language.English: "Discard changes", language.German: "Änderungen verwerfen"})
	StrResetDefaults  = i18n.MustString("nago.theme.colors.reset", i18n.Values{language.English: "Reset to defaults", language.German: "Auf Standard zurücksetzen"})
	StrResetQuestion  = i18n.MustString("nago.theme.colors.reset_question", i18n.Values{language.English: "All colors of the light and the dark mode are reset to the defaults of the application. Your adjustments are lost.", language.German: "Alle Farben des Light- und Dark-Modes werden auf den Standard der Anwendung zurückgesetzt. Ihre Anpassungen gehen verloren."})
	StrSaved          = i18n.MustString("nago.theme.colors.saved", i18n.Values{language.English: "The colors have been saved.", language.German: "Die Farben wurden gespeichert."})
	StrPreviewCard    = i18n.MustString("nago.theme.colors.preview_card", i18n.Values{language.English: "Card", language.German: "Karte"})
	StrPreviewText    = i18n.MustString("nago.theme.colors.preview_text", i18n.Values{language.English: "Normal text", language.German: "Normaler Text"})
	StrPreviewMuted   = i18n.MustString("nago.theme.colors.preview_muted", i18n.Values{language.English: "Muted text", language.German: "Gedämpfter Text"})
)

// Pages are the navigation paths of the theme module.
type Pages struct {
	Colors core.NavigationPath
}

// PageColors lets the user choose the colors of the theme for the light and the dark mode.
func PageColors(wnd core.Window, uc theme.UseCases) core.View {
	if view := alert.IfPermissionDenied(wnd, theme.PermUpdateColors); view != nil {
		return view
	}

	saved := core.StateOf[theme.Customization](wnd, "nago.theme.colors.saved").Init(func() theme.Customization {
		c, err := uc.ReadCustomization(wnd.Subject())
		if err != nil {
			slog.Error("cannot read the theme customization", "err", err)
		}
		return c
	})
	draft := core.StateOf[theme.Customization](wnd, "nago.theme.colors.draft").Init(saved.Get)
	advanced := core.StateOf[bool](wnd, "nago.theme.colors.advanced")
	scheme := core.StateOf[int](wnd, "nago.theme.colors.scheme")
	resetPresented := core.StateOf[bool](wnd, "nago.theme.colors.reset")

	dirty := !reflect.DeepEqual(saved.Get(), draft.Get())

	return ui.VStack(
		resetDialog(wnd, uc, resetPresented, saved, draft),
		ui.H1(StrColorsTitle.Get(wnd)),
		ui.Text(StrColorsDesc.Get(wnd)),
		ui.ToggleField(StrAdvancedMode.Get(wnd), advanced.Get()).InputValue(advanced),
		tabs.Tabs(
			tabs.Page(StrLight.Get(wnd), func() core.View {
				return schemeEditor(wnd, draft, advanced.Get(), false)
			}).Icon(heroOutline.Sun),
			tabs.Page(StrDark.Get(wnd), func() core.View {
				return schemeEditor(wnd, draft, advanced.Get(), true)
			}).Icon(heroOutline.Moon),
		).InputValue(scheme).FullWidth(),
		ui.HLine(),
		ui.HStack(
			ui.SecondaryButton(func() {
				resetPresented.Set(true)
			}).Title(StrResetDefaults.Get(wnd)),
			ui.Spacer(),
			ui.SecondaryButton(func() {
				draft.Set(saved.Get())
			}).Title(StrDiscard.Get(wnd)).Enabled(dirty),
			ui.PrimaryButton(func() {
				if err := uc.UpdateCustomization(wnd.Subject(), draft.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				saved.Set(draft.Get())
				alert.ShowBannerMessage(wnd, alert.Message{
					Title:   StrColorsTitle.Get(wnd),
					Message: StrSaved.Get(wnd),
					Intent:  alert.IntentOk,
				})
			}).Title(rstring.ActionSave.Get(wnd)).Enabled(dirty),
		).Gap(ui.L8).FullWidth().Wrap(true),
	).Alignment(ui.Leading).
		Gap(ui.L16).
		Frame(ui.Frame{}.Larger())
}

func resetDialog(wnd core.Window, uc theme.UseCases, presented *core.State[bool], saved, draft *core.State[theme.Customization]) core.View {
	return alert.Dialog(
		StrResetDefaults.Get(wnd),
		ui.Text(StrResetQuestion.Get(wnd)),
		presented,
		alert.Cancel(nil),
		alert.Confirm(func() (close bool) {
			if err := uc.ResetColors(wnd.Subject()); err != nil {
				alert.ShowBannerError(wnd, err)
				return true
			}

			c, err := uc.ReadCustomization(wnd.Subject())
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return true
			}

			saved.Set(c)
			draft.Set(c)
			return true
		}),
	)
}

// schemeEditor edits the customization of either the light or the dark mode.
func schemeEditor(wnd core.Window, draft *core.State[theme.Customization], advanced bool, dark bool) core.View {
	name, derive := "light", theme.LightMode
	if dark {
		name, derive = "dark", theme.DarkMode
	}

	current := func() theme.SchemeCustomization {
		if dark {
			return draft.Get().Dark
		}
		return draft.Get().Light
	}

	update := func(fn func(sc *theme.SchemeCustomization)) {
		c := draft.Get()
		sc := &c.Light
		if dark {
			sc = &c.Dark
		}

		// the map is shared with the saved customization, so it must not be changed in place
		sc.Overrides = maps.Clone(sc.Overrides)
		fn(sc)
		if len(sc.Overrides) == 0 {
			sc.Overrides = nil
		}

		draft.Set(c)
	}

	sc := current()
	colors := sc.Colors(derive)
	calculated := derive(sc.Base)

	base := func(label, desc i18n.StrHnd, key string, value ui.Color, set func(b *theme.BaseColors, c ui.Color)) core.View {
		return colorInput(wnd, name+".base."+key, label.Get(wnd), desc.Get(wnd), value, func(c ui.Color) {
			update(func(sc *theme.SchemeCustomization) { set(&sc.Base, c) })
		})
	}

	var groups []core.View
	if advanced {
		for _, grp := range colorGroups {
			var fields []core.View
			for _, f := range grp.fields {
				override, overridden := sc.Overrides[f.key]
				label := f.label.Get(wnd)
				if overridden {
					label = fmt.Sprintf("%s (%s)", label, StrCustomized.Get(wnd))
				}

				value := theme.GetColor(colors, f.key)
				if overridden {
					value = override
				}

				fields = append(fields, ui.HStack(
					colorInput(wnd, name+"."+f.key, label, f.desc.Get(wnd), value, func(c ui.Color) {
						update(func(sc *theme.SchemeCustomization) {
							if strings.EqualFold(string(c), string(theme.GetColor(calculated, f.key))) {
								delete(sc.Overrides, f.key)
								return
							}

							if sc.Overrides == nil {
								sc.Overrides = map[string]ui.Color{}
							}
							sc.Overrides[f.key] = c
						})
					}),
					ui.If(overridden, ui.TertiaryButton(func() {
						update(func(sc *theme.SchemeCustomization) { delete(sc.Overrides, f.key) })
					}).PreIcon(heroOutline.ArrowUturnLeft).AccessibilityLabel(StrRestoreColor.Get(wnd))),
				).Gap(ui.L8).Alignment(ui.Leading).FullWidth())
			}

			groups = append(groups, section(grp.title.Get(wnd), fields...))
		}

		if len(sc.Overrides) > 0 {
			groups = append(groups, ui.SecondaryButton(func() {
				update(func(sc *theme.SchemeCustomization) { sc.Overrides = nil })
			}).Title(StrDiscardCustom.Get(wnd)))
		}
	}

	return ui.VStack(
		section(StrPreview.Get(wnd), preview(wnd, colors)),
		section(StrBaseColors.Get(wnd),
			base(StrMain, StrMainDesc, "main", sc.Base.Main, func(b *theme.BaseColors, c ui.Color) { b.Main = c }),
			base(StrInteractive, StrInteractiveDsc, "interactive", sc.Base.Interactive, func(b *theme.BaseColors, c ui.Color) { b.Interactive = c }),
			base(StrAccent, StrAccentDesc, "accent", sc.Base.Accent, func(b *theme.BaseColors, c ui.Color) { b.Accent = c }),
			ui.If(dark, ui.TertiaryButton(func() {
				light := draft.Get().Light.Base
				update(func(sc *theme.SchemeCustomization) { sc.Base = light })
			}).PreIcon(heroOutline.DocumentDuplicate).Title(StrCopyFromLight.Get(wnd))),
		),
		ui.VStack(groups...).Gap(ui.L24).Alignment(ui.Leading).FullWidth(),
	).Gap(ui.L24).
		Alignment(ui.Leading).
		FullWidth().
		Padding(ui.Padding{Top: ui.L16})
}

func section(title string, views ...core.View) core.View {
	return ui.VStack(
		append([]core.View{ui.Text(title).Font(ui.SubTitle)}, views...)...,
	).Gap(ui.L12).Alignment(ui.Leading).FullWidth()
}

// colorInput edits a color. The state follows the given value, so that a discarded or recalculated color is shown.
func colorInput(wnd core.Window, id string, label, desc string, value ui.Color, onChange func(ui.Color)) core.View {
	state := core.StateOf[ui.Color](wnd, "nago.theme.colors.input."+id).Init(func() ui.Color { return value })
	if state.Get() != value {
		state.Set(value)
	}

	state.Observe(onChange)

	return colorpicker.PickerField(label, value).
		InputValue(state).
		DisableAlpha(true).
		SupportingText(desc).
		Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth())
}
