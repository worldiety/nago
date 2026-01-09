// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uievs

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/worldiety/i18n/date"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xmaps"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
)

type PageAuditOptions[Evt any] struct {
	Perms        evs.Permissions
	EntityName   string
	Pages        Pages
	Prefix       permission.ID
	Audit        func(wnd core.Window, uc evs.UseCases[Evt]) core.View
	DecorateView func(wnd core.Window, state *core.State[Evt], view core.View) core.View
	Indexer      []evs.Indexer[Evt]
}

func PageAudit[Evt any](wnd core.Window, uc evs.UseCases[Evt], opts PageAuditOptions[Evt]) core.View {
	if opts.Audit == nil {
		opts.Audit = newDefaultAudit(wnd, opts)
	}

	return opts.Audit(wnd, uc)
}

func newDefaultAudit[Evt any](wnd core.Window, opts PageAuditOptions[Evt]) func(wnd core.Window, uc evs.UseCases[Evt]) core.View {

	eventPresented := core.StateOf[bool](wnd, string(opts.Prefix)+"evt-presented")
	eventSelected := core.StateOf[evs.Envelope[Evt]](wnd, string(opts.Prefix)+"evt-selected")

	return func(wnd core.Window, uc evs.UseCases[Evt]) core.View {
		displayName, _ := core.FromContext[user.DisplayName](wnd.Context(), "")
		var indexer evs.Indexer[Evt]
		filter := wnd.Values()["primary"]
		indexId := wnd.Values()["indexer"]

		for _, e := range opts.Indexer {
			if e.Info().ID == evs.IdxID(indexId) {
				indexer = e
				break
			}
		}

		dv := dataview.FromData(wnd, dataview.Data[evs.Envelope[Evt], evs.SeqKey]{
			FindAll: func(yield func(evs.SeqKey, error) bool) {
				if filter == "" {
					for key, err := range uc.ReadAll(wnd.Subject()) {
						if !yield(key, err) {
							return
						}
					}

					return
				}

				if indexer == nil {
					return
				}

				for key, err := range indexer.GroupByPrimary(filter) {
					if !yield(key, err) {
						return
					}
				}
			},
			FindByID: func(id evs.SeqKey) (option.Opt[evs.Envelope[Evt]], error) {
				seq, err := id.Parse()
				if err != nil {
					return option.None[evs.Envelope[Evt]](), fmt.Errorf("invalid event id: %w", err)
				}

				return uc.Load(wnd.Subject(), seq)
			},
			Fields: []dataview.Field[evs.Envelope[Evt]]{
				{
					ID:   "seqid",
					Name: "ID",
					Map: func(obj evs.Envelope[Evt]) core.View {
						return ui.Text(strconv.Itoa(int(obj.Sequence))).AccessibilityLabel(string(obj.Key))
					},
				},
				{
					ID:   "type",
					Name: rstring.LabelType.Get(wnd),
					Map: func(obj evs.Envelope[Evt]) core.View {
						return ui.Text(string(obj.Discriminator))
					},
				},
				{
					ID:   "user",
					Name: rstring.LabelUser.Get(wnd),
					Map: func(obj evs.Envelope[Evt]) core.View {
						return ui.Text(displayName(obj.CreatedBy).Displayname)
					},
				},

				{
					ID:   "time",
					Name: rstring.LabelCreatedAt.Get(wnd),
					Map: func(obj evs.Envelope[Evt]) core.View {
						return ui.Text(date.Format(wnd.Locale(), date.Time, obj.EventTime.Time(wnd.Location())))
					},
				},
			},
		}).Action(func(e evs.Envelope[Evt]) {
			eventPresented.Set(true)
			eventSelected.Set(e)
		}).NextActionIndicator(true)

		if wnd.Subject().HasPermission(opts.Perms.Store) {
			var copts []dataview.CreateOption
			for t := range uc.RegisteredTypes() {
				copts = append(copts, dataview.CreateOption{
					Name: string(t.Discriminator),
					Action: func() error {
						wnd.Navigation().ForwardTo(opts.Pages.Create.Join(string(t.Discriminator)), wnd.Values())
						return nil
					},
				})
			}

			dv = dv.CreateOptions(copts...)
		}

		if wnd.Subject().HasPermission(opts.Perms.Delete) {
			dv = dv.SelectOptions(dataview.NewSelectOptionDelete(wnd, func(selected []evs.SeqKey) error {
				for _, id := range selected {
					id, err := id.Parse()
					if err != nil {
						return err
					}
					if err := uc.Delete(wnd.Subject(), id); err != nil {
						return err
					}
				}

				return nil
			}))
		}

		pageTitle := opts.EntityName
		if indexer != nil {
			pageTitle = fmt.Sprintf("%s / %s / %s", opts.EntityName, indexer.Info().Name, filter)
		}

		return ui.VStack(
			ui.Space(ui.L16),
			eventDetailsDialog(wnd, eventPresented, eventSelected),
			breadcrumb.Breadcrumbs(
				ui.TertiaryButton(func() {
					wnd.Navigation().BackwardTo("admin", wnd.Values().Put("#", string(opts.Prefix)))
				}).Title(StrDataManagement.Get(wnd)),
			).ClampLeading(),
			ui.H1(pageTitle),
			dv,
		).FullWidth().Alignment(ui.Leading)
	}
}

