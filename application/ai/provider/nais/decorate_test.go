// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nais_test

import (
	"errors"
	"testing"

	"go.wdy.de/nago/application"
	cfgai "go.wdy.de/nago/application/ai/cfg"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/provider/nais"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
)

// metered marks a provider as wrapped by the application.
type metered struct {
	provider.Provider
}

func TestApplicationDecoratesEveryProvider(t *testing.T) {
	// the service needs no server to be there, it enrolls on first use only
	t.Setenv(nais.EnvService, "https://ai.example.com")

	var mgmt cfgai.Management
	var cfg *application.Configurator
	nagotest.New(t, func(c *application.Configurator) {
		cfg = c
		var err error
		if mgmt, err = cfgai.Enable(c); err != nil {
			t.Fatal(err)
		}
	})

	secrets, err := cfg.SecretManagement()
	if err != nil {
		t.Fatal(err)
	}

	id, err := secrets.UseCases.CreateSecret(user.SU(), openai.Settings{Name: "OpenAI", BaseURL: "http://localhost:1/v1"})
	if err != nil {
		t.Fatal(err)
	}

	if err := secrets.UseCases.UpdateMySecretGroups(user.SU(), id, []group.ID{group.System}); err != nil {
		t.Fatal(err)
	}

	// a failing decorator is reported, so that an application can refuse to start without it
	if err := mgmt.DecorateProviders("meter", func(p provider.Provider) (provider.Provider, error) {
		return nil, errors.New("store down")
	}); err == nil {
		t.Fatal("the failure of the decorator must be returned")
	}

	if err := mgmt.DecorateProviders("meter", func(p provider.Provider) (provider.Provider, error) {
		return metered{Provider: p}, nil
	}); err != nil {
		t.Fatal(err)
	}

	found := map[provider.ID]bool{}
	for p, err := range mgmt.UseCases.FindAllProvider(user.SU()) {
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := p.(metered); !ok {
			t.Fatalf("provider %s escaped the decoration: %T", p.Identity(), p)
		}

		found[p.Identity()] = true
	}

	if !found[nais.ID] || !found[provider.ID(id)] {
		t.Fatalf("expected the service and the secret, got %v", found)
	}
}
