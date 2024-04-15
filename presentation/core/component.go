package core

import (
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/protocol"
)

type Component interface {
	ID() protocol.Ptr
	Properties(yield func(Property) bool) // contract of iter.Seq[Property]
	Render() protocol.Component
}

func IsDirty(dst Component) bool {
	dirty := false
	Visit(dst)(func(component Component) bool {
		dst.Properties(func(property Property) bool {
			if property.Dirty() {
				dirty = true
				return false
			}

			return true
		})

		if dirty {
			return false
		}

		return true
	})

	return dirty
}

func Visit(root Component) iter.Seq[Component] {
	return func(yield func(Component) bool) {
		visitRecursive(root, func(component2 Component) bool {
			return yield(component2)
		})
	}
}

func visitRecursive(root Component, walker func(Component) bool) bool {
	if root == nil {
		// by definition legal, properties may have just nil components but we don't want to visit them
		return true
	}

	if !walker(root) {
		return false
	}

	root.Properties(func(property Property) bool {
		switch p := property.(type) {

		case Iterable[Component]:
			p.Iter(func(c Component) bool {
				return visitRecursive(c, walker)
			})
		}
		return true
	})

	return true
}
