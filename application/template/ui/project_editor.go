package uitemplate

import (
	"bytes"
	"fmt"
	"go.wdy.de/nago/application/template"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"os"
	"time"
)

func PageEditor(wnd core.Window, uc template.UseCases) core.View {
	tid := template.ID(wnd.Values()["project"])
	optPrj, err := uc.FindByID(wnd.Subject(), tid)
	if err != nil {
		return alert.BannerError(err)
	}

	if optPrj.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	prj := optPrj.Unwrap()

	selectedFile := core.AutoState[template.File](wnd)

	sm := wnd.Info().SizeClass == core.SizeClassSmall

	consoleState := core.AutoState[string](wnd).Init(func() string {
		return fmt.Sprintf("Projekt '%s' (%s) geladen.\nType %v", prj.Name, prj.ID, prj.Type)
	})

	lastErrorState := core.AutoState[error](wnd)
	msgSaved := core.AutoState[string](wnd)
	save := func(str string) {
		if err := uc.UpdateProjectBlob(wnd.Subject(), prj.ID, selectedFile.Get().Filename, bytes.NewBuffer([]byte(str))); err != nil {
			alert.ShowBannerError(wnd, err)
			msgSaved.Set(err.Error())
			lastErrorState.Set(err)
			consoleState.Set("Speichern nicht möglich:\n" + err.Error())
			return
		}

		lastErrorState.Set(nil)

		msgSaved.Set(fmt.Sprintf("gespeichert %s", time.Now().Format(xtime.GermanDateTimeSec)))
		consoleState.Set("Projekt erfolgreich gespeichert: " + time.Now().Format(xtime.GermanDateTimeSec))
	}

	canExecute := prj.Type != template.Unprocessed
	launcherPresented := core.AutoState[bool](wnd)
	runConfigurationSelected := core.AutoState[template.RunConfiguration](wnd)

	return ui.VStack(
		ui.HStack(ui.H1(prj.Name)).Alignment(ui.Leading),
		ui.HStack(
			ui.SecondaryButton(func() {
				wnd.ExportFiles(core.ExportFilesOptions{
					Files: []core.File{
						template.AsDownloadFile(wnd.Context(), wnd.Subject(), prj.ID, uc.ExportZip),
					},
				})
			}).PreIcon(flowbiteOutline.Download).AccessibilityLabel("Projekt als Zip exportieren"),

			ui.SecondaryButton(func() {
				wnd.ImportFiles(core.ImportFilesOptions{
					MaxBytes:         0,
					AllowedMimeTypes: []string{"application/zip"},
					OnCompletion: func(files []core.File) {
						for _, file := range files {
							reader, err := file.Open()
							if err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}

							if err := uc.ImportZip(wnd.Subject(), prj.ID, reader); err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}
						}

						alert.ShowBannerMessage(wnd, alert.Message{
							Title:   "Import erfolgreich",
							Message: "Das Projekt wurde importiert.",
							Intent:  alert.IntentOk,
						})
					},
				})

			}).PreIcon(flowbiteOutline.Upload).AccessibilityLabel("Projektdateien aus Zip importieren"),

			ui.Spacer(),
			viewProjectExecute(wnd, prj, uc, runConfigurationSelected, launcherPresented, consoleState),
			ui.IfFunc(canExecute, func() core.View {
				return ui.SecondaryButton(func() {
					launcherPresented.Set(!launcherPresented.Get())
				}).PreIcon(flowbiteOutline.Play)

			}),
		).Gap(ui.L8).Alignment(ui.Trailing).FullWidth(),
		ui.HLine().Padding(ui.Padding{Top: ui.L4}),
		ui.IfFunc(sm, func() core.View {
			return ui.VStack(
				viewProjectExplorer(wnd, prj, uc, selectedFile).Frame(ui.Frame{}.FullWidth()),
				viewProjectSource(wnd, prj, selectedFile, uc, save),
			).Alignment(ui.TopLeading).FullWidth()
		}),

		ui.IfFunc(!sm, func() core.View {
			return ui.HStack(
				viewProjectExplorer(wnd, prj, uc, selectedFile),
				ui.VLine().Padding(ui.Padding{Left: ui.L4}).Frame(ui.Frame{}),
				viewProjectSource(wnd, prj, selectedFile, uc, save),
			).Alignment(ui.Stretch).FullWidth()
		}),
		ui.HLine().Padding(ui.Padding{}),
		console(wnd, consoleState),
	).Alignment(ui.Stretch).Frame(ui.Frame{Width: ui.Full})
}
