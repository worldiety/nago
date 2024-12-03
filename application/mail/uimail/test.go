package uimail

import (
	mail3 "go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	mail2 "net/mail"
)

func SendTestMailPage(wnd core.Window, send mail3.SendMail) core.View {
	mailSubject := core.AutoState[string](wnd).Init(func() string {
		return "NAGO Mail-Server Testnachricht"
	})
	mailBody := core.AutoState[string](wnd).Init(func() string {
		return "Dies ist eine Testnachricht von einem NAGO Server."
	})
	mailRec := core.AutoState[string](wnd).Init(func() string {
		return "admin@worldiety.de"
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
		),
	).FullWidth()
}
