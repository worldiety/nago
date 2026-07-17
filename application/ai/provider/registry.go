// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"reflect"
	"sync"

	"go.wdy.de/nago/application/secret"
)

// factory turns a concrete [secret.Credentials] value into a [Provider]. The registry stores one factory per
// concrete credentials type; the type-erased signature lets the reload loop dispatch without importing any
// concrete provider package.
type factory func(id ID, cfg secret.Credentials) Provider

var (
	registryMu sync.RWMutex
	registry   = map[reflect.Type]factory{}
)

// Register wires a provider factory to the concrete credentials type C. Provider packages call this at
// package-init time (next to their enum.Variant registration), so a provider is only compiled in and made
// available when the host application side-imports it (e.g. _ "go.wdy.de/nago/application/ai/provider/anthropic").
//
// The generic wrapper asserts the incoming secret.Credentials back to C, so factories keep their natural typed
// signature (func(ID, C) Provider) and the reload loop stays free of any concrete provider dependency.
func Register[C secret.Credentials](f func(id ID, cfg C) Provider) {
	var zero C
	t := reflect.TypeOf(zero)

	registryMu.Lock()
	defer registryMu.Unlock()
	registry[t] = func(id ID, cfg secret.Credentials) Provider {
		return f(id, cfg.(C))
	}
}

// NewProviderFor looks up the factory registered for the concrete type of cfg and builds a provider. The
// second return value is false when no provider package registered that credentials type (i.e. it was not
// side-imported), in which case the caller should skip the secret.
func NewProviderFor(id ID, cfg secret.Credentials) (Provider, bool) {
	registryMu.RLock()
	f, ok := registry[reflect.TypeOf(cfg)]
	registryMu.RUnlock()
	if !ok {
		return nil, false
	}
	return f(id, cfg), true
}
