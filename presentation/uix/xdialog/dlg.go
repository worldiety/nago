package xdialog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
)

// Footer returns a default formatted footer line, using the correct alignment rules.
// Use this for your custom dialogs.
func Footer(buttons ...*ui.Button) core.Component {
	return ui.NewHStack(func(hstack *ui.FlexContainer) {
		ui.HStackAlignRight(hstack)
		for _, b := range buttons {
			hstack.Append(b)
		}
	})
}

func ShowMessage(ctx ui.ModalOwner, msg string) {
	Confirm(ctx, msg, nil)
}

func Confirm(owner ui.ModalOwner, msg string, confirmed func()) {
	owner.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		dlg.Body().Set(ui.MakeText(msg))
		dlg.Size().Set(ora.ElementSizeSmall)
		dlg.Footer().Set(Footer(ui.NewActionButton("Ok", func() {
			if confirmed != nil {
				confirmed()
			}

			owner.Modals().Remove(dlg)
		})))
	}))
}

// deprecated: use RequestSupportView
func ErrorView(msg string, err error) core.Component {
	if err == nil {
		return nil
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured failure on frontend", slog.Any("err", err), slog.String("code", code), slog.String("msg", msg))
	return ui.MakeText("oh snap: " + code)
}

// RequestSupportView allocates a new support view. See also RequestSupport.
func RequestSupportView(wnd core.Window, err error) core.Component {
	if err == nil {
		err = fmt.Errorf("no error, but support requested")
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured unexpected failure in presentation code", slog.Any("err", err), slog.String("code", code))

	return ui.NewVStack(func(vstack *ui.FlexContainer) {
		ui.VStackAlignCenter(vstack)
		vstack.ContentAlignment().Set(ora.ContentCenter)
		vstack.Append(
			ui.NewHStack(func(hstack *ui.FlexContainer) {
				ui.HStackAlignCenter(hstack)
				hstack.Append(ui.NewImage(func(img *ui.Image) {
					img.SetDataURI([]byte(icon.Bug))
				}))
			}),

			ui.NewText(func(text *ui.Text) {
				text.Size().Set("2xl")
				text.Value().Set("Ein unerwarteter Fehler ist aufgetreten.")
			}),
			ui.MakeText("Wir entschuldigen uns für diese Unannehmlichkeit."),
			ui.MakeText("Sie können uns einen Bericht schicken."),
			ui.NewButton(func(btn *ui.Button) {
				btn.Style().Set(ora.Secondary)
				btn.Caption().Set("Bericht erstellen")
				btn.Action().Set(func() {
					sendReport(wnd, code)
				})
			}),

			ui.NewButton(func(btn *ui.Button) {
				btn.Caption().Set("Anwendung neu laden")
				btn.Action().Set(wnd.Navigation().Reload)
			}),
		)
	})
}

// RequestSupport communicates an unexpected technical problem, e.g. an error from the infrastructure,
// a programming error or an assertion error to the user. Only use this, if the user cannot do anything about it,
// and you cannot offer a domain-specific help. See also RequestSupportView.
func RequestSupport(wnd core.Window, owner ui.ModalOwner, err error) {
	owner.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
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
func HandleError(ctx ui.ModalOwner, msg string, err error) bool {
	if err == nil {
		return false
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured failure on frontend", slog.Any("err", err), slog.String("code", code), slog.String("msg", msg))
	ctx.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		dlg.Body().Set(ui.NewVStack(func(vstack *ui.FlexContainer) {
			vstack.Append(
				ui.MakeText(msg),
				ui.MakeText("Die Fehlerkennung für den Support lautet: "+code),
			)
		}))
		dlg.Title().Set("Ein Fehler ist aufgetreten.")

		dlg.Footer().Set(Footer(ui.NewActionButton("Ok", func() {
			ctx.Modals().Remove(dlg)
		})))

	}))

	return true
}

func ShowDelete(ctx ui.ModalOwner, msg string, onDelete, onCancel func()) {
	ctx.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		dlg.Body().Set(ui.MakeText(msg))

		dlg.Footer().Set(Footer(
			ui.NewActionButton("Abbrechen", func() {
				if onCancel != nil {
					onCancel()
				}
				ctx.Modals().Remove(dlg)
			}),
			// UX: delete is the right most action, because it brings you "forward"
			ui.NewActionButton("Löschen", func() {
				onDelete()
				ctx.Modals().Remove(dlg)
			}),
		))
	}))
}
