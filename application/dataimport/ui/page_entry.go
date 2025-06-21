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

		ui.HStack(
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/data/stagings", core.Values{"importer": string(stage.Importer)})
			}).Title("Import "+imp.Configuration().Name),
			ui.ImageIcon(flowbiteOutline.ChevronRight),
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/data/staging", core.Values{"stage": string(stage.ID)})
			}).Title(stage.Name),
			ui.ImageIcon(flowbiteOutline.ChevronRight),
			ui.TertiaryButton(nil).Title("Eintrag prüfen"),
		).Alignment(ui.Leading),
		toolbar(wnd, ucImp, stage, entry, entryState),
		ui.Space(ui.L32),
		ui.Grid(
			ui.GridCell(
				ui.VStack(viewSrc(wnd, stage, entry.In)).Alignment(ui.TopLeading).
					Padding(ui.Padding{Right: ui.L32}),
			),
			ui.GridCell(
				ui.VStack(
					ui.H2("Ziel"),
					form.Auto(form.AutoOptions{Window: wnd, ViewOnly: entry.Confirmed || entry.Imported || entry.Ignored}, entryState),
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

func toolbar(wnd core.Window, ucImp dataimport.UseCases, stage dataimport.Staging, entry dataimport.Entry, obj *core.State[any]) core.View {
	stat, err := ucImp.CalculateStagingReviewStatus(wnd.Subject(), stage.ID)
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.VStack(
		ui.VStack(
			ui.HStack(ui.Text(fmt.Sprintf("%d/%d geprüft", stat.Checked(), stat.Total))).Alignment(ui.Trailing).FullWidth(),
			progress.LinearProgress().Progress(float64(stat.Checked())/float64(stat.Total)).FullWidth(),
			ui.HStack(
				ui.TertiaryButton(func() {

				}).Title("Vorheriger").PreIcon(flowbiteOutline.ChevronLeft),
				ui.TertiaryButton(func() {

				}).Title("Nächster").PostIcon(flowbiteOutline.ChevronRight),
				ui.Spacer(),

				ui.SecondaryButton(func() {
					if err := ucImp.UpdateEntryIgnored(wnd.Subject(), entry.ID, false); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					if err := ucImp.UpdateEntryConfirmation(wnd.Subject(), entry.ID, false); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					obj.Invalidate()

				}).Title("Erneut prüfen").Enabled((entry.Ignored || entry.Confirmed) && !entry.Imported),

				ui.SecondaryButton(func() {
					if err := ucImp.UpdateEntryIgnored(wnd.Subject(), entry.ID, true); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					if err := updateTransformation(wnd, ucImp, entry, obj); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				}).Title("Ablehnen").Enabled(!entry.Ignored),

				ui.PrimaryButton(func() {
					if err := ucImp.UpdateEntryConfirmation(wnd.Subject(), entry.ID, true); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					if err := updateTransformation(wnd, ucImp, entry, obj); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

				}).Title("Bestätigen").Enabled(!entry.Confirmed),
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

func viewSrc(wnd core.Window, stage dataimport.Staging, entry *jsonptr.Obj) core.View {
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

func updateTransformation(wnd core.Window, ucImp dataimport.UseCases, entry dataimport.Entry, obj *core.State[any]) error {
	buf, err := json.Marshal(obj.Get())
	if err != nil {
		return fmt.Errorf("cannot convert entry type model to intermediate json model: %w", err)
	}

	var tmp *jsonptr.Obj
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return fmt.Errorf("cannot convert intermediate json model to jsonptr.Obj: %w", err)
	}

	if err := ucImp.UpdateEntryTransformed(wnd.Subject(), entry.ID, tmp); err != nil {
		return fmt.Errorf("cannot update entry transformed model: %w", err)
	}

	obj.Invalidate()

	return nil
}
