package picker_test

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/picker"
)

func ExamplePicker() {
	type Person struct {
		Name string
		Age  int
	}

	persons := []Person{
		{
			Name: "John",
			Age:  20,
		},
		{
			Name: "Jane",
			Age:  30,
		},
	}

	selected := core.AutoState[[]Person](nil)
	picker.Picker[Person]("Ich bin ein picker", persons, selected)
}
