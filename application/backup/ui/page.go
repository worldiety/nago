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

func BackupAndRestorePage(wnd core.Window, restore backup.Restore, bckup backup.Backup) core.View {
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
				ui.Text("Mit dieser Funktion wird ein vollständiges Backup aller Daten erstellt. Das Backup enthält unverschlüsselte und im Klartext lesbare vertrauliche und personenbezogene Daten. "+
					"Die Backup-Datei muss entsprechend vertraulich und gemäß der Richtlinien behandelt werden. Dieser Vorgang wird entsprechend im Log mit Ihren Nutzerdaten hinterlegt. "+
					"Lassen Sie diese Seite geöffnet und stellen Sie sicher, dass mit dem System nicht gearbeitet wird, um einen konsistenten Zustand über alle Store zu erhalten. "+
					"Der Download der Zip-Datei startet sofort. Warten Sie ab, bis der Download vollständig ist."),
			)).Footer(
			ui.PrimaryButton(func() {
				wnd.ExportFiles(core.ExportFilesOptions{
					Files: []core.File{backup.AsBackupFile(wnd.Context(), wnd.Subject(), bckup)},
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
					"Die Wiederherstellung kann einige Zeit dauern und wird bestätigt. Lassen Sie diese Seite geöffnet."),
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

						if err := restore(wnd.Context(), wnd.Subject(), reader); err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						alert.ShowBannerMessage(wnd, alert.Message{
							Title:   "Wiederherstellung erfolgreich",
							Message: "Die Wiederherstellung wurde abgeschlossen. Laden Sie die Anwendung oder den Webbrowser neu.",
						})
					},
				})
			}).Enabled(restoreBtnEnabled.Get()).Title("Aus Backup wiederherstellen"),
		),
	).Gap(ui.L16).FullWidth()
}
