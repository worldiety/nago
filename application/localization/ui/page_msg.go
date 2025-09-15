// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilocalization

import (
	"fmt"
	"slices"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/progress"
	"golang.org/x/text/language"
)

func PageMessage(wnd core.Window, uc localization.UseCases) core.View {
	path := localization.Path(wnd.Values()["path"])
	stringkeys := wnd.Values()["stringkeys"]

	var dir localization.Directory
	if stringkeys != "true" {
		d, err := uc.ReadDir(wnd.Subject(), path.Parent())
		if err != nil {
			return alert.BannerError(err)
		}
		dir = d
	} else {
		keys, err := uc.ReadStringKeys(wnd.Subject())
		if err != nil {
			return alert.BannerError(err)
		}

		dir.Strings = keys
	}

	tags, selectedLanguage := tagsAndSelected(wnd)

	var msg i18n.Message
	selectedLang := wnd.Locale()
	if selTags := selectedLanguage.Get(); len(selTags) > 0 {
		selectedLang = selTags[0]
	}

	resources, err := uc.FindResources(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	var bundle *i18n.Bundle
	if bnd, ok := resources.MatchBundle(selectedLang); ok {
		msg = bnd.MessageByKey(i18n.Key(path))
		bundle = bnd
	}

	strValue := core.StateOf[string](wnd, string(msg.Key)).Init(func() string {
		return msg.Value
	})

	return ui.VStack(
		ui.HStack(
			breadcrumb.Breadcrumbs(makeBreadcrumbs(wnd, path)...),
			ui.Spacer(),
			langPicker(wnd, tags, selectedLanguage),
		).FullWidth(),
		ui.Space(ui.L32),

		toolbar(wnd, uc, selectedLang, strValue, bundle, dir, msg),
		ui.Space(ui.L32),
		ui.H2(string(msg.Key)),
		messageForm(wnd, uc, strValue, selectedLang, msg),
		ui.IfFunc(msg.Kind == i18n.MessageUndefined, func() core.View {
			// give some hints from other languages
			var alternatives []string
			for _, tag := range resources.Tags() {
				if bnd, ok := resources.Bundle(tag); ok {
					if msg := bnd.MessageByKey(msg.Key); msg.Valid() {
						alternatives = append(alternatives, fmt.Sprintf("%s: %s", languageName(tag), msg.String()))
					}
				}
			}

			return ui.VStack(
				ui.Space(ui.L32),
				ui.H2(StrNotTranslated.Get(wnd)),
				ui.VStack(
					ui.ForEach(alternatives, func(t string) core.View {
						return ui.Text(t)
					})...,
				).Alignment(ui.Leading),
			).Alignment(ui.Leading)
		}),
		ui.Space(ui.L32),
	).FullWidth().Alignment(ui.Leading)
}

func messageForm(wnd core.Window, uc localization.UseCases, strValue *core.State[string], tag language.Tag, msg i18n.Message) core.View {
	t := i18n.Default.MessageType(msg.Key)
	switch t {
	case i18n.MessageString, i18n.MessageVarString:
		return staticStringForm(wnd, uc, strValue, tag, msg)
	default:
		return ui.Text(fmt.Sprintf("unsupported message type: %v", msg.Kind))
	}
}

func staticStringForm(wnd core.Window, uc localization.UseCases, strValue *core.State[string], tag language.Tag, msg i18n.Message) core.View {
	lines := strings.Count(msg.Value, "\n") + 1
	lines = min(20, lines)

	return ui.VStack(
		ui.IfFunc(msg.Kind == i18n.MessageVarString, func() core.View {
			hasHints := false
			v := ui.VStack(
				ui.Each(i18n.Default.VarHints(msg.Key), func(t i18n.VarHint) core.View {
					hasHints = true
					return ui.HStack(ui.Text(t.Name+": ").Font(ui.Font{Name: ui.MonoFontName}), ui.Text(t.Description))
				})...,
			).Alignment(ui.Leading)

			if hasHints {
				return v.Padding(ui.Padding{Bottom: ui.L32})
			}

			return v
		}),
		ui.TextField(languageName(tag), strValue.Get()).InputValue(strValue).Lines(lines).FullWidth().SupportingText(i18n.Default.Hint(msg.Key)),
	).FullWidth().Alignment(ui.Leading)
}

func languageName(tag language.Tag) string {
	switch tag {
	case language.English:
		return "English"
	case language.French:
		return "Français"
	case language.German:
		return "Deutsch"
	case language.Italian:
		return "Italiano"
	case language.Spanish:
		return "Español"
	case language.Swedish:
		return "Svenska"
	default:
		return tag.String()
	}
}

func toolbar(wnd core.Window, uc localization.UseCases, tag language.Tag, value *core.State[string], bnd *i18n.Bundle, dir localization.Directory, msg i18n.Message) core.View {
	if bnd == nil {
		return alert.Banner("no bundle", "no bundle available")
	}

	translated := 0
	for _, key := range dir.Strings {
		if bnd.MessageTypeByKey(key) != i18n.MessageUndefined {
			translated++
		}
	}

	updatedMessage := msg
	updatedMessage.Value = value.Get()

	idx := slices.Index(dir.Strings, msg.Key)

	return ui.VStack(
		ui.HStack(ui.Text(fmt.Sprintf("%d/%d übersetzt", translated, len(dir.Strings)))).Alignment(ui.Trailing).FullWidth(),
		progress.LinearProgress().Progress(float64(translated)/float64(len(dir.Strings))).FullWidth(),
		ui.HStack(
			ui.TertiaryButton(func() {
				wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put("path", string(dir.Strings[idx-1])))
			}).Title(rstring.ActionPrevious.Get(wnd)).PreIcon(flowbiteOutline.ChevronLeft).Enabled(idx > 0),

			ui.TertiaryButton(func() {
				wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put("path", string(dir.Strings[idx+1])))
			}).Title(rstring.ActionNext.Get(wnd)).PostIcon(flowbiteOutline.ChevronRight).Enabled(idx < len(dir.Strings)-1),
			ui.Spacer(),

			ui.SecondaryButton(func() {
				if err := uc.UpdateMessage(wnd.Subject(), tag, updatedMessage); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				if err := uc.Flush(wnd.Subject()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				if idx < len(dir.Strings)-1 {
					wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put("path", string(dir.Strings[idx+1])))
				}

			}).Title(rstring.ActionSaveAndNext.Get(wnd)),
		).FullWidth().Gap(ui.L8),
	).Gap(ui.L8).
		FullWidth().
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Padding(ui.Padding{}.All(ui.L16))

}
