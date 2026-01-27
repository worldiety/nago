// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func PageFormViewerCreate(wnd core.Window, loader flow.LoadWorkspace, stores blob.Stores) core.View {
	wsId := flow.WorkspaceID(wnd.Values()["workspace"])
	formId := flow.FormID(wnd.Values()["form"])

	state := core.StateOf[*jsonptr.Obj](wnd, string(wsId)+"-"+string(formId)).Init(func() *jsonptr.Obj {
		return jsonptr.NewObj(map[string]jsonptr.Value{})
	})

	return ui.VStack(
		FormViewer(loader, wsId, formId, state),
		ui.HLine(),
		ui.HStack(
			ui.PrimaryButton(func() {
				fmt.Println("TODO store", state.Get().String())
			}).Title(rstring.ActionCreate.Get(wnd)),
		).FullWidth().Alignment(ui.Trailing),
	).FullWidth()

}
