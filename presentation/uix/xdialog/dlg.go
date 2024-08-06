package xdialog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
	"log/slog"
)

// Footer returns a default formatted footer line, using the correct alignment rules.
// Use this for your custom dialogs.
func Footer(buttons ...*uilegacy.Button) core.View {
	return uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
		hstack.SetAlignment(ora.Leading)
		for _, b := range buttons {
			hstack.Append(b)
		}
	})
}

func ShowMessage(ctx uilegacy.ModalOwner, msg string) {
	Confirm(ctx, msg, nil)
}

func Confirm(owner uilegacy.ModalOwner, msg string, confirmed func()) {
	owner.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
		dlg.Body().Set(uilegacy.MakeText(msg))
		dlg.Size().Set(ora.ElementSizeSmall)
		dlg.Footer().Set(Footer(uilegacy.NewActionButton("Ok", func() {
			if confirmed != nil {
				confirmed()
			}

			owner.Modals().Remove(dlg)
		})))
	}))
}

// deprecated: use RequestSupportView
func ErrorView(msg string, err error) core.View {
	if err == nil {
		return nil
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured failure on frontend", slog.Any("err", err), slog.String("code", code), slog.String("msg", msg))
	return uilegacy.MakeText("oh snap: " + code)
}

// RequestSupportView allocates a new support view. See also RequestSupport.
func RequestSupportView(wnd core.Window, err error) core.View {
	if err == nil {
		err = fmt.Errorf("no error, but support requested")
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured unexpected failure in presentation code", slog.Any("err", err), slog.String("code", code))

	return uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
		vstack.Append(
			uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
				hstack.Append(uilegacy.NewImage(func(img *uilegacy.Image) {
					img.SetDataURI([]byte(icon.Bug))
				}))
			}),

			uilegacy.NewText(func(text *uilegacy.Text) {
				text.Size().Set("2xl")
				text.Value().Set("Ein unerwarteter Fehler ist aufgetreten.")
			}),
			uilegacy.MakeText("Wir entschuldigen uns für diese Unannehmlichkeit."),
			uilegacy.MakeText("Sie können uns einen Bericht schicken."),
			uilegacy.NewButton(func(btn *uilegacy.Button) {
				btn.Style().Set(ora.Secondary)
				btn.Caption().Set("Bericht erstellen")
				btn.Action().Set(func() {
					sendReport(wnd, code)
				})
			}),

			uilegacy.NewButton(func(btn *uilegacy.Button) {
				btn.Caption().Set("Anwendung neu laden")
				btn.Action().Set(wnd.Navigation().Reload)
			}),
		)
	})
}

// RequestSupport communicates an unexpected technical problem, e.g. an error from the infrastructure,
// a programming error or an assertion error to the user. Only use this, if the user cannot do anything about it,
// and you cannot offer a domain-specific help. See also RequestSupportView.
func RequestSupport(wnd core.Window, owner uilegacy.ModalOwner, err error) {
	owner.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
		dlg.Body().Set(RequestSupportView(wnd, err))
	}))
}

func sendReport(wnd core.Window, code string) {

	msg := "# error report\n\n"
	msg += fmt.Sprintf("application-id: %s\n", wnd.Application().ID())
	msg += fmt.Sprintf("application-name: %s\n", wnd.Application().Name())
	msg += fmt.Sprintf("application-version: %s\n", wnd.Application().Version())
	msg += fmt.Sprintf("component: %s\n", wnd.ViewRoot().Factory())
	msg += fmt.Sprintf("values: %v\n", wnd.Values())
	msg += fmt.Sprintf("code: %s\n", code)

	sendErr := wnd.SendFiles(core.FilesIter(mem.From(mem.Entries{
		"report.txt": []byte(msg),
	})))

	if sendErr != nil {
		slog.Error("failed to send report", slog.Any("err", sendErr))
	}
}

// deprecated: this has an unexpected API. Use if-err clause with RequestSupport as replacement
func HandleError(ctx uilegacy.ModalOwner, msg string, err error) bool {
	if err == nil {
		return false
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured failure on frontend", slog.Any("err", err), slog.String("code", code), slog.String("msg", msg))
	ctx.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
		dlg.Body().Set(uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
			vstack.Append(
				uilegacy.MakeText(msg),
				uilegacy.MakeText("Die Fehlerkennung für den Support lautet: "+code),
			)
		}))
		dlg.Title().Set("Ein Fehler ist aufgetreten.")

		dlg.Footer().Set(Footer(uilegacy.NewActionButton("Ok", func() {
			ctx.Modals().Remove(dlg)
		})))

	}))

	return true
}

func ShowDelete(ctx uilegacy.ModalOwner, msg string, onDelete, onCancel func()) {
	ctx.Modals().Append(uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
		dlg.Body().Set(uilegacy.MakeText(msg))

		dlg.Footer().Set(Footer(
			uilegacy.NewActionButton("Abbrechen", func() {
				if onCancel != nil {
					onCancel()
				}
				ctx.Modals().Remove(dlg)
			}),
			// UX: delete is the right most action, because it brings you "forward"
			uilegacy.NewActionButton("Löschen", func() {
				onDelete()
				ctx.Modals().Remove(dlg)
			}),
		))
	}))
}
