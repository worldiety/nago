package enum

import (
	"fmt"
	"reflect"
)

// see also https://serde.rs/enum-representations.html#adjacently-tagged.
// we don't use the default variant, because it is ineffcient to express that naturally in go due
// to map usage.
type adjacentlyTagged[T any] struct {
	Type  string `json:"type,omitempty"`
	Value T      `json:"value,omitempty"`
}

type adjacentlyTaggedPreflight struct {
	Type string `json:"type"`
}

type Enumeration interface {
	Ordinal() int
	Unwrap() any
	Nil() bool
}

func toString(e Enumeration) string {
	if e.Nil() {
		return "nil"
	}

	if _, isEnum := e.Unwrap().(Enumeration); !isEnum {
		if err, ok := e.Unwrap().(error); ok {
			return err.Error()
		}

		return fmt.Sprintf("%s: %v", reflect.TypeOf(e.Unwrap()).Name(), e.Unwrap())
	}

	return fmt.Sprintf("%v", e.Unwrap())
}
