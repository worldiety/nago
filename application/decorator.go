// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import "go.wdy.de/nago/presentation/core"

type Decorator func(wnd core.Window, view core.View) core.View

// Decorator returns the default system page decorator.
func (c *Configurator) Decorator() Decorator {
	if c.decorator == nil {
		c.decorator = func(wnd core.Window, view core.View) core.View {
			return view
		}
	}

	return c.decorator
}

// SetDecorator replaces the current decorator instance.
func (c *Configurator) SetDecorator(decorator Decorator) *Configurator {
	c.decorator = decorator
	return c
}

// DecorateRootView uses the current Decorator to wrap a root view factory within the decorator.
func (c *Configurator) DecorateRootView(factory func(wnd core.Window) core.View) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		view := factory(wnd)
		return c.Decorator()(wnd, view)
	}
}
