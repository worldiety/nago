// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"fmt"
	"slices"
	"sync"

	"go.wdy.de/nago/application/ai/provider"
)

// Decorator wraps a provider, e.g. to meter or limit its use per subject. It is applied to every provider, whether
// it comes from a secret of the vault or from a service like the Nago AI Service, and again whenever the providers
// are reloaded. A decorator therefore
//   - must keep the [provider.Provider.Identity] of the provider it wraps, which sessions and files refer to,
//   - must keep any state, like counters, outside of the wrapper, e.g. in a store, because each reload wraps anew,
//   - should account a stream also when its consumer stops early, as the usage only comes with the last delta,
//   - must not be bypassed by caching providers: callers resolve them by id whenever they need one.
type Decorator func(p provider.Provider) (provider.Provider, error)

// Decorators are the decorators of the application, applied in the order they were first set. Setting a name again
// replaces its decorator in place, so that a repeated registration does not wrap twice.
type Decorators struct {
	mutex sync.Mutex
	names []string
	fns   map[string]Decorator
}

// Set registers or replaces the decorator of the name.
func (d *Decorators) Set(name string, fn Decorator) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.fns == nil {
		d.fns = map[string]Decorator{}
	}

	if !slices.Contains(d.names, name) {
		d.names = append(d.names, name)
	}

	d.fns[name] = fn
}

// Apply wraps the provider in every decorator, the first one innermost.
func (d *Decorators) Apply(p provider.Provider) (provider.Provider, error) {
	d.mutex.Lock()
	names := slices.Clone(d.names)
	fns := make([]Decorator, 0, len(names))
	for _, name := range names {
		fns = append(fns, d.fns[name])
	}
	d.mutex.Unlock()

	for i, fn := range fns {
		decorated, err := checked(p, fn)
		if err != nil {
			return nil, fmt.Errorf("decorator %s: %w", names[i], err)
		}

		p = decorated
	}

	return p, nil
}

// checked applies the decorator and refuses what would break the provider for its callers.
func checked(p provider.Provider, fn Decorator) (provider.Provider, error) {
	decorated, err := fn(p)
	if err != nil {
		return nil, err
	}

	if decorated == nil {
		return nil, fmt.Errorf("the decorator returned no provider")
	}

	if decorated.Identity() != p.Identity() {
		return nil, fmt.Errorf("the decorator changed the identity of the provider from %s to %s", p.Identity(), decorated.Identity())
	}

	return decorated, nil
}
