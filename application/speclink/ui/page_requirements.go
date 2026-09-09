// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package uispeclink shows the requirement catalogue of the running application.
//
// It exists because the catalogue is useful to more than a model. An operator wondering whether a behaviour
// is intended, and a support desk answering the same question for somebody else, both want to read it - and
// neither has a checkout.
package uispeclink

import (
	"strings"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/speclink"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"golang.org/x/text/language"
)

var (
	StrTitle = i18n.MustString("nago.speclink.page.title", i18n.Values{
		language.German:  "Anforderungen",
		language.English: "Requirements",
	})
	StrIntro = i18n.MustString("nago.speclink.page.intro", i18n.Values{
		language.German:  "Die Anforderungen, die in dieses Programm einkompiliert sind. Nicht der ganze Baum des Projekts: ein Anforderungspaket, das kein Einstiegspunkt einbindet, fehlt hier.",
		language.English: "The requirements compiled into this program. Not the whole tree of the project: a requirement package no entry point imports is absent here.",
	})
	StrEmpty = i18n.MustString("nago.speclink.page.empty", i18n.Values{
		language.German:  "Keine sichtbaren Anforderungen. Entweder deklariert dieses Programm keine, oder sie sind höher eingestuft als Ihre Berechtigung erlaubt.",
		language.English: "No visible requirements. Either this program declares none, or they are classified above your permission.",
	})

	StrColID     = i18n.MustString("nago.speclink.col.id", i18n.Values{language.German: "Kennung", language.English: "Identifier"})
	StrColTitle  = i18n.MustString("nago.speclink.col.title", i18n.Values{language.German: "Titel", language.English: "Title"})
	StrColKind   = i18n.MustString("nago.speclink.col.kind", i18n.Values{language.German: "Art", language.English: "Kind"})
	StrColStatus = i18n.MustString("nago.speclink.col.status", i18n.Values{language.German: "Zustand", language.English: "Status"})
	StrColText   = i18n.MustString("nago.speclink.col.text", i18n.Values{language.German: "Aussage", language.English: "Statement"})
)

// Pages are the routes of this context.
type Pages struct {
	Requirements core.NavigationPath
}

// PageRequirements lists the requirement catalogue of this binary.
//
// The listing is the same one the assistant reads, through the same use case, so what an operator sees here
// and what the assistant may say are the same set by construction rather than by agreement.
func PageRequirements(wnd core.Window, uc speclink.UseCases) core.View {
	reqs, err := collect(wnd, uc)
	if err != nil {
		return alert.BannerError(err)
	}

	if len(reqs) == 0 {
		return ui.VStack(
			ui.H1(StrTitle.Get(wnd)),
			ui.Text(StrEmpty.Get(wnd)),
		).Gap(ui.L16).Alignment(ui.Leading).FullWidth().Padding(ui.Padding{}.All(ui.L16))
	}

	byID := make(map[speclink.ID]speclink.Requirement, len(reqs))
	for _, r := range reqs {
		byID[r.ID] = r
	}

	return ui.VStack(
		ui.H1(StrTitle.Get(wnd)),
		ui.Text(StrIntro.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[speclink.Requirement, speclink.ID]{
			FindAll: func(yield func(speclink.ID, error) bool) {
				for _, r := range reqs {
					if !yield(r.ID, nil) {
						return
					}
				}
			},
			FindByID: func(id speclink.ID) (option.Opt[speclink.Requirement], error) {
				r, ok := byID[id]
				if !ok {
					return option.None[speclink.Requirement](), nil
				}
				return option.Some(r), nil
			},
			Fields: []dataview.Field[speclink.Requirement]{
				{
					ID:         "id",
					Name:       StrColID.Get(wnd),
					Map:        func(r speclink.Requirement) core.View { return ui.Text(string(r.ID)) },
					Comparator: func(a, b speclink.Requirement) int { return strings.Compare(string(a.ID), string(b.ID)) },
				},
				{
					ID:         "title",
					Name:       StrColTitle.Get(wnd),
					Map:        func(r speclink.Requirement) core.View { return ui.Text(r.Title) },
					Comparator: func(a, b speclink.Requirement) int { return strings.Compare(a.Title, b.Title) },
				},
				{
					ID:   "kind",
					Name: StrColKind.Get(wnd),
					Map:  func(r speclink.Requirement) core.View { return ui.Text(r.Kind) },
				},
				{
					ID:   "status",
					Name: StrColStatus.Get(wnd),
					Map:  func(r speclink.Requirement) core.View { return ui.Text(r.Status) },
				},
				{
					ID:   "text",
					Name: StrColText.Get(wnd),
					Map:  func(r speclink.Requirement) core.View { return ui.Text(r.Text) },
				},
			},
		}).Search(true),
	).Gap(ui.L16).Alignment(ui.Leading).FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

// collect reads the whole visible catalogue once.
//
// The dataview walks its source as a listing followed by a lookup per entry, and the catalogue is a slice in
// memory - so materialising it once and answering both from that is simpler and cheaper than iterating
// twice.
func collect(wnd core.Window, uc speclink.UseCases) ([]speclink.Requirement, error) {
	var out []speclink.Requirement

	for r, err := range uc.FindAllRequirements(wnd.Subject(), speclink.RequirementFilter{}) {
		if err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, nil
}
