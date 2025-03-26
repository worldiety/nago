// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"os"
)

func loadMyCircle(wnd core.Window, useCases usercircle.UseCases) (usercircle.Circle, error) {
	id := usercircle.ID(wnd.Values()["circle"])
	optCircle, err := useCases.FindByID(wnd.Subject(), id)
	if err != nil {
		return usercircle.Circle{}, err
	}

	if optCircle.IsNone() {
		return usercircle.Circle{}, os.ErrNotExist
	}

	circle := optCircle.Unwrap()
	return circle, nil
}
