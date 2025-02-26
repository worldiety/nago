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
	"os"
	"slices"
)

type Intent int

const (
	IntentError Intent = iota
	IntentOk
)

type Message struct {
	Title   string
	Message string
	Intent  Intent
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
						Intent(t.Intent).
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

func makeMessageFromError(err error) (Message, bool) {
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

		return Message{
			Title:   "Zugriff verweigert",
			Message: "Es besteht keine Berechtigung, um diese Inhalte oder Funktionen zu verwenden. Ein übergeordneter Rechteinhaber muss diese zunächst explizit erteilen.",
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
			Message: "Die Anwendungsfall konnte nicht ausgeführt werden, da ein Element erwartet aber nicht gefunden wurde.",
		}, true
	}

	if errors.Is(err, os.ErrExist) {
		return Message{
			Title:   "Element bereits vorhanden",
			Message: "Die Anwendungsfall konnte nicht ausgeführt werden, da ein Element nicht bereits vorhanden sein darf, aber gefunden wurde.",
		}, true
	}

	return Message{}, false
}

func BannerError(err error) core.View {
	if err == nil {
		return nil
	}

	if msg, ok := makeMessageFromError(err); ok {
		return Banner(msg.Title, msg.Message)
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

	if msg, ok := makeMessageFromError(err); ok {
		ShowBannerMessage(wnd, msg)
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
