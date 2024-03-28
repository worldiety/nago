package xdialog

import (
	"crypto/rand"
	"encoding/hex"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
)

func ShowMessage(ctx ui.ModalOwner, msg string) {
	ctx.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		dlg.Body().Set(ui.MakeText(msg))
		dlg.Actions().Append(ui.NewButton(func(btn *ui.Button) {
			btn.Caption().Set("Ok")
			btn.Action().Set(func() {
				ctx.Modals().Remove(dlg)
			})
		}))
	}))
}

func ErrorView(msg string, err error) ui.LiveComponent {
	if err == nil {
		return nil
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured failure on frontend", slog.Any("err", err), slog.String("code", code), slog.String("msg", msg))
	return ui.MakeText("oh snap: " + code)
}

func HandleError(ctx ui.ModalOwner, msg string, err error) bool {
	if err == nil {
		return false
	}

	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	code := hex.EncodeToString(tmp[:])
	slog.Error("captured failure on frontend", slog.Any("err", err), slog.String("code", code), slog.String("msg", msg))
	ctx.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		dlg.Body().Set(ui.MakeText(msg + "(" + code + ")"))
		dlg.Actions().Append(ui.NewButton(func(btn *ui.Button) {
			btn.Caption().Set("Ok")
			btn.Action().Set(func() {
				ctx.Modals().Remove(dlg)
			})
		}))

	}))

	return true
}

func ShowDelete(ctx ui.ModalOwner, msg string, onDelete, onCancel func()) {
	ctx.Modals().Append(ui.NewDialog(func(dlg *ui.Dialog) {
		dlg.Body().Set(ui.MakeText(msg))
		dlg.Actions().Append(ui.NewButton(func(btn *ui.Button) {
			btn.Caption().Set("Löschen")
			btn.Style().Set("destructive")
			btn.Action().Set(func() {
				onDelete()
				ctx.Modals().Remove(dlg)
			})
		}))
		dlg.Actions().Append(ui.NewButton(func(btn *ui.Button) {
			btn.Caption().Set("Abbrechen")
			btn.Action().Set(func() {
				if onCancel != nil {
					onCancel()
				}
				ctx.Modals().Remove(dlg)
			})
		}))
	}))
}
