// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uirebac

import (
	"os"
	"slices"
	"strings"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	ucrebac "go.wdy.de/nago/application/rebac/uc"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/navsplitview"
)

const (
	usrRelPrefix = "usr_rel_"
	relUsrPrefix = "rel_"
)

func PageEditor(wnd core.Window, uc ucrebac.UseCases) core.View {
	var res rebac.Resources
	err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
		r, ok := rdb.LookupResources(rebac.Namespace(wnd.Values()["resources"]))
		if !ok {
			return os.ErrNotExist
		}

		res = r
		return nil
	})
	if err != nil {
		return alert.BannerError(err)
	}

	return navsplitview.ThreeColumn(navsplitview.NavFn(func(id navsplitview.ViewID) core.View {
		switch {
		case id == "source":
			return allSourceEntries(wnd, res)
		case strings.HasPrefix(string(id), usrRelPrefix):
			id := rebac.Instance(strings.TrimPrefix(string(id), usrRelPrefix))
			//return allRelations(wnd, res, id)

			return declaredRelations(wnd, uc, rebac.Entity{
				Namespace: res.Identity(),
				Instance:  id,
			})
		default:

			rid := rebac.Instance(strings.TrimPrefix(wnd.Values()[navsplitview.KindContent.QueryKey("")], usrRelPrefix))
			rel := rebac.Relation(strings.TrimPrefix(wnd.Values()[navsplitview.KindDetail.QueryKey("")], relUsrPrefix))

			return declaredTargets(wnd, uc, rebac.Select().Where().Source().Is(res.Identity(), rid).Where().Relation().Has(rel))
		}
	})).Default("source", "", "").
		AlignmentSidebar(ui.Stretch).
		AlignmentContent(ui.Stretch).
		AlignmentDetail(ui.Stretch).
		Frame(ui.Frame{Height: "calc(100vh - 6rem)", Width: ui.Full})

}

func declaredTargets(wnd core.Window, uc ucrebac.UseCases, query rebac.Query) core.View {
	var targets []rebac.InstanceInfo

	err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
		for triple, err := range rdb.Query(query) {
			if err != nil {
				return err
			}

			targetRes, ok := rdb.LookupResources(triple.Target.Namespace)
			requireDebugInfo := false
			if ok {
				optInfo, err := targetRes.FindByID(wnd.Context(), rebac.NewInfoID(triple.Target.Namespace, triple.Target.Instance))
				if err != nil {
					return err
				}

				if optInfo.IsSome() {
					targets = append(targets, optInfo.Unwrap())
				} else {
					requireDebugInfo = true
				}

			} else {
				requireDebugInfo = true
			}

			if requireDebugInfo {
				targets = append(targets, rebac.InstanceInfo{
					ID:          triple.Target.Instance,
					Name:        string(triple.Target.Instance),
					Description: string(triple.Target.Namespace),
				})
			}

		}

		return nil
	})

	if err != nil {
		return alert.BannerError(err)
	}

	return ui.ScrollView(
		dataview.FromData(wnd, dataview.Data[rebac.InstanceInfo, rebac.InfoID]{
			ID: "declaredTargets",
			FindAll: func(yield func(rebac.InfoID, error) bool) {
				for _, triple := range targets {
					if !yield(triple.Identity(), nil) {
						return
					}
				}
			},
			FindByID: func(id rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
				for _, triple := range targets {
					if triple.Identity() == id {
						return option.Some(triple), nil
					}
				}

				return option.None[rebac.InstanceInfo](), nil
			},
			Fields: []dataview.Field[rebac.InstanceInfo]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj rebac.InstanceInfo) core.View {
						return ui.Text(obj.Name)
					},
				},
				{
					ID:   "desc",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj rebac.InstanceInfo) core.View {
						return ui.Text(obj.Description)
					},
					Comparator: func(a, b rebac.InstanceInfo) int {
						return strings.Compare(a.Description, b.Description)
					},
				},
			},
		}).Search(true).
			ListOptions(dataview.ListOptions[rebac.InfoID]{
				ColorBody:    option.Some(ui.Color("")),
				ColorCaption: option.Some(ui.Color("")),
				ColorFooter:  option.Some(ui.Color("")),
			}).
			Style(dataview.List).
			CreateOptions(dataview.CreateOption{
				Name: rstring.ActionCreate.Get(wnd),
				Action: func() error {
					return nil
				},
			}).
			SelectOptions(dataview.NewSelectOptionDelete(wnd, func(selected []rebac.InfoID) error {
				return nil
			})).
			Selection(true),
	).Axis(ui.ScrollViewAxisVertical).FullWidth()
}

