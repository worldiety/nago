package alert

import (
	"encoding/hex"
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/std"
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

// BannerMessages may return nil, if no information needs to be displayed. Otherwise, it appends to
// the modal overlay.
func BannerMessages(wnd core.Window) core.View {
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

var permDeniedMsg = Message{
	Title:   "Zugriff verweigert",
	Message: "Es besteht keine Berechtigung, um diese Inhalte oder Funktionen zu verwenden. Ein übergeordneter Rechteinhaber muss diese zunächst explizit erteilen.",
}

var permReqLogin = Message{
	Title:   "Zugriff verweigert",
	Message: "Diese Funktion steht nur eingeloggten Nutzern zur Verfügung.",
}

func BannerError(err error) core.View {
	if err == nil {
		return nil
	}

	var permNotLoggedIn interface {
		NotLoggedIn() bool
	}

	if errors.As(err, &permNotLoggedIn) && permNotLoggedIn.NotLoggedIn() {
		return Banner(permReqLogin.Title, permReqLogin.Message)
	}

	var permissionDenied interface {
		PermissionDenied() bool
	}

	if errors.As(err, &permissionDenied) && permissionDenied.PermissionDenied() {
		return Banner(permDeniedMsg.Title, permDeniedMsg.Message)
	}

	var localError std.LocalizedError

	if errors.As(err, &localError) {
		return Banner(localError.Title(), localError.Description())
	}

	tmp := sha3.Sum224([]byte(err.Error()))
	token := hex.EncodeToString(tmp[:16])
	msg := Message{
		Title:   "Ein unerwarteter Fehler ist aufgetreten",
		Message: fmt.Sprintf("Sie können sich mit dem folgenden Code an den Support wenden: %s", token),
	}

	slog.Error("unexpected banner error", "token", token, "err", err)

	return Banner(msg.Title, msg.Message)
}

// ShowBannerError is like ShowBannerMessage but specialized on internal unhandled errors and hides
// the actual error message from the user to avoid leaking secret details. Just a token is communicated,
// so that the original message can be found from the log.
// This is thread safe.
func ShowBannerError(wnd core.Window, err error) {
	if err == nil {
		return
	}

	var permNotLoggedIn interface {
		NotLoggedIn() bool
	}

	if errors.As(err, &permNotLoggedIn) && permNotLoggedIn.NotLoggedIn() {
		ShowBannerMessage(wnd, permReqLogin)

		return
	}

	var permissionDenied interface {
		PermissionDenied() bool
	}

	if errors.As(err, &permissionDenied) && permissionDenied.PermissionDenied() {
		ShowBannerMessage(wnd, permDeniedMsg)

		return
	}

	var localError std.LocalizedError

	if errors.As(err, &localError) {
		ShowBannerMessage(wnd, Message{
			Title:   localError.Title(),
			Message: localError.Description(),
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
