// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"

	"go.wdy.de/nago/pkg/blob"
)

// ServiceEnv is what a [ServiceFactory] may use of the application.
type ServiceEnv struct {
	Context context.Context
	// HandleMethod registers an http handler of the application, e.g. for the call back of a token exchange.
	HandleMethod func(method, pattern string, handler http.HandlerFunc)
	// Origin returns the origin of the application, like https://my-app.example.com or http://localhost:3000.
	Origin func() string
	// Store opens a persistent store of the application.
	Store func(name string) (blob.Store, error)
	// Getenv reads the environment of the process.
	Getenv func(key string) string
	// LookupEnv reads the environment of the process and tells whether the variable is set at all.
	LookupEnv func(key string) (string, bool)
}

// ServiceFactory creates a provider, which is not configured by a secret in the vault but provisions itself, like the
// Nago AI Service. It returns false, if the service is disabled.
type ServiceFactory func(env ServiceEnv) (Provider, bool, error)

var (
	servicesMu sync.RWMutex
	services   = map[string]ServiceFactory{}
)

// RegisterService wires a self-provisioning provider. Like [Register], provider packages call this at package-init
// time, so that the provider is only compiled in and started when the host application side-imports its package.
func RegisterService(name string, f ServiceFactory) {
	servicesMu.Lock()
	defer servicesMu.Unlock()
	services[name] = f
}

// Services returns the registered factories in the order of their names.
func Services() []ServiceFactory {
	servicesMu.RLock()
	defer servicesMu.RUnlock()

	var names []string
	for name := range services {
		names = append(names, name)
	}

	slices.SortFunc(names, strings.Compare)

	var res []ServiceFactory
	for _, name := range names {
		res = append(res, services[name])
	}

	return res
}
