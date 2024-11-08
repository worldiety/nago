package iamui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

type DashboardModel struct {
	Accounts    core.NavigationPath
	Permissions core.NavigationPath
	Groups      core.NavigationPath
	Roles       core.NavigationPath
}

func Dashboard(wnd core.Window, model DashboardModel) core.View {
	return cardlayout.CardLayout(
		ui.WindowTitle("Übersicht Benutzerverwaltung"),
		ui.If(model.Accounts != "",
			cardlayout.Card("Konten").
				Body(ui.Text("Über die Kontenverwaltung können die einzelnen bekannten Identitäten der Nutzer verwaltet werden. Hierüber können Rollen, Gruppen und Einzelberechtigungen einem Individuum zugeordnet werden.")).
				Footer(ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo(model.Accounts, nil)
				}).Title("Auswählen")),
		),
		ui.If(model.Roles != "",
			cardlayout.Card("Rollen").
				Body(ui.Text("Über die Rollenverwaltung können einzelne Berechtigungen in einer Rolle zusammengefasst werden. Dies ist die empfohlene Art einem Konto eine Menge an Berechtigungen zuzuteilen. Die konkreten Rollen ergeben sich aus der Domäne.")).
				Footer(ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(model.Roles, nil)
				}).Title("Auswählen")),
		),
		ui.If(model.Groups != "",
			cardlayout.Card("Gruppen").
				Body(ui.Text("Mittels der Gruppenverwaltung können Nutzer in Gruppen organisiert werden. Dieses Szenario wird typischerweise genutzt, um Nutzergruppen dynamisch und unabhängig von ihren Rollen zu organisieren. Damit dies Sinn macht, muss die Domäne auch Gruppen unterstützen.")).
				Footer(ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(model.Groups, nil)
				}).Title("Auswählen")),
		),
		ui.If(model.Roles != "",
			cardlayout.Card("Berechtigungen").
				Body(ui.Text("Jeder in der Domäne modellierte Anwendungsfall hat eine individuelle Berechtigung, sodass im Zweifel jedes Konto mit feingranularen Berechtigungen ausgestattet werden kann. Die Anwendungsfälle werden zur Entwicklungszeit festgelegt und können daher nicht editiert werden.")).
				Footer(ui.SecondaryButton(func() {
					wnd.Navigation().ForwardTo(model.Permissions, nil)
				}).Title("Auswählen")),
		),
	)
}
