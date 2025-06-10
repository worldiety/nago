// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/progress"
)

const msgInfoPasswordStrength = "Die Kennwortstärke wird algorithmisch bewertet und beschreibt wie vorhersehbar das Kennwort ist. Ein langes Passwort mit Sequenzen von sich wiederholenden Zeichen ist beispielweise grundsätzlich unsicherer als eine zufällige Abfolge.\n\nWenn das Passwort zu vorhersehbar ist, wird es als schwaches Passwort abgelehnt."

func PasswordStrengthView(wnd core.Window, indicator user.PasswordStrengthIndicator) core.View {
	progressColor := ui.ColorSemanticError
	if indicator.Acceptable {
		progressColor = ui.ColorSemanticGood
	}
	return ui.VStack(
		progress.LinearProgress().Progress(indicator.ComplexityScale).Color(progressColor),
		ui.If(indicator.Complexity == user.VeryWeak, checkText(wnd, false, "Kennwortstärke: Sehr Schwach", msgInfoPasswordStrength)),
		ui.If(indicator.Complexity == user.Weak, checkText(wnd, false, "Kennwortstärke: Schwach", msgInfoPasswordStrength)),
		ui.If(indicator.Complexity == user.Strong, checkText(wnd, true, "Kennwortstärke: Stark", msgInfoPasswordStrength)),
		checkText(wnd, indicator.ContainsMinLength, fmt.Sprintf("Mindestens %d Zeichen", indicator.MinLengthRequired), ""),
		checkText(wnd, indicator.ContainsSpecial, "Enthält ein Sonderzeichen", ""),
		checkText(wnd, indicator.ContainsNumber, "Enthält eine Zahl", ""),
		checkText(wnd, indicator.ContainsUpperAndLowercase, "Enthält Groß- und Kleinbuchstaben", ""),
	).Gap(ui.L8).Alignment(ui.Leading).FullWidth()
}

func checkText(wnd core.Window, ok bool, text string, info string) core.View {
	if ok {
		return ui.HStack(ui.ImageIcon(icons.Check).StrokeColor(ui.ColorSemanticGood).Frame(ui.Frame{}.Size(ui.L16, ui.L16)), ui.Text(text).Color(ui.ColorSemanticGood).Font(ui.Small), ui.IfFunc(info != "", func() core.View {
			return infoIcon(wnd, info)
		}))
	} else {
		return ui.HStack(ui.ImageIcon(icons.XMark).StrokeColor(ui.ColorSemanticError).Frame(ui.Frame{}.Size(ui.L16, ui.L16)), ui.Text(text).Color(ui.ColorSemanticError).Font(ui.Small), ui.IfFunc(info != "", func() core.View {
			return infoIcon(wnd, info)
		}))
	}
}

func infoIcon(wnd core.Window, text string) core.View {
	presented := core.StateOf[bool](wnd, "ifo-presented"+text)
	return ui.HStack(
		ui.TertiaryButton(func() {
			presented.Set(true)
		}).PreIcon(flowbiteOutline.InfoCircle).AccessibilityLabel("Erklärung anzeigen").Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
		alert.Dialog("Info", ui.Text(text), presented, alert.Ok()),
	)

}
