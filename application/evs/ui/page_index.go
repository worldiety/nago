// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uievs

import (
	"fmt"
	"strconv"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
)

type PageIndexOptions[Evt any] struct {
	Perms        evs.Permissions
	EntityName   string
	Pages        Pages
	Prefix       permission.ID
	Index        func(wnd core.Window, uc evs.UseCases[Evt]) core.View
	DecorateView func(wnd core.Window, state *core.State[Evt], view core.View) core.View
	Indexer      []evs.Indexer[Evt]
}

func PageIndex[Evt any](wnd core.Window, uc evs.UseCases[Evt], opts PageIndexOptions[Evt]) core.View {
	if opts.Index == nil {
		opts.Index = newDefaultIndex(wnd, opts)
	}

	return opts.Index(wnd, uc)
}

func newDefaultIndex[Evt any](wnd core.Window, opts PageIndexOptions[Evt]) func(wnd core.Window, uc evs.UseCases[Evt]) core.View {
	var indexer evs.Indexer[Evt]
	for _, e := range opts.Indexer {
		if e.Info().ID == evs.IdxID(wnd.Path().Base()) {
			indexer = e
			break
		}
	}

	if indexer == nil {
		return func(wnd core.Window, uc evs.UseCases[Evt]) core.View {
			return alert.BannerError(std.NewLocalizedError("Indexer not found", fmt.Sprintf("The uri referenced indexer '%s' is not defined (anymore).", wnd.Path().Base())))
		}
	}

	return func(wnd core.Window, uc evs.UseCases[Evt]) core.View {
		it, err := indexer.GroupByPrimaryAsString()
		if err != nil {
			return alert.BannerError(err)
		}

		var entries []indexerEntry
		entriesLookup := map[string]indexerEntry{}
		for p, c := range it {
			e := indexerEntry{Primary: p, Count: c}
			entries = append(entries, e)
			entriesLookup[p] = e
		}

		dv := dataview.FromData(wnd, dataview.Data[indexerEntry, string]{
			FindAll: func(yield func(string, error) bool) {
				for _, entry := range entries {
					if !yield(entry.Identity(), nil) {
						return
					}
				}
			},
			FindByID: func(id string) (option.Opt[indexerEntry], error) {
				if entry, ok := entriesLookup[id]; ok {
					return option.Some(entry), nil
				}

				return option.None[indexerEntry](), nil
			},
			Fields: []dataview.Field[indexerEntry]{
				{
					ID:   "primary",
					Name: indexer.Info().Name,
					Map: func(obj indexerEntry) core.View {
						return ui.Text(obj.Primary)
					},
				},
				{
					ID:   "count",
					Name: rstring.LabelAmount.Get(wnd),
					Map: func(obj indexerEntry) core.View {
						return ui.Text(strconv.Itoa(obj.Count))
					},
				},
			},
		}).Action(func(e indexerEntry) {
			wnd.Navigation().ForwardTo(opts.Pages.Audit, wnd.Values().Put("primary", e.Primary).Put("indexer", string(indexer.Info().ID)))
		}).NextActionIndicator(true)
		
		if wnd.Subject().HasPermission(opts.Perms.Delete) {
			dv = dv.SelectOptions(dataview.NewSelectOptionDelete(wnd, func(selected []string) error {
				for _, id := range selected {
					if err := uc.DeleteByPrimary(wnd.Subject(), indexer.Info().ID, id); err != nil {
						return err
					}
				}

				return nil
			}))
		}

		return ui.VStack(
			ui.Space(ui.L16),
			breadcrumb.Breadcrumbs(
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo("admin", wnd.Values().Put("#", string(opts.Prefix)))
				}).Title(StrDataManagement.Get(wnd)),
			).ClampLeading(),
			ui.H1(opts.EntityName),
			dv,
		).FullWidth().Alignment(ui.Leading)
	}
}

type indexerEntry struct {
	Primary string
	Count   int
}

func (i indexerEntry) Identity() string {
	return i.Primary
}
