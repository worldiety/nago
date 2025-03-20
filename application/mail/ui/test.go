package uimail

import (
	"bytes"
	"encoding/json"
	mail3 "go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"golang.org/x/text/language"
	"io"
	mail2 "net/mail"
)

func SendTestMailPage(wnd core.Window, send mail3.SendMail, exec template.Execute) core.View {
	mailSubject := core.AutoState[string](wnd).Init(func() string {
		return "NAGO Mail-Server Testnachricht"
	})
	mailBody := core.AutoState[string](wnd).Init(func() string {
		return "Dies ist eine Testnachricht von einem NAGO Server."
	})
	mailRec := core.AutoState[string](wnd).Init(func() string {
		return "admin@worldiety.de"
	})

	tplMailRec := core.AutoState[string](wnd).Init(func() string {
		return "admin@worldiety.de"
	})

	tplMailSubject := core.AutoState[string](wnd).Init(func() string {
		return "NAGO Mail-Server Template Testnachricht"
	})

	tplMailId := core.AutoState[string](wnd).Init(func() string {
		return "nago.template.system.mails"
	})

	tplName := core.AutoState[string](wnd).Init(func() string {
		return "ResetPassword"
	})

	tplLang := core.AutoState[string](wnd).Init(func() string {
		return "de_DE"
	})

	tplMailModel := core.AutoState[string](wnd).Init(func() string {
		return "{\n}"
	})

	return ui.VStack(
		ui.H1("Mailkonfiguration testen"),

		cardlayout.Card("Test Mail verschicken").
			Body(
				ui.VStack(
					ui.TextField("Empfänger", mailRec.Get()).InputValue(mailRec).FullWidth(),
					ui.TextField("Betreff", mailSubject.Get()).InputValue(mailSubject).FullWidth(),
					ui.TextField("Nachricht", mailBody.Get()).InputValue(mailBody).Lines(5).FullWidth(),
				).FullWidth().Gap(ui.L16),
			).Footer(
			ui.PrimaryButton(func() {
				_, err := send(wnd.Subject(), mail3.Mail{
					To: []mail2.Address{{
						Address: mailRec.Get(),
					}},
					Subject: mailSubject.Get(),
					Parts:   []mail3.Part{mail3.NewTextPart(mailBody.Get())},
				})

				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				alert.ShowBannerMessage(wnd, alert.Message{Title: "Nachricht versendet", Message: "Prüfen Sie das Postfach. Dies kann sofort passieren aber auch Minuten bis Stunden z.B. wegen `Grey listing` dauern."})
			}).Title("Senden"),
		).Frame(ui.Frame{MaxWidth: ui.L560, Width: ui.Full}),

		cardlayout.Card("Template Mail verschicken").
			Body(
				ui.VStack(
					ui.TextField("Empfänger", tplMailRec.Get()).InputValue(tplMailRec).FullWidth(),
					ui.TextField("Betreff", tplMailSubject.Get()).InputValue(tplMailSubject).FullWidth(),
					ui.TextField("Template-ID", tplMailId.Get()).InputValue(tplMailId).FullWidth(),
					ui.TextField("Language", tplLang.Get()).InputValue(tplLang).FullWidth(),
					ui.CodeEditor(tplMailModel.Get()).InputValue(tplMailModel).Frame(ui.Frame{Width: ui.Full, Height: ui.L320}),
				).Gap(ui.L16).FullWidth(),
			).Footer(
			ui.PrimaryButton(func() {
				langTag, err := language.Parse(tplLang.Get())
				if err != nil {
					alert.ShowBannerError(wnd, std.NewLocalizedError("Language-Fehler", err.Error()))
					return
				}

				var obj any
				if err := json.Unmarshal([]byte(tplMailModel.Get()), &obj); err != nil {
					alert.ShowBannerError(wnd, std.NewLocalizedError("Modell-Fehler", err.Error()))
					return
				}

				reader, err := exec(wnd.Subject(), template.ID(tplMailId.Get()), template.ExecOptions{
					Context:      wnd.Context(),
					Language:     langTag,
					TemplateName: tplName.Get(),
					Model:        obj,
				})

				if err != nil {
					alert.ShowBannerError(wnd, std.NewLocalizedError("Template-Fehler", err.Error()))
					return
				}

				buf, err := io.ReadAll(reader)
				if err != nil {
					alert.ShowBannerError(wnd, std.NewLocalizedError("Template-Fehler", err.Error()))
					return
				}

				var parts []mail3.Part
				if bytes.Contains(buf, []byte("<html")) {
					parts = append(parts, mail3.NewHtmlPart(string(buf)))
				} else {
					parts = append(parts, mail3.NewTextPart(string(buf)))
				}

				_, err = send(wnd.Subject(), mail3.Mail{
					To: []mail2.Address{{
						Address: tplMailRec.Get(),
					}},
					Subject: tplMailSubject.Get(),
					Parts:   parts,
				})

				if err != nil {
					alert.ShowBannerError(wnd, std.NewLocalizedError("Send-Fehler", err.Error()))
					return
				}

				alert.ShowBannerMessage(wnd, alert.Message{
					Title:  "Nachricht versendet",
					Intent: alert.IntentOk,
				})
			}).Title("Senden"),
		).Frame(ui.Frame{MaxWidth: ui.L560, Width: ui.Full}),
	).Gap(ui.L32).FullWidth()
}
