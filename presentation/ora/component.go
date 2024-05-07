package ora

import "reflect"

// Ptr is a unique identifier or address for a specific allocated property.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Ptr int64

func (p Ptr) Nil() bool {
	return p == 0
}

// Property represents the current value of an allocated property which is uniquely addressed through a pointer
// within the backend process.
// Note, that these pointers are not real pointers and only unique and valid for a specific scope.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Property[T any] struct {
	Ptr   Ptr `json:"p"`
	Value T   `json:"v"`
}

// ComponentType defines the defined set of components.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ComponentType string

const (
	ButtonT        ComponentType = "Button"
	GridT          ComponentType = "Grid"
	GridCellT      ComponentType = "GridCell"
	DialogT        ComponentType = "Dialog"
	TextT          ComponentType = "Text"
	PageT          ComponentType = "Page"
	VBoxT          ComponentType = "VBox"
	HBoxT          ComponentType = "HBox"
	SliderT        ComponentType = "Slider"
	ScaffoldT      ComponentType = "Scaffold"
	NumberFieldT   ComponentType = "NumberField"
	TextFieldT     ComponentType = "TextField"
	PasswordFieldT ComponentType = "PasswordField"
	TableT         ComponentType = "Table"
	TableCellT     ComponentType = "TableCell"
	TableRowT      ComponentType = "TableRow"
	ToggleT        ComponentType = "Toggle"
	DatePickerT    ComponentType = "DatePicker"
	DividerT       ComponentType = "Divider"
	DropdownT      ComponentType = "Dropdown"
	DropdownItemT  ComponentType = "DropdownItem"
	ChipT          ComponentType = "Chip"
	CardT          ComponentType = "Card"
	StepperT       ComponentType = "Stepper"
	StepInfoT      ComponentType = "StepInfo"
	WebViewT       ComponentType = "WebView"
	TextAreaT      ComponentType = "TextArea"
	FileFieldT     ComponentType = "FileField"
	ImageT         ComponentType = "Image"
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
		reflect.TypeOf(PasswordField{}),
		reflect.TypeOf(Table{}),
		reflect.TypeOf(TableCell{}),
		reflect.TypeOf(TableRow{}),
		reflect.TypeOf(Text{}),
		reflect.TypeOf(Dialog{}),
		reflect.TypeOf(Toggle{}),
		reflect.TypeOf(DatePicker{}),
		reflect.TypeOf(NumberField{}),
		reflect.TypeOf(Slider{}),
		reflect.TypeOf(Divider{}),
		reflect.TypeOf(Dropdown{}),
		reflect.TypeOf(DropdownItem{}),
		reflect.TypeOf(Chip{}),
		reflect.TypeOf(Card{}),
		reflect.TypeOf(Stepper{}),
		reflect.TypeOf(StepInfo{}),
		reflect.TypeOf(WebView{}),
		reflect.TypeOf(TextArea{}),
		reflect.TypeOf(FileField{}),
		reflect.TypeOf(Image{}),
		reflect.TypeOf(Grid{}),
		reflect.TypeOf(GridCell{}),
	}
}

type component struct {
}

func (component) isComponent() {}