func declaredRelations(wnd core.Window, uc ucrebac.UseCases, src rebac.Entity) core.View {

	rel := rebac.Triple{
		Relation: rebac.Relation(strings.TrimPrefix(wnd.Values()[navsplitview.KindDetail.QueryKey("")], relUsrPrefix)),
	}

	var triples []rebac.Triple
	err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
		r, err := xslices.Collect2(rdb.Query(rebac.Select().Where().Source().Set(src).GroupByRelation()))
		if err != nil {
			return err
		}

		slices.SortFunc(r, func(a, b rebac.Triple) int {
			return strings.Compare(string(a.Relation), string(b.Relation))
		})

		triples = r
		return nil
	})

	if err != nil {
		return alert.BannerError(err)
	}

	return ui.ScrollView(
		dataview.FromData(wnd, dataview.Data[rebac.Triple, string]{
			ID: "declaredRels",
			FindAll: func(yield func(string, error) bool) {
				for _, triple := range triples {
					if !yield(triple.Identity(), nil) {
						return
					}
				}
			},
			FindByID: func(id string) (option.Opt[rebac.Triple], error) {
				for _, triple := range triples {
					if triple.String() == id {
						return option.Some(triple), nil
					}
				}

				return option.None[rebac.Triple](), nil
			},
			Fields: []dataview.Field[rebac.Triple]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj rebac.Triple) core.View {
						if p, ok := permission.Find(permission.ID(obj.Relation)); ok {
							return ui.Text(wnd.Bundle().Resolve(p.Name)).WordBreak(ui.WordBreakBreakAll)
						}

						return ui.Text(string(obj.Relation))
					},
				},
				{
					ID:   "desc",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj rebac.Triple) core.View {
						if p, ok := permission.Find(permission.ID(obj.Relation)); ok {
							return ui.Text(wnd.Bundle().Resolve(p.Description)).WordBreak(ui.WordBreakBreakAll)
						}

						return nil
					},
				},
			},
		}).Search(true).
			Selection(false).
			Action(func(e rebac.Triple) {
				navsplitview.NavigateDetail(wnd, "", navsplitview.ViewID(relUsrPrefix+e.Relation))
			}).
			NextActionIndicator(true).
			ListOptions(dataview.ListOptions[string]{
				ColorBody:    option.Some(ui.Color("")),
				ColorCaption: option.Some(ui.Color("")),
				ColorFooter:  option.Some(ui.Color("")),
				Highlight: map[string]bool{
					rel.String(): true,
				},
			}).
			Style(dataview.List),
	).Axis(ui.ScrollViewAxisVertical).FullWidth()
}

func allRelations(wnd core.Window, res rebac.Resources, id rebac.Instance) core.View {
	/*triples := slices.Collect(res.Relations(wnd.Context(), id))
	return ui.ScrollView(
		dataview.FromData(wnd, dataview.Data[rebac.Triple, string]{
			FindAll: func(yield func(string, error) bool) {
				for _, triple := range triples {
					if !yield(triple.Identity(), nil) {
						return
					}
				}
			},
			FindByID: func(id string) (option.Opt[rebac.Triple], error) {
				for _, triple := range triples {
					if triple.String() == id {
						return option.Some(triple), nil
					}
				}

				return option.None[rebac.Triple](), nil
			},
			Fields: []dataview.Field[rebac.Triple]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj rebac.Triple) core.View {
						if p, ok := permission.Find(permission.ID(obj.Relation)); ok {
							return ui.Text(wnd.Bundle().Resolve(p.Name)).WordBreak(ui.WordBreakBreakAll)
						}

						return ui.Text(string(obj.Relation))
					},
				},
				{
					ID:   "desc",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj rebac.Triple) core.View {
						if p, ok := permission.Find(permission.ID(obj.Relation)); ok {
							return ui.Text(wnd.Bundle().Resolve(p.Description)).WordBreak(ui.WordBreakBreakAll)
						}

						return nil
					},
				},
			},
		}).Search(true).
			Selection(false).
			Action(func(e rebac.Triple) {
				navsplitview.NavigateDetail(wnd, "", navsplitview.ViewID(relUsrPrefix+e.Relation))
			}).
			NextActionIndicator(true).
			ListOptions(dataview.ListOptions{
				ColorBody:    option.Some(ui.Color("")),
				ColorCaption: option.Some(ui.Color("")),
				ColorFooter:  option.Some(ui.Color("")),
			}).
			Style(dataview.List),
	).Axis(ui.ScrollViewAxisVertical).FullWidth()*/
	return nil
}

func allSourceEntries(wnd core.Window, res rebac.Resources) core.View {
	rid := rebac.Instance(strings.TrimPrefix(wnd.Values()[navsplitview.KindContent.QueryKey("")], usrRelPrefix))
	selected := rebac.NewInfoID(res.Identity(), rid)

	return ui.ScrollView(
		dataview.FromData(wnd, dataview.Data[rebac.InstanceInfo, rebac.InfoID]{
			ID:      "declaredSources",
			FindAll: res.All(wnd.Context()),
			FindByID: func(id rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
				return res.FindByID(wnd.Context(), id)
			},
			Fields: []dataview.Field[rebac.InstanceInfo]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj rebac.InstanceInfo) core.View {
						return ui.Text(obj.Name).WordBreak(ui.WordBreakBreakAll)
					},
				},
				{
					ID:   "desc",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj rebac.InstanceInfo) core.View {
						return ui.Text(obj.Description).WordBreak(ui.WordBreakBreakAll)
					},
				},
			},
		}).Search(true).
			Selection(false).
			Action(func(e rebac.InstanceInfo) {
				navsplitview.NavigateContent(wnd, "", navsplitview.ViewID(usrRelPrefix+e.ID))
			}).
			NextActionIndicator(true).
			ListOptions(dataview.ListOptions[rebac.InfoID]{
				ColorBody:    option.Some(ui.Color("")),
				ColorCaption: option.Some(ui.Color("")),
				ColorFooter:  option.Some(ui.Color("")),
				Highlight: map[rebac.InfoID]bool{
					selected: true,
				},
			}).
			Style(dataview.List),
	).Axis(ui.ScrollViewAxisVertical)
}
