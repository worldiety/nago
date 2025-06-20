// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidataimport

import (
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/picker"
	"maps"
	"path"
	"slices"
)

func DialogFieldMapping(wnd core.Window, presented *core.State[bool], imp importer.Importer, stage dataimport.Staging, exampleData []dataimport.Entry, uc dataimport.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	transformation := core.AutoState[dataimport.Transformation](wnd).Init(func() dataimport.Transformation {
		return stage.Transformation
	})

	return alert.Dialog(
		"Feld-Transformation",
		ViewFieldMapping(wnd, imp, stage, transformation, exampleData, uc),
		presented,
		alert.Larger(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := uc.UpdateStagingTransformation(wnd.Subject(), stage.ID, transformation.Get()); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		}),
	)
}

func ViewFieldMapping(wnd core.Window, imp importer.Importer, stage dataimport.Staging, transformation *core.State[dataimport.Transformation], exampleData []dataimport.Entry, uc dataimport.UseCases) core.View {
	if imp.Configuration().Passthrough {
		return ui.Text("Der Importer erhält die unveränderten Originaldaten. Eine Transformation ist nicht erforderlich.")
	}

	stub := importer.Stub(imp.Configuration().ExpectedType)
	stubFields := determineStubFields(stub)

	return ui.VStack(
		ui.Text("Definition der Abbildung der geparsten Daten in die Import-Struktur."),

		ui.Grid(
			slices.Collect(func(yield func(cell ui.TGridCell) bool) {
				yield(ui.GridCell(ui.HStack(ui.Text("Quell-Feld"))))
				yield(ui.GridCell(ui.HStack(ui.ImageIcon(icons.ArrowRight))))
				yield(ui.GridCell(ui.HStack(ui.Text("Ziel-Feld"))))

				for _, ptr := range determineFieldsFromOrigin(exampleData) {
					pickerState := core.StateOf[[]jsonptr.Ptr](wnd, "picker-"+ptr).Init(func() []jsonptr.Ptr {
						if rule, ok := transformation.Get().RuleBySrc(ptr); ok {
							return []jsonptr.Ptr{rule.DstKey}
						}

						return nil
					}).Observe(func(newValue []jsonptr.Ptr) {
						t := transformation.Get()
						t.CopyRules = slices.DeleteFunc(t.CopyRules, func(rule dataimport.CopyRule) bool {
							return rule.SrcKey == ptr
						})

						for _, dst := range newValue {
							t.CopyRules = append(t.CopyRules, dataimport.CopyRule{
								SrcKey: ptr,
								DstKey: dst,
							})
						}

						transformation.Set(t)
						transformation.Notify()
					})

					yield(ui.GridCell(ui.TextField("", ptr).Disabled(true)).Padding(ui.Padding{Bottom: ui.L8}))
					yield(ui.GridCell(ui.HStack(ui.ImageIcon(icons.ArrowRight))))
					yield(ui.GridCell(picker.Picker[jsonptr.Ptr]("", stubFields, pickerState)))
				}

			})...,
		).Columns(3).FullWidth().Widths("1fr", ui.L32, "1fr"),
	).FullWidth().
		Alignment(ui.Leading).
		Gap(ui.L32)
}

func determineStubFields(stub *jsonptr.Obj) []jsonptr.Ptr {
	tmp := map[jsonptr.Ptr]bool{}
	insertKeys("/", tmp, stub)
	return slices.Sorted(maps.Keys(tmp))
}

func determineFieldsFromOrigin(sampleData []dataimport.Entry) []jsonptr.Ptr {
	tmp := map[jsonptr.Ptr]bool{}

	for _, d := range sampleData {
		insertKeys("/", tmp, d.In)
	}

	return slices.Sorted(maps.Keys(tmp))
}

func insertKeys(parent jsonptr.Ptr, dst map[jsonptr.Ptr]bool, src *jsonptr.Obj) {
	for key, val := range src.All() {
		vk := path.Join(parent, key)
		dst[vk] = true
		if obj, ok := val.(*jsonptr.Obj); ok {
			insertKeys(vk, dst, obj)
		}
	}
}
