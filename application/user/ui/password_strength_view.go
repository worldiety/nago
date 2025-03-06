package uiuser

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/progress"
)

func PasswordStrengthView(indicator user.PasswordStrengthIndicator) core.View {
	progressColor := ui.ColorSemanticError
	if indicator.Acceptable {
		progressColor = ui.ColorSemanticGood
	}
	return ui.VStack(
		progress.LinearProgress().Progress(indicator.ComplexityScale).Color(progressColor),
		ui.If(indicator.Complexity == user.VeryWeak, checkText(false, "Kennwortstärke: Sehr Schwach")),
		ui.If(indicator.Complexity == user.Weak, checkText(false, "Kennwortstärke: Schwach")),
		ui.If(indicator.Complexity == user.Strong, checkText(true, "Kennwortstärke: Stark")),
		checkText(indicator.ContainsMinLength, fmt.Sprintf("Mindestens %d Zeichen", indicator.MinLengthRequired)),
		checkText(indicator.ContainsSpecial, "Enthält ein Sonderzeichen"),
		checkText(indicator.ContainsNumber, "Enthält eine Zahl"),
		checkText(indicator.ContainsUpperAndLowercase, "Enthält Groß- und Kleinbuchstaben"),
	).Gap(ui.L8).Alignment(ui.Leading)
}

func checkText(ok bool, text string) core.View {
	if ok {
		return ui.HStack(ui.ImageIcon(icons.Check).StrokeColor(ui.ColorSemanticGood).Frame(ui.Frame{}.Size(ui.L16, ui.L16)), ui.Text(text).Color(ui.ColorSemanticGood).Font(ui.Small))
	} else {
		return ui.HStack(ui.ImageIcon(icons.XMark).StrokeColor(ui.ColorSemanticError).Frame(ui.Frame{}.Size(ui.L16, ui.L16)), ui.Text(text).Color(ui.ColorSemanticError).Font(ui.Small))
	}
}
