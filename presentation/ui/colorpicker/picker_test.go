package colorpicker_test

import (
	"fmt"
	abc "go.wdy.de/nago/presentation/ui/colorpicker"
)

func ExamplePalettePicker() {
	abc.PalettePicker("Colorpicker", abc.DefaultPalette)
}

func ExamplePalettePicker_colors() {
	fmt.Println("klappt")
	abc.PalettePicker("2. Example", abc.DefaultPalette)
}
