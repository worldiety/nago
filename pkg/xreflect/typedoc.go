package xreflect

import (
	"fmt"
	"reflect"
)

func PackageOf(path string) (*Package, bool) {
	p, ok := packages[path]
	return p, ok
}

// TypeFor returns the best Type information possible. At worst, an *Incomplete is returned.
func TypeFor[T any]() Type {
	mutex.Lock()
	defer mutex.Unlock()

	// fast path
	typ := reflect.TypeFor[T]()
	if doc, ok := typeLookup[typ]; ok {
		return doc
	}

	// pkgname is indeed the import path
	pkgName := typ.PkgPath()
	if pkgName == "" {
		return &Incomplete{
			doc:   "",
			path:  typ.PkgPath(),
			name:  typ.Name(),
			error: fmt.Errorf("package less universe type is unsupported"),
		}
	}

	name := typ.Name()

	if types, ok := typesInPackage[pkgName]; ok {
		for _, t := range types {
			if t.Name() == name {
				return t
			}
		}
	}

	return &Incomplete{
		doc:   "",
		path:  typ.PkgPath(),
		name:  typ.Name(),
		error: fmt.Errorf("no AST type information available: try to xreflect.Import the according package"),
	}
}