func eventDetailsDialog[Evt any](wnd core.Window, presented *core.State[bool], evt *core.State[evs.Envelope[Evt]]) core.View {
	if !presented.Get() {
		return nil
	}

	var raw string
	buf, err := json.MarshalIndent(evt.Get().Data, "", "  ")
	if err != nil {
		slog.Error("failed to marshal event details for detail view", "err", err.Error())
		raw = string(evt.Get().Raw)
	} else {
		raw = string(buf)
	}

	var rows []ui.TTableRow
	for k, v := range xmaps.All(evt.Get().Metadata) {
		rows = append(rows, ui.TableRow(
			ui.TableCell(ui.Text(k)),
			ui.TableCell(ui.Text(v)),
		))
	}

	var pkgName string
	var tName string
	if t := reflect.TypeOf(evt.Get().Data); t != nil {
		tName = t.Name()
		pkgName = t.PkgPath()
	}

	ns := wnd.Path().Dir().Base()

	return alert.Dialog(
		rstring.LabelDetails.Get(wnd),
		ui.VStack(
			ui.Table(
				ui.TableColumn(ui.Text(rstring.LabelProperty.Get(wnd))),
				ui.TableColumn(ui.Text(rstring.LabelValue.Get(wnd))),
			).Rows(
				ui.TableRow(
					ui.TableCell(ui.Text("Namespace")),
					ui.TableCell(ui.Text(ns)),
				),

				ui.TableRow(
					ui.TableCell(ui.Text("SeqNo")),
					ui.TableCell(ui.Text(strconv.Itoa(int(evt.Get().Sequence)))),
				),

				ui.TableRow(
					ui.TableCell(ui.Text("Key")),
					ui.TableCell(ui.Text(string(evt.Get().Key))),
				),
				ui.TableRow(
					ui.TableCell(ui.Text("Alias")),
					ui.TableCell(ui.Text(fmt.Sprintf("%s", evt.Get().Discriminator))),
				),

				ui.TableRow(
					ui.TableCell(ui.Text("Type")),
					ui.TableCell(ui.Text(fmt.Sprintf("%s.%s", pkgName, tName))),
				),

				ui.TableRow(
					ui.TableCell(ui.Text("EventTime")),
					ui.TableCell(ui.Text(date.Format(wnd.Locale(), date.Time, evt.Get().EventTime.Time(wnd.Location())))),
				),

				ui.TableRow(
					ui.TableCell(ui.Text("Unix Time")),
					ui.TableCell(ui.Text(strconv.Itoa(int(evt.Get().EventTime)))),
				),

				ui.TableRow(
					ui.TableCell(ui.Text("CreatedBy")),
					ui.TableCell(ui.Text(string(evt.Get().CreatedBy))),
				),
			).Rows(rows...).
				Frame(ui.Frame{}.FullWidth()),
			ui.CodeEditor(raw).Disabled(true).Language("json").FullWidth(),
		).FullWidth(),
		presented,
		alert.Closeable(),
		alert.Close(nil),
		alert.Larger(),
	)
}
