package annotation

import (
	"fmt"
	"reflect"
)

// DocElem allows a bit more type safety and better autocompletion.
type DocElem string

// Proposal: TypeLink has the advantage, that we can resolve the type trivially across any package or module boundaries at runtime.
func TypeLink[T any](name string) DocElem {
	t := reflect.TypeFor[T]()
	return DocElem(fmt.Sprintf("[%s](%s)", t.String(), name))
}
