// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/ui/alert"
)

// AlertIntent is the intent of a banner, see AlertIntentOk. Like alert.Intent of nago.
type AlertIntent int32

const (
	// AlertIntentError marks an error, which is the default.
	AlertIntentError AlertIntent = 0
	// AlertIntentOk marks a neutral information.
	AlertIntentOk AlertIntent = 1
	// AlertIntentWarning marks a warning.
	AlertIntentWarning AlertIntent = 2
	// AlertIntentSuccess marks a success.
	AlertIntentSuccess AlertIntent = 3
)

// AlertMessage is a message, which AlertShowBannerMessage shows as a banner. Like alert.Message of nago.
type AlertMessage struct {
	Title   string
	Message string
	Intent  AlertIntent
	// Duration is the time in milliseconds, which the message is shown.
	Duration int32
}

// AlertShowBannerMessage shows the message as a banner, which is rendered by the window with AlertBannerMessages.
// Like alert.ShowBannerMessage of nago.
//
//gostrict:host
func AlertShowBannerMessage(wnd Window, msg AlertMessage) {
	alert.ShowBannerMessage(wnd.nago(), msg.nago())
}

// TAlertBannerMessages shows the banners of the window, see AlertBannerMessages.
type TAlertBannerMessages struct{ ref js.Ref }

func (TAlertBannerMessages) isView() {}

// AlertBannerMessages returns the overlay of the banner messages of the window. It is empty, if no message is shown.
// Like alert.BannerMessages of nago.
//
//gostrict:host
func AlertBannerMessages(wnd Window) TAlertBannerMessages {
	return TAlertBannerMessages{ref: wrap(alert.BannerMessages(wnd.nago()))}
}
