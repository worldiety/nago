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
	Ptr   Ptr `json:"id"`
	Value T   `json:"value"`
}

// ComponentType defines the defined set of components.
type ComponentType string

const (
	ButtonT ComponentType = "button"
)

type Component interface {
	isComponent()
}

var Components []reflect.Type

func init() {
	Components = []reflect.Type{
		reflect.TypeOf(Button{}),
	}
}

type component struct {
}

func (component) isComponent() {}
