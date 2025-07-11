package ui_test

import (
	"fmt"
	"go.wdy.de/nago/presentation/ui"
)

// Example-Img: [/images/components/basic/text/text-with-methods-example.png].
func ExampleText() {
	ui.Text("hello world").
		Action(func() {
			fmt.Print("Nago is easy to use")
		}).
		Underline(true).
		Color("#eb4034").
		Border(ui.Border{}.Width("2px").Color("#4287f5"))
}
