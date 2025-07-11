package ui_test

import (
	"fmt"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

// Example-Img: [/images/components/basic/buttons/primary-button-with-pre-icon.png]
func ExamplePrimaryButton_withPreIcon() {
	ui.PrimaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World").PreIcon(icons.SpeakerWave)
	//Output:
}
