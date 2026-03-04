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

	"github.com/worldiety/i18n"
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
	"go.wdy.de/nago/presentation/ui/picker"
	"golang.org/x/text/language"
)

var (
	StrNothingSelected = i18n.MustString("nago.rebac.editor.nothing_selected", i18n.Values{language.German: "Noch nichts gewählt.", language.English: "Nothing is selected."})
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
		case id == "_":
			return ui.Text(StrNothingSelected.Get(wnd))
		default:
			rid := rebac.Instance(strings.TrimPrefix(wnd.Values()[navsplitview.KindContent.QueryKey("")], usrRelPrefix))
			rel := rebac.Relation(strings.TrimPrefix(wnd.Values()[navsplitview.KindDetail.QueryKey("")], relUsrPrefix))

			return declaredTargets(wnd, uc, rebac.Select().Where().Source().Is(res.Identity(), rid).Where().Relation().Has(rel))
		}
	})).Default("source", "_", "_").
		AlignmentSidebar(ui.Stretch).
		AlignmentContent(ui.Stretch).
		AlignmentDetail(ui.Stretch).
		Frame(ui.Frame{Height: "calc(100vh - 6rem)", Width: ui.Full})

}

// declaredTargets is the after source and relation has been defined. This is the content in the most right
// detail view.
func declaredTargets(wnd core.Window, uc ucrebac.UseCases, query rebac.Query) core.View {
	var targets []rebac.InstanceInfo

	createPresented := core.AutoState[bool](wnd)

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
		ui.VStack(
			dialogAddTarget(wnd, uc, createPresented, query.AsTriple().Source, query.AsTriple().Relation),
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
						createPresented.Set(true)
						return nil
					},
				}).
				SelectOptions(dataview.NewSelectOptionDelete(wnd, func(selected []rebac.InfoID) error {
					return uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
						for _, iid := range selected {
							ns, id, err := iid.Parse()
							if err != nil {
								return err
							}

							err = rdb.DeleteByQuery(rebac.Select().
								Where().Source().Set(query.AsTriple().Source).
								Where().Relation().Has(query.AsTriple().Relation).
								Where().Target().Is(ns, id),
							)

							if err != nil {
								return err
							}
						}

						return nil
					})
				})).
				Selection(true),
		).FullWidth(),
	).Axis(ui.ScrollViewAxisVertical).FullWidth()
}

