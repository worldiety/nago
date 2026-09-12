// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package alert

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"go.wdy.de/nago/application/xerror"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type Intent int

const (
	IntentError Intent = iota
	IntentOk
	IntentWarning
	IntentSuccess
)

type Message struct {
	Title    string
	Message  string
	Intent   Intent
	Duration time.Duration
}

type alertNotifications struct {
	wnd      core.Window
	messages *core.TransientState[[]Message]
}

func (an alertNotifications) Render(ctx core.RenderContext) core.RenderNode {
	closeAllFunc := func() {
		an.messages.Set(make([]Message, 0))
	}

	notifications := proto.AlertNotifications{
		CloseAll: ctx.MountCallback(closeAllFunc),
	}
	notifications.Notifications = make([]proto.Component, 0)
	for i, msg := range an.messages.Get() {
		presented := core.StateOf[bool](an.wnd, ".msg-"+msg.Title+msg.Message).Init(func() bool {
			return true
		})
		if !presented.Get() {
			continue
		}

		notifications.Notifications = append(
			notifications.Notifications,
			Banner(msg.Title, msg.Message).
				ID(fmt.Sprintf("banner-%d", i)).
				Intent(msg.Intent).
				Closeable(presented).
				OnClosed(func() {
					an.messages.Set(slices.DeleteFunc(an.messages.Get(), func(message Message) bool {
						return message == msg
					}))
				}).Render(ctx),
		)
	}

	return &notifications
}

// TBannerMessages is a feedback component (Banner Messages).
// It manages and displays transient banner-style notifications within a window.
// This component is typically used for showing short-lived feedback messages
// (e.g., success, error, info) that appear temporarily and may stack if multiple
// messages are triggered.
type TBannerMessages struct {
	wnd core.Window
}

// BannerMessages may return nil, if no information needs to be displayed. Otherwise, it appends to
// the modal overlay.
func BannerMessages(wnd core.Window) TBannerMessages {
	return TBannerMessages{
		wnd: wnd,
	}
}

// Render displays all currently active banner messages in an overlay.
// The messages are shown in a scrollable vertical stack, with automatic padding
// adjustments for small screen sizes. Each message is wrapped in a Banner
// component that can auto-close after a duration or be dismissed manually.
// When a banner is closed, it is removed from the transient state so the list
// stays up-to-date.
func (t TBannerMessages) Render(ctx core.RenderContext) core.RenderNode {
	messages := core.TransientStateOf[[]Message](t.wnd, ".nago-messages")
	if len(messages.Get()) == 0 {
		return nil
	}

	return ui.NotificationsOverlay(alertNotifications{wnd: t.wnd, messages: messages}).
		Right(ui.L24).
		Top(ui.L120).
		Bottom(ui.L24).
		Render(ctx)
}

// ShowBannerMessage puts the given msg into the global messages state list.
// Just include [MessageList] always in your view tree, which will overlay the message as required.
// This is thread safe.
func ShowBannerMessage(wnd core.Window, msg Message) {
	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")
	//messages.Mutex().Lock() deadlock or logical races on message body or mutex for all or mutex on window?
	//defer messages.Mutex().Unlock()

	if slices.Contains(messages.Get(), msg) {
		return
	}

	messages.Set(append(messages.Get(), msg))
}

// TBannerError is a feedback component(Banner Error).
type TBannerError struct {
	err error
}

// BannerError wraps a given error into a TBannerError,
// which can later be rendered as a user-visible banner.
func BannerError(err error) TBannerError {
	return TBannerError{err: err}
}

// Render transforms the stored error into a banner message.
// Known errors are classified via [xerror.Present] into user-friendly,
// localized messages. For unknown errors, it shows a generic fallback
// message with a support token for reference.
func (t TBannerError) Render(ctx core.RenderContext) core.RenderNode {
	if t.err == nil {
		return nil
	}

	p := xerror.PresentOrGeneric(ctx.Window(), t.err)
	token := p.Token()

	if p.Recognized() {
		slog.Error("handled customized banner error", "err", t.err.Error(), "token", token)
		return Banner(p.Title, p.Message+" Code: "+token).Render(ctx)
	}

	slog.Error("unexpected banner error", "token", token, "err", t.err.Error())

	return Banner(p.Title, p.Message).Render(ctx)
}

// ShowBannerError is like ShowBannerMessage but specialized on internal unhandled errors and hides
// the actual error message from the user to avoid leaking secret details. Just a token is communicated,
// so that the original message can be found from the log.
// This is thread safe.
func ShowBannerError(wnd core.Window, err error) {
	if err == nil {
		return
	}

	p := xerror.PresentOrGeneric(wnd, err)
	token := p.Token()

	if p.Recognized() {
		slog.Error("handled customized show banner error", "err", err.Error(), "token", token)
		ShowBannerMessage(wnd, Message{Title: p.Title, Message: p.Message + " Code: " + token})
		return
	}

	msg := Message{Title: p.Title, Message: p.Message}

	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")

	if slices.Contains(messages.Get(), msg) {
		return
	}

	slog.Error("banner snackbar handled error", "token", token, "err", err)

	messages.Set(append(messages.Get(), msg))
}
