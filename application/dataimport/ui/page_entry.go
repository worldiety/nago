// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"fmt"
	"github.com/worldiety/enum/json"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/progress"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

func PageEntry(wnd core.Window, ucImp dataimport.UseCases) core.View {
	sid := dataimport.SID(wnd.Values()["stage"])
	optStage, err := ucImp.FindStagingByID(wnd.Subject(), sid)
	if err != nil {
		return alert.BannerError(err)
	}

	if optStage.IsNone() {
		return alert.BannerError(fmt.Errorf("stage not found: %w", os.ErrNotExist))
	}

	stage := optStage.Unwrap()

	optImp, err := ucImp.FindImporterByID(wnd.Subject(), stage.Importer)
	if err != nil {
		return alert.BannerError(err)
	}

	if optImp.IsNone() {
		return alert.BannerError(fmt.Errorf("importer not found: %w", os.ErrNotExist))
	}

	imp := optImp.Unwrap()

	optEntry, err := ucImp.FindEntryByID(wnd.Subject(), dataimport.Key(wnd.Values()["entry"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optEntry.IsNone() {
		return alert.BannerError(fmt.Errorf("entry not found: %w", os.ErrNotExist))
	}

	entry := optEntry.Unwrap()
	entryState := core.AutoState[any](wnd).Init(func() any {
		actualValue := reflect.New(imp.Configuration().ExpectedType).Interface()
		transformed := entry.Transform(stage.Transformation)
		if err := json.Unmarshal([]byte(transformed.String()), &actualValue); err != nil {
			alert.ShowBannerError(wnd, err)
		}

		return reflect.ValueOf(actualValue).Elem().Interface()
	})

	return ui.VStack(
		ui.H1("Prüfung Eintrag"),

		toolbar(wnd),
		ui.Space(ui.L32),
		ui.Grid(
			ui.GridCell(
				ui.VStack(viewSrc(wnd, imp, stage, entry.In)).Alignment(ui.TopLeading).
					Padding(ui.Padding{Right: ui.L32}),
			),
			ui.GridCell(
				ui.VStack(
					ui.H2("Ziel"),
					form.Auto(form.AutoOptions{Window: wnd}, entryState),
				).Alignment(ui.TopLeading).
					BackgroundColor(ui.ColorCardBody).
					Border(ui.Border{}.Radius(ui.L16)).
					Padding(ui.Padding{}.All(ui.L16)),
			),
		).Columns(2),
	).
		FullWidth().
		Alignment(ui.Leading)

}

func toolbar(wnd core.Window) core.View {
	return ui.VStack(
		ui.VStack(
			ui.HStack(ui.Text(fmt.Sprintf("?/? geprüft"))).Alignment(ui.Trailing).FullWidth(),
			progress.LinearProgress().Progress(0.3).FullWidth(),
			ui.HStack(
				ui.TertiaryButton(func() {

				}).Title("Vorheriger").PreIcon(flowbiteOutline.ChevronLeft),
				ui.TertiaryButton(func() {

				}).Title("Nächster").PostIcon(flowbiteOutline.ChevronRight),
				ui.Spacer(),
				ui.SecondaryButton(func() {

				}).Title("Ablehnen"),
				ui.PrimaryButton(func() {

				}).Title("Bestätigen"),
			).FullWidth().Gap(ui.L8),
		).Gap(ui.L8).
			FullWidth().
			BackgroundColor(ui.ColorCardBody).
			Border(ui.Border{}.Radius(ui.L16).Shadow(ui.L16)).
			Padding(ui.Padding{}.All(ui.L16)),
	).FullWidth().
		Position(ui.Position{
			Type: ui.PositionSticky,
			Top:  "6rem", // height navbar
			//Bottom: "6rem", // height of the footer
			ZIndex: 10,
		}).
		Padding(ui.Padding{}.All(ui.L32))
}

func viewSrc(wnd core.Window, imp importer.Importer, stage dataimport.Staging, entry *jsonptr.Obj) core.View {
	fields := determineStubFields(entry)
	viewRawMode := core.AutoState[bool](wnd)

	var styleBtnFields ui.StylePreset
	var styleBtnRaw ui.StylePreset

	if viewRawMode.Get() {
		styleBtnFields = ui.StyleButtonSecondary
		styleBtnRaw = ui.StyleButtonPrimary
		fields = nil
	} else {
		styleBtnFields = ui.StyleButtonPrimary
		styleBtnRaw = ui.StyleButtonSecondary
	}

	showHiddenFields := core.AutoState[bool](wnd)

	hiddenFields := 0
	return ui.VStack(
		ui.HStack(
			ui.H2("Quelle"),
			ui.Spacer(),
			ui.TertiaryButton(func() {
				viewRawMode.Set(false)
			}).Title("Ansicht Felder").Preset(styleBtnFields),
			ui.SecondaryButton(func() {
				viewRawMode.Set(true)
			}).Title("Ansicht Rohdaten").Preset(styleBtnRaw),
		).FullWidth().
			Alignment(ui.Trailing).
			Gap(ui.L8),
	).
		Append(
			ui.IfFunc(viewRawMode.Get(), func() core.View {
				return ui.ScrollView(
					ui.CodeEditor(entry.String()).
						Disabled(true).
						Language("json")).
					Frame(ui.Frame{}.FullWidth()).
					Axis(ui.ScrollViewAxisBoth)
			}),
		).
		Append(
			ui.ForEach(fields, func(t jsonptr.Ptr) core.View {

				if _, ok := stage.Transformation.RuleBySrc(t); ok {
					hiddenFields++
					if !showHiddenFields.Get() {
						return nil
					}
				}

				val, err := jsonptr.Eval(entry, t)
				if err != nil {
					slog.Error("failed to eval ptr", "err", err.Error())
				}
				if val == nil {
					val = jsonptr.Null{}
				}

				label := strings.TrimPrefix(t, "/")

				return ui.TextField(label, val.String()).Disabled(true).FullWidth()
			})...,
		).
		Append(ui.IfFunc(len(fields) > 0, func() core.View {
			return ui.SecondaryButton(func() {
				showHiddenFields.Set(!showHiddenFields.Get())
			}).Title(fmt.Sprintf("%d Feld(er) ein/ausblenden", hiddenFields))
		})).
		FullWidth().
		Gap(ui.L16).
		Border(ui.Border{}.Radius(ui.L16).Color(ui.ColorCardBody).Width(ui.L1)).
		Padding(ui.Padding{}.All(ui.L16))

}
