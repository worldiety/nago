package protocol

import "reflect"

// Ptr is a unique identifier or address for a specific allocated property.
type Ptr int64

func (p Ptr) Nil() bool {
	return p == 0
}

// Property represents the current value of an allocated property which is uniquely addressed through a pointer
// within the backend process.
// Note, that these pointers are not real pointers and only unique and valid for a specific scope.
type Property[T any] struct {
	Ptr   Ptr `json:"p"`
	Value T   `json:"v"`
}

// ComponentType defines the defined set of components.
type ComponentType string

const (
	ButtonT      ComponentType = "Button"
	GridT        ComponentType = "Grid"
	GridCellT    ComponentType = "GridCell"
	DialogT      ComponentType = "Dialog"
	TextT        ComponentType = "Text"
	PageT        ComponentType = "Page"
	VBoxT        ComponentType = "VBox"
	HBoxT        ComponentType = "HBox"
	SliderT      ComponentType = "Slider"
	ScaffoldT    ComponentType = "Scaffold"
	NumberFieldT ComponentType = "NumberField"
	TextFieldT   ComponentType = "TextField"
	TableT       ComponentType = "Table"
	TableCellT   ComponentType = "TableCell"
	TableRowT    ComponentType = "TableRow"
)

type Component interface {
	isComponent()
}

var Components []reflect.Type

func init() {
	Components = []reflect.Type{
		reflect.TypeOf(Button{}),
		reflect.TypeOf(Page{}),
		reflect.TypeOf(Scaffold{}),
		reflect.TypeOf(VBox{}),
		reflect.TypeOf(HBox{}),
		reflect.TypeOf(TextField{}),
		reflect.TypeOf(Table{}),
		reflect.TypeOf(TableCell{}),
		reflect.TypeOf(TableRow{}),
		reflect.TypeOf(Text{}),
	}
}

type component struct {
}

func (component) isComponent() {}
