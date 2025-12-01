// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package uienv

import (
	"fmt"

	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

func loadApp(wnd core.Window, uc environment.UseCases) (app.App, core.View) {
	var zero app.App
	envId := environment.ID(wnd.Values()["env"])
	appId := app.ID(wnd.Values()["app"])
	optApp, err := uc.FindAppByID(wnd.Subject(), envId, appId)
	if err != nil {
		return zero, alert.BannerError(err)
	}

	if optApp.IsNone() {
		return zero, alert.BannerError(fmt.Errorf("app not found %s.%s: %w", envId, appId, err))
	}

	return optApp.Unwrap(), nil
}
