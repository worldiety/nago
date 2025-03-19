package uitemplate

import (
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func console(wnd core.Window, consoleState *core.State[string]) core.View {
	popOut := core.AutoState[bool](wnd)

	stack := ui.HStack(
		ui.VStack(
			ui.ScrollView(ui.Text(consoleState.Get())).Frame(ui.Frame{Height: ui.Full}.FullWidth()),
		).Frame(ui.Frame{Width: ui.Full, Height: ui.Full}),
		ui.VLine().Padding(ui.Padding{Left: ui.L4}).Frame(ui.Frame{}),
		ui.VStack(
			ui.TertiaryButton(func() {
				popOut.Set(!popOut.Get())
			}).PreIcon(flowbiteOutline.RestoreWindow).AccessibilityLabel("Konsole verkleinern/vergrößern"),

			ui.TertiaryButton(func() {
				if err := wnd.Clipboard().SetText(consoleState.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
			}).PreIcon(flowbiteOutline.Clipboard).AccessibilityLabel("Ausgabe in Zwischenablage kopieren"),
		).Alignment(ui.Top).
			Frame(ui.Frame{Width: ui.L48, Height: ui.Full}),
	).Alignment(ui.Stretch)

	if popOut.Get() {
		return ui.VStack(
			//stack.Frame(ui.Frame{Height: "calc(100vh - 16rem)"}), // 14rem = pos.Top+pos.Bottom+Padding
			stack.Frame(ui.Frame{Height: "100%"}),
		).
			Alignment(ui.Stretch).
			Position(ui.Position{
				Type:   ui.PositionFixed,
				Left:   ui.L64,
				Top:    ui.L160,
				Right:  ui.L560,
				Bottom: ui.L64,
			}).
			BackgroundColor(ui.ColorCardBody).
			Padding(ui.Padding{}.All(ui.L16)).
			Border(ui.Border{}.Radius(ui.L20).Elevate(4))
	} else {
		return stack.Frame(ui.Frame{Width: ui.Full, Height: ui.L160})
	}
}
