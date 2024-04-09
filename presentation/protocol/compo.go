package protocol

// Ptr is a unique identifier or address for a specific allocated property.
type Ptr int64

// Property represents the current value of an allocated property which is uniquely addressed through a pointer
// within the backend process.
// Note, that these pointers are not real pointers and only unique and valid for a specific scope.
type Property[T any] struct {
	Ptr   Ptr `json:"id"`
	Value T   `json:"value"`
}

// ComponentType defines the defined set of components.
type ComponentType string

const (
	ButtonT ComponentType = "button"
)

type Button struct {
	Ptr      Ptr              `json:"id"`
	Type     ComponentType    `json:"type" value:"Button"`
	Caption  Property[string] `json:"caption" description:"Caption of the button"`
	PreIcon  Property[SVGSrc] `json:"preIcon"`
	PostIcon Property[SVGSrc] `json:"postIcon"`
	Color    Property[Color]  `json:"color"`
	Disabled Property[bool]   `json:"disabled"`
	Action   Property[Ptr]    `json:"action"`
	component
}

type Component interface {
	isComponent()
}

type component struct {
}

func (component) isComponent() {}
