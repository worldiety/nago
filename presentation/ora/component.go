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
	// Ptr is short for "Pointer" and references a property instance within the backend.
	// Because it is so common, the json field name is just p.
	Ptr Ptr `json:"p"` // TODO the frontend is not prepared for omitempty but we definitely should?
	// Value contains the actual value specified by the generic type parameter and shortend to v in json.
	Value T `json:"v"` // TODO the frontend is not prepared for omitempty but we definitely should?
}

// ComponentType defines the defined set of components.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ComponentType string

const (
	ButtonT              ComponentType = "Button"
	GridT                ComponentType = "G"
	GridCellT            ComponentType = "C"
	DialogT              ComponentType = "Dialog"
	TextT                ComponentType = "T"
	PageT                ComponentType = "Page"
	VBoxT                ComponentType = "VBox"
	HBoxT                ComponentType = "HBox"
	SliderT              ComponentType = "Slider"
	ScaffoldT            ComponentType = "A"
	NavigationComponentT ComponentType = "NavigationComponent"
	MenuEntryT           ComponentType = "MenuEntry"
	NumberFieldT         ComponentType = "NumberField"
	TextFieldT           ComponentType = "F"
	PasswordFieldT       ComponentType = "p"
	TableT               ComponentType = "B"
	ToggleT              ComponentType = "t"
	DatePickerT          ComponentType = "P"
	DividerT             ComponentType = "d"
	DropdownT            ComponentType = "Dropdown"
	DropdownItemT        ComponentType = "DropdownItem"
	ChipT                ComponentType = "Chip"
	CardT                ComponentType = "Card"
	StepperT             ComponentType = "Stepper"
	StepInfoT            ComponentType = "StepInfo"
	WebViewT             ComponentType = "WebView"
	TextAreaT            ComponentType = "TextArea"
	FileFieldT           ComponentType = "FileField"
	ImageT               ComponentType = "I"
	BreadcrumbsT         ComponentType = "Breadcrumbs"
	BreadcrumbItemT      ComponentType = "BreadcrumbItem"
	CheckboxT            ComponentType = "c"
	RadiobuttonT         ComponentType = "R"
	FlexContainerT       ComponentType = "FlexContainer"
	ProgressBarT         ComponentType = "ProgressBar"
	StrT                 ComponentType = "S"
	HStackT              ComponentType = "hs"
	VStackT              ComponentType = "vs"
	BoxT                 ComponentType = "bx"
	SpacerT              ComponentType = "s"
	ModalT               ComponentType = "M"
	WindowTitleT         ComponentType = "W"
)

type Component interface {
	isComponent()
	// VisitFuncs(yield()bool)
	// VisitShared(yield()bool)
}

var Components []reflect.Type

func Slice(elems ...Component) []Component {
	return elems
}

func init() {
	Components = []reflect.Type{
		reflect.TypeOf(Button{}),
		reflect.TypeOf(Page{}),
		reflect.TypeOf(Scaffold{}),
		reflect.TypeOf(TextField{}),
		reflect.TypeOf(PasswordField{}),
		reflect.TypeOf(Table{}),
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
		reflect.TypeOf(FileField{}),
		reflect.TypeOf(Image{}),
		reflect.TypeOf(Breadcrumbs{}),
		reflect.TypeOf(BreadcrumbItem{}),
		reflect.TypeOf(Grid{}),
		reflect.TypeOf(GridCell{}),
		reflect.TypeOf(Checkbox{}),
		reflect.TypeOf(Radiobutton{}),
		reflect.TypeOf(FlexContainer{}),
		reflect.TypeOf(HStack{}),
		reflect.TypeOf(VStack{}),
		reflect.TypeOf(Box{}),
		reflect.TypeOf(Spacer{}),
		reflect.TypeOf(Modal{}),
		reflect.TypeOf(WindowTitle{}),
	}
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type _component interface {
	WindowTitle | Modal | Spacer | Box | VStack | HStack | Button | Page | Scaffold | TextField | PasswordField | Table | Text | Dialog | Toggle | DatePicker | NumberField | Slider | Divider | Dropdown | DropdownItem | Chip | Card | Stepper | StepInfo | WebView | FileField | Image | Breadcrumbs | BreadcrumbItem | Grid | GridCell | FlexContainer
}

type component struct {
}

func (component) isComponent() {}
