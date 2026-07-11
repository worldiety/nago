// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfginspector

import (
	"log/slog"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/inspector"
	inspectorndb "go.wdy.de/nago/application/inspector/ndb"
	uindbinspector "go.wdy.de/nago/application/inspector/ndb/ui"
	"go.wdy.de/nago/application/inspector/rest"
	"go.wdy.de/nago/application/inspector/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// Management is a Nago system(Inspector Management).
// It provides functionality to inspect and manage entity and blob stores, and –
// when an ndb database has been configured – to inspect ndb message streams
// (msgstore) and time series (tsdb).
// The system allows users to view and edit repository entries, download and delete blob files,
// and interact with stores through the Admin Center UI.
type Management struct {
	UseCases inspector.UseCases
	Pages    uiinspector.Pages

	// NDB holds the ndb inspector wiring. It is always installed, but its admin
	// card only appears once at least one ndb database is registered.
	NDBUseCases inspectorndb.UseCases
	NDBPages    uindbinspector.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	stores, err := cfg.Stores()
	if err != nil {
		return Management{}, err
	}

	// The ndb inspector browses every ndb database registered with the
	// Configurator (via cfg.NDB / cfg.OpenNDB). The provider is live so
	// databases opened after Enable still appear.
	ndbProvider := func() []inspectorndb.Instance {
		var out []inspectorndb.Instance
		for _, in := range cfg.NDBInstances() {
			out = append(out, inspectorndb.Instance{Path: in.Path, Name: in.Name, DB: in.DB})
		}
		return out
	}

	management = Management{
		UseCases: inspector.NewUseCases(stores),
		Pages: uiinspector.Pages{
			PageDataInspector: "admin/inspector",
		},
		NDBUseCases: inspectorndb.NewUseCases(ndbProvider),
		NDBPages: uindbinspector.Pages{
			PageMessages:   "admin/inspector/ndb/messages",
			PageTimeseries: "admin/inspector/ndb/timeseries",
		},
	}

	cfg.NoFooter(management.Pages.PageDataInspector)
	cfg.NoFooter(management.NDBPages.PageMessages)
	cfg.NoFooter(management.NDBPages.PageTimeseries)

	cfg.RootViewWithDecoration(management.Pages.PageDataInspector, func(wnd core.Window) core.View {
		return uiinspector.PageInspector(wnd, management.UseCases)
	})

	cfg.RootViewWithDecoration(management.NDBPages.PageMessages, func(wnd core.Window) core.View {
		return uindbinspector.PageMessages(wnd, management.NDBUseCases)
	})

	cfg.RootViewWithDecoration(management.NDBPages.PageTimeseries, func(wnd core.Window) core.View {
		return uindbinspector.PageTimeseries(wnd, management.NDBUseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		group := admin.Group{
			Title: "Inspektor",
			Entries: []admin.Card{
				{Title: "Stores", Text: "Stores bilden die Grundlage für Repositories. Es gibt spezialisierte Stores für Entities und Blobs.", Target: management.Pages.PageDataInspector, Permission: inspector.PermDataInspector},
			},
		}

		// The ndb cards appear automatically once an ndb database has been
		// configured anywhere in the application.
		if len(cfg.NDBInstances()) > 0 {
			group.Entries = append(group.Entries,
				admin.Card{
					Title:      "ndb Nachrichten",
					Text:       "Message-Streams (msgstore) über alle ndb Datenbanken einsehen: nach Seq oder Zeit springen, Nachrichten fensterweise anzeigen und einzelne Einträge oder ganze Streams löschen.",
					Target:     management.NDBPages.PageMessages,
					Permission: inspectorndb.PermNDBInspector,
				},
				admin.Card{
					Title:      "ndb Zeitreihen",
					Text:       "Zeitreihen (tsdb) über alle ndb Datenbanken einsehen: Spalte und Zeitbereich wählen, als M4-Chart darstellen und einzelne Punkte oder Bereiche löschen bzw. kompaktieren.",
					Target:     management.NDBPages.PageTimeseries,
					Permission: inspectorndb.PermNDBInspector,
				},
			)
		}

		return group
	})
	cfg.AddContextValue(core.ContextValue("nago.inspector", management))

	if err := cfg.HandleFuncSubject(rest.PathDownloadAsJSONArray, rest.NewDownloadAsJSONArray(stores)); err != nil {
		return Management{}, err
	}

	if err := cfg.HandleFuncSubject(rest.PathDownloadAsJSONObject, rest.NewDownloadAsJSONObject(stores)); err != nil {
		return Management{}, err
	}

	if err := cfg.HandleFuncSubject(rest.PathDownloadAsZip, rest.NewDownloadAsZip(stores)); err != nil {
		return Management{}, err
	}

	if err := cfg.HandleFuncSubject(rest.PathDownloadAsRaw, rest.NewDownloadAsRaw(stores)); err != nil {
		return Management{}, err
	}

	slog.Info("installed inspector management")

	return management, nil
}