func dialogAddTarget(wnd core.Window, uc ucrebac.UseCases, presented *core.State[bool], src rebac.Entity, rel rebac.Relation) core.View {
	if !presented.Get() {
		return nil
	}

	selectedRelation := core.AutoState[[]rebac.InstanceInfo](wnd).Init(func() []rebac.InstanceInfo {
		var sel []rebac.InstanceInfo
		if len(rel) == 0 {
			return sel
		}

		err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
			res, ok := rdb.LookupResources(rebac.Relations)
			if !ok {
				return nil
			}
			optInfo, err := res.FindByID(wnd.Context(), rebac.NewInfoID(rebac.Relations, rebac.Instance(rel)))
			if err != nil {
				return err
			}

			if optInfo.IsSome() {
				sel = append(sel, optInfo.Unwrap())
			}

			return nil
		})

		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return sel
	})
	selectedResources := core.AutoState[[]rebac.Resources](wnd)
	selectedInstance := core.AutoState[[]rebac.InstanceInfo](wnd)

	return alert.Dialog(
		rstring.ActionAdd.Get(wnd),
		ui.VStack(
			ui.H2("Relation"),
			picker.FromData[rebac.InstanceInfo, rebac.InfoID]("Relation", selectedRelation, picker.Data[rebac.InstanceInfo, rebac.InfoID]{
				FindAll: func(yield func(rebac.InfoID, error) bool) {
					err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
						for rule := range rdb.AllStaticRules() {
							if rule.Source != src.Namespace {
								continue
							}

							if !yield(rebac.NewInfoID(rebac.Relations, rebac.Instance(rule.Relation)), nil) {
								return nil
							}
						}

						return nil
					})

					if err != nil {
						yield("", err)
						return
					}
				},
				FindByID: func(id rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
					var optInfo option.Opt[rebac.InstanceInfo]
					err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
						res, ok := rdb.LookupResources(rebac.Relations)
						if !ok {
							return nil
						}
						oi, err := res.FindByID(wnd.Context(), id)
						optInfo = oi
						return err
					})

					return optInfo, err
				},
				Stringer: func(info rebac.InstanceInfo) string {
					return info.Name
				},
			}).FullWidth(),
			ui.H2("Target"),
			picker.FromData[rebac.Resources, rebac.Namespace]("Namespace", selectedResources, picker.Data[rebac.Resources, rebac.Namespace]{
				FindAll: func(yield func(rebac.Namespace, error) bool) {
					err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
						selRels := selectedRelation.Get()
						if len(selRels) == 0 {
							return nil
						}

						selRel := selRels[0]
						for resources := range rdb.AllResources() {
							allowed := false
							for rule := range rdb.AllStaticRules() {
								if rule.Source == src.Namespace && rule.Relation == rebac.Relation(selRel.ID) && rule.Target == resources.Identity() {
									allowed = true
									break
								}
							}

							if !allowed {
								continue
							}

							if !yield(resources.Identity(), nil) {
								return nil
							}
						}

						return nil
					})

					if err != nil {
						yield("", err)
						return
					}
				},
				FindByID: func(id rebac.Namespace) (option.Opt[rebac.Resources], error) {
					sel := selectedRelation.Get()
					if len(sel) == 0 {
						return option.None[rebac.Resources](), nil
					}

					var optRes option.Opt[rebac.Resources]
					err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
						v, ok := rdb.LookupResources(id)
						if ok {
							optRes = option.Some(v)
						}

						return nil
					})

					return optRes, err
				},
				Stringer: func(e rebac.Resources) string {
					return e.Info(wnd).Name
				},
				ID: "",
			}).FullWidth().Disabled(len(selectedRelation.Get()) == 0),
			picker.FromData[rebac.InstanceInfo, rebac.InfoID]("Instance", selectedInstance, picker.Data[rebac.InstanceInfo, rebac.InfoID]{
				FindAll: func(yield func(rebac.InfoID, error) bool) {
					res := selectedResources.Get()
					if len(res) == 0 {
						return
					}

					err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
						res, ok := rdb.LookupResources(res[0].Identity())
						if !ok {
							return nil
						}

						for id, err := range res.All(wnd.Context()) {
							if err != nil {
								return err
							}

							if !yield(id, nil) {
								return nil
							}
						}

						return nil
					})

					if err != nil {
						yield("", err)
						return
					}
				},
				FindByID: func(id rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
					res := selectedResources.Get()
					if len(res) == 0 {
						return option.None[rebac.InstanceInfo](), nil
					}

					return res[0].FindByID(wnd.Context(), id)
				},
				Stringer: func(info rebac.InstanceInfo) string {
					return info.Name
				},
				ID: "",
			}).Disabled(len(selectedResources.Get()) == 0).FullWidth(),
		).FullWidth().Gap(ui.L8).Alignment(ui.Leading),
		presented,
		alert.Closeable(),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				err := uc.WithReBAC(wnd.Subject(), func(rdb *rebac.DB) error {
					return rdb.Put(rebac.Triple{
						Source:   src,
						Relation: rebac.Relation(selectedRelation.Get()[0].ID),
						Target: rebac.Entity{
							Namespace: selectedResources.Get()[0].Identity(),
							Instance:  selectedInstance.Get()[0].ID,
						},
					})
				})
				if err != nil {
					alert.ShowBannerError(wnd, err)
					close(false)
					return
				}

				close(true)
			}).Title(rstring.ActionAdd.Get(wnd)).Enabled(len(selectedRelation.Get()) > 0 && len(selectedResources.Get()) > 0 && len(selectedInstance.Get()) > 0)
		}),
	)

}

func declaredRelations(wnd core.Window, uc ucrebac.UseCases, src rebac.Entity) core.View {
	rel := rebac.Triple{
		Relation: rebac.Relation(strings.TrimPrefix(wnd.Values()[navsplitview.KindDetail.QueryKey("")], relUsrPrefix)),
	}

	createPresented := core.AutoState[bool](wnd)

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
		ui.VStack(
			dialogAddTarget(wnd, uc, createPresented, src, ""),
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
								return ui.Text(wnd.Bundle().Resolve(p.Name)).AccessibilityLabel(wnd.Bundle().Resolve(p.Description))
							}

							return ui.Text(string(obj.Relation))
						},
					},
				},
			}).Search(true).
				CreateOptions(dataview.CreateOption{
					Name: rstring.ActionCreate.Get(wnd),
					Action: func() error {
						createPresented.Set(true)
						return nil
					},
				}).
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
		).FullWidth(),
	).Axis(ui.ScrollViewAxisVertical).FullWidth()

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
