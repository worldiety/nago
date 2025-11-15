// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package alert

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/crypto/sha3"
)

type Intent int

const (
	IntentError Intent = iota
	IntentOk
)

type Message struct {
	Title    string
	Message  string
	Intent   Intent
	Duration time.Duration
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

	padRight := ui.L24
	overlayRight := ui.L8
	isSmall := t.wnd.Info().SizeClass <= core.SizeClassSmall
	if isSmall {
		padRight = ui.L8
		overlayRight = ""
	}

	return ui.Overlay(
		ui.ScrollView(
			ui.VStack(
				ui.Each(slices.Values(messages.Get()), func(m Message) core.View {
					presented := core.StateOf[bool](t.wnd, ".msg-"+m.Title+m.Message).Init(func() bool {
						return true
					})

					return Banner(m.Title, m.Message).
						Intent(m.Intent).
						Closeable(presented).
						AutoCloseTimeoutOrDefault(m.Duration).
						OnClosed(func() {
							messages.Set(slices.DeleteFunc(messages.Get(), func(message Message) bool {
								return message == m
							}))
						})
				})...,
			).Gap(ui.L8).Padding(ui.Padding{Right: padRight}),
		).
			Frame(ui.Frame{MaxHeight: "calc(100dvh - 8rem)", MaxWidth: "100dvw"}),
	).
		Right(overlayRight).
		Top(ui.L120).
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

// makeMessageFromError converts different error types into user-facing Message structs.
// It checks for known error interfaces (e.g., NotLoggedIn, PermissionDenied, LocalizedError)
// and standard errors like os.ErrNotExist or password strength errors, returning a
// localized title and description suitable for display in a banner or snackbar.
// If the error is not recognized, it returns false so no message is shown.
func makeMessageFromError(wnd core.Window, err error) (Message, bool) {
	if err == nil {
		return Message{}, false
	}

	var permNotLoggedIn interface {
		NotLoggedIn() bool
	}

	if errors.As(err, &permNotLoggedIn) && permNotLoggedIn.NotLoggedIn() {

		return Message{
			Title:   "Zugriff verweigert",
			Message: "Diese Funktion steht nur eingeloggten Nutzern zur Verfügung.",
		}, true
	}

	var permissionDenied interface {
		PermissionDenied() bool
	}

	if errors.As(err, &permissionDenied) && permissionDenied.PermissionDenied() {
		name := "?"
		if str, ok := permissionDenied.(user.PermissionDeniedError); ok {
			localized := wnd.Bundle().Resolve(string(str))
			if localized == string(user.PermissionDeniedErr) {
				name = ""
			} else {
				name = "'" + localized + "'"
			}
		}

		msg := "Es besteht keine Berechtigung, um diese Inhalte oder Funktionen zu verwenden."
		if name != "" {
			msg += " Ein übergeordneter Rechteinhaber muss " + name + " zunächst explizit erteilen."
		}
		
		return Message{
			Title:   "Zugriff verweigert",
			Message: msg,
		}, true
	}

	var localError std.LocalizedError

	if errors.As(err, &localError) {
		return Message{
			Title:   localError.Title(),
			Message: localError.Description(),
		}, true
	}

	if errors.Is(err, os.ErrNotExist) {
		return Message{
			Title:   "Element nicht gefunden",
			Message: "Der Anwendungsfall konnte nicht ausgeführt werden, da ein Element erwartet aber nicht gefunden wurde.",
		}, true
	}

	if errors.Is(err, os.ErrExist) {
		return Message{
			Title:   "Element bereits vorhanden",
			Message: "Der Anwendungsfall konnte nicht ausgeführt werden, da ein Element nicht bereits vorhanden sein darf, aber gefunden wurde.",
		}, true
	}

	var passwordStrengthErr user.PasswordStrengthError
	if errors.As(err, &passwordStrengthErr) {
		var msg string
		if passwordStrengthErr.Strength.Complexity < user.Strong {
			msg = "Die Kennwortkomplexität ist zu niedrig."
		} else if !passwordStrengthErr.Strength.ContainsUpperAndLowercase {
			msg = "Das Kennwort muss mindestens einen Groẞ- und einen Kleinbuchstaben enthalten."
		} else if !passwordStrengthErr.Strength.ContainsMinLength {
			msg = fmt.Sprintf("Das Kennwort muss mindestens %d Zeichen enthalten.", passwordStrengthErr.Strength.MinLengthRequired)
		} else if !passwordStrengthErr.Strength.ContainsSpecial {
			msg = "Das Kennwort muss mindestens ein Sonderzeichen enthalten."
		} else if !passwordStrengthErr.Strength.ContainsBelowMaxLength {
			msg = "Das Kennwort ist zu lang."
		} else if !passwordStrengthErr.Strength.ContainsNumber {
			msg = "Das Kennwort muss mindestens eine Zahl enthalten."
		} else {
			msg = "Das Kennwort kann nicht verwendet werden."
		}
		return Message{
			Title:   "Kennwort zu schwach",
			Message: msg,
		}, true
	}

	return Message{}, false
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
// Known errors are mapped via makeMessageFromError into
// user-friendly messages. For unknown errors, it shows a
// generic fallback message with a support token for reference.
func (t TBannerError) Render(ctx core.RenderContext) core.RenderNode {
	if t.err == nil {
		return nil
	}

	tmp := sha3.Sum224([]byte(t.err.Error()))
	token := hex.EncodeToString(tmp[:16])

	if msg, ok := makeMessageFromError(ctx.Window(), t.err); ok {
		slog.Error("handled customized banner error", "err", t.err.Error(), "token", token)
		return Banner(msg.Title, msg.Message+" Code: "+token).Render(ctx)
	}

	msg := Message{
		Title:   "Ein unerwarteter Fehler ist aufgetreten",
		Message: fmt.Sprintf("Sie können sich mit dem folgenden Code an den Support wenden: %s", token),
	}

	slog.Error("unexpected banner error", "token", token, "err", t.err.Error())

	return Banner(msg.Title, msg.Message).Render(ctx)
}

// ShowBannerError is like ShowBannerMessage but specialized on internal unhandled errors and hides
// the actual error message from the user to avoid leaking secret details. Just a token is communicated,
// so that the original message can be found from the log.
// This is thread safe.
func ShowBannerError(wnd core.Window, err error) {
	if err == nil {
		return
	}

	tmp := sha3.Sum224([]byte(err.Error()))
	token := hex.EncodeToString(tmp[:16])

	if msg, ok := makeMessageFromError(wnd, err); ok {
		slog.Error("handled customized show banner error", "err", err.Error(), "token", token)
		msg.Message += " Code: " + token
		ShowBannerMessage(wnd, msg)
		return
	}

	msg := Message{
		Title:   "Ein unerwarteter Fehler ist aufgetreten",
		Message: fmt.Sprintf("Sie können sich mit dem folgenden Code an den Support wenden: %s", token),
	}

	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")

	if slices.Contains(messages.Get(), msg) {
		return
	}

	slog.Error("banner snackbar handled error", "token", token, "err", err)

	messages.Set(append(messages.Get(), msg))
}
