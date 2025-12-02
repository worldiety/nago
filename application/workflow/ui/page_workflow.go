// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiworkflow

import (
	"os"

	"go.wdy.de/nago/application/workflow"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/hero"
	"go.wdy.de/nago/presentation/ui/tabs"
)

func PageWorkflow(wnd core.Window, uc workflow.UseCases) core.View {
	id := workflow.ID(wnd.Values()["id"])
	idx := tabs.NewIndexState(wnd, "pager-index")

	optWf, err := uc.FindDeclaredWorkflow(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optWf.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	wf := optWf.Unwrap()

	return ui.VStack(
		hero.Hero(wf.Name).Subtitle(wf.Description).SideSVG(icons.CodeMerge),
		ui.Space(ui.L32),
		tabs.Tabs(
			tabs.Page("Spezifikation", func() core.View {
				return specPage(wnd, uc, id)
			}),
			tabs.Page("Instanzen", func() core.View {
				return specInstances(wnd, uc, id)
			}),
		).InputValue(idx).
			FullWidth(),
	).
		FullWidth()
}
