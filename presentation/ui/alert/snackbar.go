package alert

import (
	"encoding/hex"
	"errors"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/crypto/sha3"
	"log/slog"
	"slices"
)

type Message struct {
	Title   string
	Message string
}

// MessageList may return nil, if no information needs to be displayed. Otherwise, it appends to
// the modal overlay.
func MessageList(wnd core.Window) core.View {
	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")
	if len(messages.Get()) == 0 {
		return nil
	}

	return ui.Overlay(
		ui.ScrollView(
			ui.VStack(
				ui.Each(slices.Values(messages.Get()), func(t Message) core.View {
					presented := core.StateOf[bool](wnd, ".msg-"+t.Title+t.Message).Init(func() bool {
						return true
					})

					return Banner(t.Title, t.Message).
						Closeable(presented).
						OnClosed(func() {
							messages.Set(slices.DeleteFunc(messages.Get(), func(message Message) bool {
								return message == t
							}))
						})
				})...,
			).Gap(ui.L8).Padding(ui.Padding{Right: ui.L16}),
		).
			Frame(ui.Frame{MaxHeight: "calc(100dvh - 8rem)"}),
	).Right(ui.L8).Top(ui.L120)
}

// ShowBannerMessage puts the given msg into the global messages state list.
// Just include [MessageList] always in your view tree, which will overlay the message as required.
func ShowBannerMessage(wnd core.Window, msg Message) {
	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")
	if slices.Contains(messages.Get(), msg) {
		return
	}

	messages.Set(append(messages.Get(), msg))
}

// ShowBannerError is like ShowBannerMessage but specialized on internal unhandled errors and hides
// the actual error message from the user to avoid leaking secret details. Just a token is communicated,
// so that the original message can be found from the log.
func ShowBannerError(wnd core.Window, err error) {
	if err == nil {
		return
	}

	var permissionDenied interface {
		PermissionDenied() bool
	}

	if errors.As(err, &permissionDenied) && permissionDenied.PermissionDenied() {
		ShowBannerMessage(wnd, Message{
			Title:   "Zugriff verweigert",
			Message: "Es besteht keine Berechtigung, um diese Inhalte oder Funktionen zu verwenden. Ein übergeordneter Rechteinhaber muss diese zunächst explizit erteilen.",
		})

		return
	}

	tmp := sha3.Sum224([]byte(err.Error()))
	token := hex.EncodeToString(tmp[:16])
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
