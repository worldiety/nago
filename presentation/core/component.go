package core

import (
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/ora"
)

type Component interface {
	ID() ora.Ptr
	Properties(yield func(Property) bool) // contract of iter.Seq[Property]
	Render() ora.Component
}

func IsDirty(dst Component) bool {
	dirty := false
	Visit(dst)(func(component Component) bool {
		component.Properties(func(property Property) bool {
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

func ClearDirty(dst Component) {
	Visit(dst)(func(component Component) bool {
		component.Properties(func(property Property) bool {
			if property.Dirty() {
				property.SetDirty(false)
			}

			return true
		})

		return true
	})
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
		property.AnyIter(func(a any) bool {
			if c, ok := a.(Component); ok {
				return visitRecursive(c, walker)
			}
			return true
		})
		return true
	})

	return true
}
