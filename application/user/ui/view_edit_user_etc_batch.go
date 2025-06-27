// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"time"
)

func dialogEtcBatch(wnd core.Window, ucUsers user.UseCases, presented *core.State[bool], uids []user.ID) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		fmt.Sprintf("Stapelverarbeitung für %d Konten", len(uids)),
		viewEtcBatch(wnd, ucUsers, uids),
		presented,
		alert.Larger(),
		alert.Closeable(),
		alert.Close(nil),
	)
}

func viewEtcBatch(wnd core.Window, ucUsers user.UseCases, uids []user.ID) core.View {
	return ui.VStack(
		etcActionNotifyUser(wnd, func() {
			bus := wnd.Application().EventBus()
			for _, uid := range uids {
				optUsr, err := ucUsers.FindByID(wnd.Subject(), uid)
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
				if optUsr.IsNone() {
					// stale ref
					continue
				}

				usr := optUsr.Unwrap()
				user.PublishUserCreated(bus, usr, true)
			}

			alert.ShowBannerMessage(wnd, alert.Message{
				Title:    "Nutzer erstellt",
				Message:  fmt.Sprintf("Ereignis für %d Nutzer erstellt.", len(uids)),
				Intent:   alert.IntentOk,
				Duration: time.Second * 2,
			})
		}),

		etcActionDeleteUser(wnd, func() {
			for _, uid := range uids {
				if err := ucUsers.Delete(wnd.Subject(), uid); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
			}
		}),
		
	).FullWidth().Gap(ui.L32)
}
