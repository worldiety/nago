package uibackup

import (
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

type Pages struct {
	BackupAndRestore core.NavigationPath
}

func BackupAndRestorePage(wnd core.Window, uc backup.UseCases) core.View {
	backupBtnEnabled := core.AutoState[bool](wnd).Init(func() bool {
		return true
	})

	restoreBtnEnabled := core.AutoState[bool](wnd).Init(func() bool {
		return true
	})

	return ui.VStack(
		ui.H1("Backup und Wiederherstellung"),
		cardlayout.Card("Backup").
			Body(ui.VStack(
				ui.Text("Mit dieser Funktion wird ein vollständiges Backup aller Daten erstellt. Das Backup enthält ggf. unverschlüsselte und im Klartext lesbare vertrauliche und personenbezogene Daten. "+
					"Die Backup-Datei muss entsprechend vertraulich und gemäß der Richtlinien behandelt werden. Dieser Vorgang wird entsprechend im Log mit Ihren Nutzerdaten hinterlegt. "+
					"Lassen Sie diese Seite geöffnet und stellen Sie sicher, dass mit dem System nicht gearbeitet wird, um einen konsistenten Zustand über alle Store zu erhalten. "+
					"Der Download der Zip-Datei startet sofort. Warten Sie ab, bis der Download vollständig ist. "+
					"\n\n"+
					"Mit Nago verschlüsselte Stores bleiben verschlüsselt und können ohne den Masterkey nicht wieder hergestellt werden. Dieser Schlüssel ist nicht im Backup enthalten. Zu den standardmäßig verschlüsselten Stores gehören die Sessions und Secrets."),
			)).Footer(
			ui.PrimaryButton(func() {
				wnd.ExportFiles(core.ExportFilesOptions{
					Files: []core.File{backup.AsBackupFile(wnd.Context(), wnd.Subject(), uc.Backup)},
				})
				backupBtnEnabled.Set(false)
			}).Enabled(backupBtnEnabled.Get()).
				Title("Backup erstellen"),
		),

		cardlayout.Card("Wiederherstellung").
			Body(ui.VStack(
				ui.Text("Mit dieser Funktion wird ein Zustand aus einem Backup wiederhergestellt. Alle vorhandenen Daten werden dabei gelöscht und aus dem Backup neu erzeugt. "+
					"Blob- oder Bucket-Stores, die nicht Teil des Backups sind, bleiben unverändert. "+
					"Dieser Vorgang vernichtet Daten und ist nicht reversibel. "+
					"Versichern Sie sich, dass die Backup-Datei aus einer vertraulichen Quelle stammt, da ansonsten Dritte Zugriff auf das System erlangen können. "+
					"Dieser Vorgang wird entsprechend im Log mit Ihren Nutzerdaten hinterlegt. "+
					"Die Wiederherstellung kann einige Zeit dauern und wird bestätigt. Lassen Sie diese Seite geöffnet. "+
					"\n\n"+
					"Mit Nago verschlüsselte Stores bleiben verschlüsselt und können ohne den Masterkey nicht wieder hergestellt werden. Dieser Schlüssel ist nicht im Backup enthalten. Zu den standardmäßig verschlüsselten Stores gehören die Sessions und Secrets."),
			)).Footer(
			ui.PrimaryButton(func() {
				restoreBtnEnabled.Set(false)
				wnd.ImportFiles(core.ImportFilesOptions{
					Multiple:         false,
					MaxBytes:         1024 * 1024 * 1024 * 1024, // 1TiB limit
					AllowedMimeTypes: []string{"application/zip"},
					OnCompletion: func(files []core.File) {
						if len(files) != 1 {
							alert.ShowBannerError(wnd, std.NewLocalizedError("Fehlerhafter Upload", "Exakt eine Datei erwartet."))
							return
						}

						reader, err := files[0].Open()
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						defer reader.Close()

						if err := uc.Restore(wnd.Context(), wnd.Subject(), reader); err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						alert.ShowBannerMessage(wnd, alert.Message{
							Title:   "Wiederherstellung erfolgreich",
							Message: "Die Wiederherstellung wurde abgeschlossen. Laden Sie die Anwendung oder den Webbrowser neu.",
							Intent:  alert.IntentOk,
						})
					},
				})
			}).Enabled(restoreBtnEnabled.Get()).Title("Aus Backup wiederherstellen"),
		),

		// export key
		ui.IfFunc(wnd.Subject().HasPermission(backup.PermExportMasterKey), func() core.View {
			presentKey := core.AutoState[bool](wnd)
			return cardlayout.Card("Hauptschlüssel exportieren").
				Body(ui.VStack(
					ui.IfFunc(presentKey.Get(), func() core.View {
						key, err := uc.ExportMasterKey(wnd.Subject())
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return nil
						}

						return alert.Dialog("MasterKey", ui.PasswordField("MasterKey / Hauptschlüssel", key).FullWidth(), presentKey, alert.Ok())
					}),
					ui.Text("Mit dieser Funktion kann der NAGO Masterkey bzw. Hauptschlüssel exportiert werden. Dieser erlaubt die vollständige Wiederherstellung eines Backups mit NAGO verschlüsselten Stores, wie z.B. den Sessions oder Secrets."+
						"\n\n"+
						"Wird dieser Schlüssel veröffentlicht, sind alle Sessions und Secrets als kompromittiert anzusehen. Die Besitzer der Secrets müssen dann benachrichtigt werden und alle Secrets müssen gemäß Richtlinie rotiert werden."),
				)).Footer(
				ui.PrimaryButton(func() {
					presentKey.Set(true)
				}).Title("Masterkey auslesen"),
			)
		}),

		// import key
		ui.IfFunc(wnd.Subject().HasPermission(backup.PermReplaceMasterKey), func() core.View {
			presentKey := core.AutoState[bool](wnd)
			keyState := core.AutoState[string](wnd)

			return cardlayout.Card("Hauptschlüssel ersetzen").
				Body(ui.VStack(
					ui.IfFunc(presentKey.Get(), func() core.View {

						return alert.Dialog(
							"MasterKey",
							ui.Form(
								ui.PasswordField("Neuer Masterkey", keyState.Get()).ID("masterkey-restore").AutoComplete(false).InputValue(keyState).FullWidth(),
							).Autocomplete(false),
							presentKey,
							alert.Cancel(nil),
							alert.Save(func() (close bool) {
								if err := uc.ReplaceMasterKey(wnd.Subject(), keyState.Get()); err != nil {
									alert.ShowBannerError(wnd, err)
									return false
								}

								return true
							}),
						)
					}),
					ui.Text("Mit dieser Funktion kann der NAGO Masterkey bzw. Hauptschlüssel durch einen anderen ersetzt werden. "+
						"Dieser erlaubt die vollständige Wiederherstellung eines Backups mit NAGO verschlüsselten Stores, wie z.B. den Sessions oder Secrets. "+
						"Das funktioniert nur, sofern der Schlüssel nicht über die Hostingumgebung verwaltet wird. "+
						"\n\n"+
						"Damit der neue Masterkey angewendet wird, muss dieser Service neugestartet werden."),
				)).Footer(
				ui.PrimaryButton(func() {
					presentKey.Set(true)
				}).Title("Masterkey ersetzen"),
			)
		}),
	).Alignment(ui.Leading).Gap(ui.L16).FullWidth()
}
