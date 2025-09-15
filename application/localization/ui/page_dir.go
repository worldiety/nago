// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilocalization

import (
	"fmt"
	"strings"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/localization"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/picker"
	"golang.org/x/text/language"
)

func PageDir(wnd core.Window, uc localization.UseCases) core.View {
	path := localization.Path(wnd.Values()["path"])

	var dir localization.Directory
	if path != "" {
		d, err := uc.ReadDir(wnd.Subject(), path)
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
	var tag language.Tag
	if len(tags) > 0 {
		tag = tags[0]
	}

	if sel := selectedLanguage.Get(); len(sel) > 0 {
		tag = sel[0]
	}

	bnd, ok := i18n.Default.MatchBundle(tag)
	if !ok {
		return alert.BannerError(fmt.Errorf("no bundle for tag '%s' found", tag))
	}

	return ui.VStack(
		ui.HStack(
			breadcrumb.Breadcrumbs(makeBreadcrumbs(wnd, path)...),
			ui.Spacer(),
			langPicker(wnd, tags, selectedLanguage),
		).FullWidth(),
		ui.Space(ui.L32),

		cardlayout.Layout(
			ui.ForEach(dir.Directories, func(t localization.DirInfo) core.View {
				return cardlayout.Card(t.Name).
					Body(
						ui.Text(StrTranslationSecText.Get(wnd, i18n.Int("totalAmount", t.TotalKeys), i18n.Int("missingAmount", t.TotalMissingKeys))),
					).Footer(ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo("admin/localization/directory", wnd.Values().Put("path", string(t.Path)))
				}).Title(rstring.ActionOpen.Get(wnd)))
			})...,
		),

		ui.Space(ui.L32),

		ui.IfFunc(len(dir.Strings) > 0, func() core.View {
			return dataview.FromData(wnd, dataview.Data[i18n.Message, i18n.Key]{
				FindAll: func(yield func(i18n.Key, error) bool) {
					for _, key := range dir.Strings {
						if !yield(key, nil) {
							return
						}
					}
				},
				FindByID: func(id i18n.Key) (option.Opt[i18n.Message], error) {
					return option.Some(bnd.MessageByKey(id)), nil
				},
				Fields: []dataview.Field[i18n.Message]{
					{
						Name: "Key", Map: func(obj i18n.Message) core.View {
							return ui.Text(string(obj.Key))
						},
					},
					{
						Name: languageName(bnd.Tag()), Map: func(obj i18n.Message) core.View {
							switch obj.Kind {
							case i18n.MessageString, i18n.MessageVarString:
								return ui.Text(obj.Value)
							case i18n.MessageUndefined:
								return ui.Text("?").AccessibilityLabel(StrNotTranslated.Get(wnd))
							case i18n.MessageQuantities:
								return ui.Text(obj.Quantities.String())
							default:
								return ui.Text(fmt.Sprintf("unknown kind: %v", obj.Kind))
							}

						},
					},
				},
			}).Action(func(e i18n.Message) {
				wnd.Navigation().ForwardTo("admin/localization/message", wnd.Values().Put("path", string(e.Key)))
			})
		}),
	).FullWidth().Alignment(ui.Leading)
}

func makeBreadcrumbs(wnd core.Window, path localization.Path) []core.View {
	var res []core.View
	var prefix localization.Path
	elems := strings.Split(string(path), ".")
	for i := 0; i < len(elems)-1; i++ {
		elem := elems[i]
		npath := xstrings.Join2(".", prefix, localization.Path(elem))
		res = append(res, ui.TertiaryButton(func() {
			wnd.Navigation().ForwardTo("admin/localization/directory", wnd.Values().Put("path", string(npath)))
		}).Title(localization.NormalizeAndTitle(elem)))

		prefix = npath
	}

	res = append(res, ui.TertiaryButton(nil).Title(localization.NormalizeAndTitle(elems[len(elems)-1])))

	return res
}

func tagsAndSelected(wnd core.Window) ([]language.Tag, *core.State[[]language.Tag]) {
	tags := i18n.Default.Tags()
	selectedLanguage := core.AutoState[[]language.Tag](wnd).Init(func() []language.Tag {
		tmp, _, _ := language.ParseAcceptLanguage(wnd.Values()["language"])
		if len(tmp) == 0 && len(tags) > 0 {
			return []language.Tag{tags[0]}
		}

		if len(tmp) > 0 {
			return []language.Tag{tmp[0]}
		}

		return nil
	}).Observe(func(newValue []language.Tag) {
		if len(newValue) > 0 {
			wnd.Navigation().ForwardTo(wnd.Path(), wnd.Values().Put("language", newValue[0].String()))
		}
	})

	return tags, selectedLanguage
}

func langPicker(wnd core.Window, tags []language.Tag, selectedLanguage *core.State[[]language.Tag]) picker.TPicker[language.Tag] {
	return picker.Picker[language.Tag]("Sprache", tags, selectedLanguage).
		MultiSelect(false).
		ItemRenderer(func(tag language.Tag) core.View {
			return ui.Text(languageName(tag))
		}).ItemPickedRenderer(func(tags []language.Tag) core.View {
		if len(tags) == 0 {
			return ui.Text(rstring.LabelNothingSelected.Get(wnd))
		}

		return ui.Text(languageName(tags[0]))
	})
}
