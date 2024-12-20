package uiuser

import (
	"fmt"
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/picker"
	"slices"
)

func licensePicker(wnd core.Window, statistics billing.UserLicenseStatistics, state *core.State[user.User]) ui.DecoredView {
	presented := core.DerivedState[bool](state, "licenses-dlg")
	var btnText string
	switch len(state.Get().Licenses) {
	case 0:
		btnText = "Keine Lizenzen zugewiesen"
	case 1:
		for _, stat := range statistics.Stats {
			if state.Get().Licenses[0] == stat.License.ID {
				btnText = stat.License.Name
				break
			}
		}

		if btnText == "" {
			btnText = "orphaned license"
		}

	default:
		btnText = fmt.Sprintf("%d Lizenzen", len(state.Get().Licenses))
	}

	return picker.Button(func() {
		presented.Set(true)
	}).Dialog(licensePickerDialog(wnd, presented, statistics, state)).
		Content(ui.Text(btnText)).
		Frame(ui.Frame{}.FullWidth())
}

func licensePickerDialog(wnd core.Window, presented *core.State[bool], statistics billing.UserLicenseStatistics, state *core.State[user.User]) core.View {
	content := ui.VStack(
		ui.Each(slices.Values(statistics.Stats), func(t billing.PerUserLicenseStats) core.View {
			checked := core.StateOf[bool](wnd, state.ID()+"-check-"+string(t.License.ID)).Init(func() bool {
				return slices.Contains(state.Get().Licenses, t.License.ID)
			})

			checked.Observe(func(newValue bool) {
				usr := state.Get()

				// always clean up, to avoid whatever multiple corruptions occured
				usr.Licenses = slices.DeleteFunc(usr.Licenses, func(id license.ID) bool {
					return id == t.License.ID
				})

				if newValue && t.Avail() > 0 {
					usr.Licenses = append(usr.Licenses, t.License.ID)
				}

				if newValue && t.Avail() <= 0 {
					checked.Set(false)
				}

				state.Set(usr)
				state.Notify()
			})
			return ui.CheckboxField(fmt.Sprintf("%s (noch %d von %d verfügbar)", t.License.Name, t.Avail(), t.License.MaxUsers), checked.Get()).
				InputValue(checked)
		})...,
	).Alignment(ui.Leading)

	return alert.Dialog("Lizenzen zuweisen", content, presented, alert.Ok())
}
