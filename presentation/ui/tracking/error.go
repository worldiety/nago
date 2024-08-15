package tracking

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"log/slog"
	"runtime/debug"
)

type AnonymousErrorCode string

type UnhandledError struct {
	FirstErr error
	rendered string
	Code     AnonymousErrorCode
}

type UnhandledErrors struct {
	errors []UnhandledError
}

func (e *UnhandledErrors) Put(wnd core.Window, err error) AnonymousErrorCode {
	if err == nil {
		return ""
	}

	msg := fmt.Sprintf("%v", err)
	for _, occurredError := range e.errors {
		if occurredError.rendered == msg {
			return occurredError.Code
		}
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := AnonymousErrorCode(hex.EncodeToString(tmp[:]))
	slog.Error("captured unexpected failure in presentation code", slog.String("rootView", string(wnd.Path())), slog.Any("err", err), slog.String("code", string(code)))

	e.errors = append(e.errors, UnhandledError{
		FirstErr: err,
		rendered: msg,
		Code:     code,
	})

	return code
}

// requestSupportView allocates a new support view. See also RequestSupport.
func requestSupportView(wnd core.Window, code AnonymousErrorCode) core.View {
	showErrState := core.StateOf[bool](wnd, ".nago.global.errors.show")

	return ui.VStack(
		ui.HStack(ui.Image().Embed(icons.Bug).Frame(ui.Frame{}.Size(ui.L44, ui.L44))),
		ui.Text("Ein unerwarteter Fehler ist aufgetreten.").Font(ui.Title),
		ui.Text("Wir entschuldigen uns für diese Unannehmlichkeit."),
		ui.Text("Sie können uns einen Bericht schicken."),
		ui.SecondaryButton(func() {
			sendReport(wnd, code)
		}).Title("Bericht erstellen"),
		ui.SecondaryButton(func() {
			showErrState.Set(false)
		}).Title("weiter versuchen"),
		ui.PrimaryButton(wnd.Navigation().Reload).Title("Anwendung neu laden"),
	).Gap(ui.L16)

}

// RequestSupport communicates an unexpected technical problem, e.g. an error from the infrastructure,
// a programming error or an assertion error to the user. Only use this, if the user cannot do anything about it,
// and you cannot offer a domain-specific help. See also [SupportRequestDialog]. For security reason, the error is
// hidden from the user, but instead he will get a random error code, which is also printed into the log,
// so you can find it later. See also RequestSupportView without triggering a dialog.
func RequestSupport(wnd core.Window, err error) {
	if err == nil {
		return
	}

	showErrState := core.StateOf[bool](wnd, ".nago.global.errors.show")
	errState := core.StateOf[UnhandledErrors](wnd, ".nago.global.errors")
	errors := errState.Get()
	if core.Debug {
		errors.Put(wnd, fmt.Errorf("error: %w, caught near: %s", err, debug.Stack()))
	} else {
		errors.Put(wnd, err)
	}
	errState.Set(errors)
	showErrState.Set(true)
}

// SupportRequestDialog return either nil or the dialog to which allows contacting the developers. It shows the latest
// anonymous error code, to avoid security sensitive details. Use
// [RequestSupport] to insert an error into the global error list.
func SupportRequestDialog(wnd core.Window) core.View {
	showErrState := core.StateOf[bool](wnd, ".nago.global.errors.show")
	if !showErrState.Get() {
		return nil
	}

	errState := core.StateOf[UnhandledErrors](wnd, ".nago.global.errors")
	if len(errState.Get().errors) == 0 {
		panic("unreachable")
	}

	err := errState.Get().errors[len(errState.Get().errors)-1]
	return alert.Dialog("Ein unerwarteter Fehler ist aufgetreten", requestSupportView(wnd, err.Code), showErrState)
}

func sendReport(wnd core.Window, code AnonymousErrorCode) {

	msg := "# error report\n\n"
	msg += fmt.Sprintf("application-id: %s\n", wnd.Application().ID())
	msg += fmt.Sprintf("application-name: %s\n", wnd.Application().Name())
	msg += fmt.Sprintf("application-version: %s\n", wnd.Application().Version())
	msg += fmt.Sprintf("component: %s\n", wnd.Path())
	msg += fmt.Sprintf("values: %v\n", wnd.Values())
	msg += fmt.Sprintf("code: %s\n", code)

	wnd.ExportFiles(core.ExportFile("report.txt", []byte(msg)))

}
