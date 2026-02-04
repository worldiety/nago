// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"os"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

func PageShareViewer(wnd core.Window, loader flow.LoadWorkspace, loadShare flow.FindFormShareByID) core.View {
	shareId := flow.FormShareID(wnd.Values()["share"])
	optShare, err := loadShare(wnd.Subject(), shareId)
	if err != nil {
		return alert.BannerError(err)
	}

	if optShare.IsNone() {
		return alert.BannerError(fmt.Errorf("share not found: %s: %w", shareId, os.ErrNotExist))
	}

	share := optShare.Unwrap()

	state := core.StateOf[*jsonptr.Obj](wnd, string(shareId)).Init(func() *jsonptr.Obj {
		return jsonptr.NewObj(map[string]jsonptr.Value{})
	})

	return FormViewer(user.SU(), loader, share.Workspace, share.Form, state)
}
