package xreflect

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"reflect"
	"sync"
)

var typeCache = map[reflect.Type]TypeID{}
var typeCacheMutex sync.RWMutex

// TypeID is a unique string type name of a type. It is serializable. The current implementation produces the following
// output (not guaranteed, because Go does not guarantee that either):
//
//	(<module path>.)?<type name>([(TypeID,)+])?
//
// Note, that the module path is not the import path, because the denoted import path refers to packages with different
// package identifiers.
type TypeID string

// TypeIDOf returns the unique type id of the type specified by the functions type parameter.
// Even though this uses reflection, the benchmarks show, that this does not cause any heap allocations after the first
// call. Due to type instantiation, there is always a TypeID.
func TypeIDOf[T any]() TypeID {
	id, ok := idFrom(reflect.TypeFor[T]())
	if !ok {
		panic(fmt.Errorf("unreachable")) // due to instantiations this cannot happen, even for same stencils
	}

	return id
}

// TypeIDFrom cannot create an ID for nil interfaces. Remember that an interface is a fat pointer which points
// to a concrete type and the type instance. However, both pointers can be nil (and only this means, that the interface
// is equal to nil), only the instance pointer may be nil or both are not nil. This implementation
// cannot determine an ID for a nil interface. Note, that all interfaces (not just the empty interface) are the
// same and there is no information about the asserted interface type:
//
//	type MyIface interface {
//	   SomeMethod()
//	}
//
//	var x MyIface
//	_, ok := std.TypeIDFrom(x)
//	fmt.Println(ok)  // will print false
func TypeIDFrom(v any) (TypeID, bool) {
	return idFrom(reflect.TypeOf(v))
}

func TypeIDFromOpt(v any) std.Option[TypeID] {
	x, ok := idFrom(reflect.TypeOf(v))
	return std.Option[TypeID]{
		V:     x,
		Valid: ok,
	}
}

func idFrom(t reflect.Type) (TypeID, bool) {
	typeCacheMutex.RLock()
	if id, ok := typeCache[t]; ok {
		typeCacheMutex.RUnlock()
		return id, true
	}
	typeCacheMutex.RUnlock()

	typeCacheMutex.Lock()
	defer typeCacheMutex.Unlock()

	s, ok := typeToString(t)
	if !ok {
		return "", false
	}

	id := TypeID(s)
	typeCache[t] = id
	return id, true
}

func typeToString(t reflect.Type) (string, bool) {
	if t == nil {
		// remember that a variable of a specific interface is nil, just as the empty interface itself, because
		// it has no associated type. There is nothing like a "typed" interface in reflection
		// (well beside the pointer trick, so you may use a nil pointer to a specific interface type)
		return "", false
	}

	if t.Kind() == reflect.Ptr {
		s, ok := typeToString(t.Elem())
		return "*" + s, ok
	}
	if t.PkgPath() != "" {
		return t.PkgPath() + "." + t.Name(), true
	}
	return t.String(), true
}
