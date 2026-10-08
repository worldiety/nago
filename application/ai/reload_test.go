// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai_test

import (
	"errors"
	"iter"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

// testCredentials is a provider type of its own, so that the test counts how often a provider is built.
type testCredentials struct {
	Name string
}

func (c testCredentials) GetName() string   { return c.Name }
func (c testCredentials) Credentials() bool { return true }
func (c testCredentials) IsZero() bool      { return c == testCredentials{} }

var built atomic.Int32

var _ = registerTestProvider()

func registerTestProvider() any {
	provider.Register[testCredentials](func(id provider.ID, cfg testCredentials) provider.Provider {
		built.Add(1)
		return testProvider{id: id, name: cfg.Name}
	})
	return nil
}

type testProvider struct {
	id   provider.ID
	name string
}

func (p testProvider) Identity() provider.ID             { return p.id }
func (p testProvider) Name() string                      { return p.name }
func (p testProvider) Description() string               { return "" }
func (p testProvider) Files() option.Opt[provider.Files] { return option.None[provider.Files]() }
func (p testProvider) Completions() option.Opt[completion.Completions] {
	return option.None[completion.Completions]()
}

// wrapped marks a provider as decorated by a name.
type wrapped struct {
	provider.Provider
	by string
}

// vault holds the secrets of the system group.
type vault struct {
	mutex   sync.Mutex
	secrets []secret.Secret
}

func (v *vault) set(secrets ...secret.Secret) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.secrets = secrets
}

func (v *vault) find(auth.Subject, group.ID) iter.Seq2[secret.Secret, error] {
	v.mutex.Lock()
	list := append([]secret.Secret(nil), v.secrets...)
	v.mutex.Unlock()

	return func(yield func(secret.Secret, error) bool) {
		for _, s := range list {
			if !yield(s, nil) {
				return
			}
		}
	}
}

func layers(p provider.Provider) []string {
	var res []string
	for {
		w, ok := p.(wrapped)
		if !ok {
			return res
		}

		res = append(res, w.by)
		p = w.Provider
	}
}

func setup(t *testing.T, services ...provider.Provider) (*vault, *ai.Decorators, *concurrent.RWMap[provider.ID, provider.Provider], ai.ReloadProvider) {
	t.Helper()
	v := &vault{}
	v.set(secret.Secret{ID: "a", Credentials: testCredentials{Name: "A"}}, secret.Secret{ID: "b", Credentials: testCredentials{Name: "B"}})

	d := &ai.Decorators{}
	inner := func(p provider.Provider) (provider.Provider, error) {
		return d.Apply(wrapped{Provider: p, by: "nago"})
	}

	var m concurrent.RWMap[provider.ID, provider.Provider]
	return v, d, &m, ai.NewReloadProvider(&m, v.find, inner, services...)
}

func TestEveryProviderIsDecoratedOnce(t *testing.T) {
	service := testProvider{id: "nais", name: "Service"}
	_, d, m, reload := setup(t, service)

	d.Set("meter", func(p provider.Provider) (provider.Provider, error) { return wrapped{Provider: p, by: "meter"}, nil })
	d.Set("limit", func(p provider.Provider) (provider.Provider, error) { return wrapped{Provider: p, by: "limit"}, nil })
	// registering a name again replaces it in place and does not wrap twice
	d.Set("meter", func(p provider.Provider) (provider.Provider, error) { return wrapped{Provider: p, by: "meter2"}, nil })

	for range 3 {
		if err := reload(user.SU()); err != nil {
			t.Fatal(err)
		}
	}

	if m.Len() != 3 {
		t.Fatalf("expected the service and two secrets, got %d", m.Len())
	}

	for id, p := range m.All() {
		got := layers(p)
		if len(got) != 3 || got[0] != "limit" || got[1] != "meter2" || got[2] != "nago" {
			t.Fatalf("%s: expected limit around meter2 around nago, got %v", id, got)
		}
	}
}

func TestProvidersAreOnlyBuiltForNewCredentials(t *testing.T) {
	v, _, m, reload := setup(t)
	before := built.Load()

	for range 3 {
		if err := reload(user.SU()); err != nil {
			t.Fatal(err)
		}
	}

	if n := built.Load() - before; n != 2 {
		t.Fatalf("each secret must be built once, got %d builds", n)
	}

	// a changed secret is built anew, a deleted one is gone
	v.set(secret.Secret{ID: "a", Credentials: testCredentials{Name: "A2"}})
	if err := reload(user.SU()); err != nil {
		t.Fatal(err)
	}

	p, ok := m.Get("a")
	if n := built.Load() - before; n != 3 || !ok || p.Name() != "A2" || m.Len() != 1 {
		t.Fatalf("expected a rebuilt a only, got %d builds, %v", n, p)
	}
}

func TestFailingDecoratorLeavesOutTheProvider(t *testing.T) {
	_, d, m, reload := setup(t)

	for name, fn := range map[string]ai.Decorator{
		"error": func(p provider.Provider) (provider.Provider, error) { return nil, errors.New("store down") },
		"nil":   func(p provider.Provider) (provider.Provider, error) { return nil, nil },
		"id":    func(p provider.Provider) (provider.Provider, error) { return testProvider{id: "other"}, nil },
	} {
		d.Set("bad", fn)
		err := reload(user.SU())
		if err == nil {
			t.Fatalf("%s: the failure must be reported", name)
		}

		if m.Len() != 0 {
			t.Fatalf("%s: an undecorated provider must not be published, got %d", name, m.Len())
		}
	}

	d.Set("bad", func(p provider.Provider) (provider.Provider, error) { return p, nil })
	if err := reload(user.SU()); err != nil || m.Len() != 2 {
		t.Fatalf("a fixed decorator must publish again, got %d %v", m.Len(), err)
	}
}

func TestReadersNeverSeeAHalfReload(t *testing.T) {
	_, _, m, reload := setup(t)
	if err := reload(user.SU()); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				_ = reload(user.SU())
			}
		}()
	}

	go func() {
		wg.Wait()
		close(stop)
	}()

	for {
		select {
		case <-stop:
			return
		default:
			if n := m.Len(); n != 2 {
				t.Fatalf("a reader found %d providers in the middle of a reload", n)
			}
		}
	}
}
