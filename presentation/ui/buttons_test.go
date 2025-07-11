package ui

import (
	"fmt"
)

// Example-Img: [/images/components/basic/buttons/primary-button.png]
func ExamplePrimaryButton() {
	PrimaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World")
	//Output:
}

// Example-Img: [/images/components/basic/buttons/secondary-button.png].
func ExampleSecondaryButton() {
	SecondaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World")
	//Output:
}

// Example-Img: [/images/components/basic/buttons/tertiary-button.png].
func ExampleTertiaryButton() {
	TertiaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World")
	//Output:
}
