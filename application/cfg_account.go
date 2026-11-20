// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"slices"

	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/presentation/core"
)

// AddAccountSections registers fn for the account view of every window: in nago's scaffold, in
// uiuser.AccountDialog and wherever uiuser.AccountView is built. Several registrations run in the order of
// registration, each on the result of the former.
func (c *Configurator) AddAccountSections(fn uiuser.AccountSections) *Configurator {
	if fn == nil {
		return c
	}

	c.mutex.Lock()
	first := len(c.accountSections) == 0
	c.accountSections = append(c.accountSections, fn)
	c.mutex.Unlock()

	if first {
		// one context value for all registrations, which runs those of the moment of the call
		c.AddContextValue(core.ContextValue("", uiuser.AccountSections(func(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
			c.mutex.Lock()
			fns := slices.Clone(c.accountSections)
			c.mutex.Unlock()

			for _, fn := range fns {
				sections = fn(wnd, sections)
			}

			return sections
		})))
	}

	return c
}
