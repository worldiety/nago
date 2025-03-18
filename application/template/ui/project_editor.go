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
		if err := uc.UpdateProjectBlob(wnd.Subject(), prj.ID, selectedFile.Get().Blob, bytes.NewBuffer([]byte(str))); err != nil {
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

	var savedIcon core.SVG
	if msgSaved.Get() == "" {
		savedIcon = flowbiteOutline.Check
	} else {
		savedIcon = flowbiteOutline.CheckCircle
	}

	canExecute := prj.Type != template.Unprocessed
	launcherPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.HStack(ui.H1(prj.Name)).Alignment(ui.Leading),
		ui.HStack(
			viewProjectExecute(wnd, prj, uc, launcherPresented, consoleState),
			ui.IfFunc(canExecute, func() core.View {
				return ui.TertiaryButton(func() {
					launcherPresented.Set(true)
				}).PreIcon(flowbiteOutline.Play).AccessibilityLabel("als Vorschau ausführen")
			}),
			ui.ImageIcon(savedIcon).AccessibilityLabel(msgSaved.Get()),
		).Alignment(ui.Trailing).FullWidth(),
		ui.HLine().Padding(ui.Padding{}),
		ui.IfFunc(sm, func() core.View {
			return ui.VStack(
				viewProjectExplorer(wnd, prj, selectedFile).Frame(ui.Frame{}.FullWidth()),
				viewProjectSource(wnd, prj, selectedFile, uc, save),
			).Alignment(ui.TopLeading).FullWidth()
		}),

		ui.IfFunc(!sm, func() core.View {
			return ui.HStack(
				viewProjectExplorer(wnd, prj, selectedFile),
				ui.VLine().Padding(ui.Padding{Left: ui.L4}).Frame(ui.Frame{}),
				viewProjectSource(wnd, prj, selectedFile, uc, save),
			).Alignment(ui.Stretch).FullWidth()
		}),
		ui.HLine().Padding(ui.Padding{}),
		console(wnd, consoleState),
	).Alignment(ui.Stretch).Frame(ui.Frame{Width: ui.Full})
}

func console(wnd core.Window, consoleState *core.State[string]) core.View {
	return ui.HStack(
		ui.VStack(
			ui.ScrollView(ui.Text(consoleState.Get())).Frame(ui.Frame{Height: ui.Full}.FullWidth()),
		).Frame(ui.Frame{Width: ui.Full, Height: ui.Full}),
		ui.VLine().Padding(ui.Padding{Left: ui.L4}).Frame(ui.Frame{}),
		ui.VStack(
			ui.TertiaryButton(func() {
				if err := wnd.Clipboard().SetText(consoleState.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
			}).PreIcon(flowbiteOutline.Clipboard).AccessibilityLabel("Ausgabe in Zwischenablage kopieren"),
		).Frame(ui.Frame{Width: ui.L48, Height: ui.Full}),
	).Alignment(ui.Stretch).Frame(ui.Frame{Width: ui.Full, Height: ui.L160})
}
