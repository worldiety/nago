package uisettings

import (
	"fmt"
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"reflect"
)

func PageSettings(wnd core.Window, load settings.LoadGlobal, store settings.StoreGlobal) core.View {
	typeName := wnd.Values()["type"]
	decl, ok := enum.DeclarationFor[settings.GlobalSettings]()
	if !ok {
		return alert.Banner("Einstellung unbekannt", "Es sind keine globalen Einstellungen registriert.")
	}

	var rType reflect.Type
	for r := range decl.Variants() {
		if r.Name() == typeName {
			rType = r
			break
		}
	}

	if rType == nil {
		return alert.Banner("Einstellung unbekannt", fmt.Sprintf("Der Typ '%s' ist als Einstellungen nicht registriert.", typeName))
	}

	meta := settings.ReadMetaData(rType)

	state := core.AutoState[settings.GlobalSettings](wnd).Init(func() settings.GlobalSettings {
		s, err := load(wnd.Subject(), rType)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return s // intentionally continue
		}

		return s
	})

	stateCanSave := core.AutoState[bool](wnd)

	state.Observe(func(newValue settings.GlobalSettings) {
		stateCanSave.Set(true)
	})

	return ui.VStack(
		ui.H1(meta.Title),
		ui.Text(meta.Description),
		ui.HLine(),
		form.Auto[settings.GlobalSettings](form.AutoOptions{}, state),
		ui.HLine(),
		ui.HStack(
			ui.SecondaryButton(func() {
				wnd.Navigation().Back()
			}).Title("Zurück"),
			ui.PrimaryButton(func() {
				if err := store(wnd.Subject(), state.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				alert.ShowBannerMessage(wnd, alert.Message{
					Title:   "Einstellungen gespeichert",
					Message: "Die Einstellungen wurden aktualisiert.",
					Intent:  alert.IntentOk,
				})

				stateCanSave.Set(false)
			}).Title("Speichern").Enabled(stateCanSave.Get()),
		).Gap(ui.L8).FullWidth().Alignment(ui.Trailing),
	).Alignment(ui.Leading).Frame(ui.Frame{MaxWidth: ui.L560}.FullWidth())
}
