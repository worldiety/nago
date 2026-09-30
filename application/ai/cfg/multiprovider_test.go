// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
)

// fakeOpenAI is an OpenAI compatible server which offers the given models and answers with its name.
type fakeOpenAI struct {
	*httptest.Server
	mutex  sync.Mutex
	models []string // requested models
}

func newFakeOpenAI(t *testing.T, name string, models ...string) *fakeOpenAI {
	f := &fakeOpenAI{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			var data []map[string]any
			for _, m := range models {
				data = append(data, map[string]any{"id": m, "object": "model"})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
		case "/v1/chat/completions":
			var req struct {
				Model string `json:"model"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			f.mutex.Lock()
			f.models = append(f.models, req.Model)
			f.mutex.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "1", "object": "chat.completion", "model": req.Model,
				"choices": []any{map[string]any{"index": 0, "finish_reason": "stop",
					"message": map[string]any{"role": "assistant", "content": "Antwort von " + name}}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

func (f *fakeOpenAI) requested() []string {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return slices.Clone(f.models)
}

// TestTwoProviders configures two providers at once and ensures that the operator decides which provider and
// model the assistant uses, instead of the order of the secrets.
func TestTwoProviders(t *testing.T) {
	alpha := newFakeOpenAI(t, "Alpha", "alpha-1")
	beta := newFakeOpenAI(t, "Beta", "beta-1", "beta-2")

	var mgmt Management
	var cfg *application.Configurator
	nagotest.New(t, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.assistanttest")
		cfg = c
		var err error
		mgmt, err = Enable(c)
		if err != nil {
			t.Fatal(err)
		}
	})

	secrets, err := cfg.SecretManagement()
	if err != nil {
		t.Fatal(err)
	}

	addProvider := func(name, url string) {
		id, err := secrets.UseCases.CreateSecret(user.SU(), openai.Settings{Name: name, BaseURL: url + "/v1"})
		if err != nil {
			t.Fatal(err)
		}
		if err := secrets.UseCases.UpdateMySecretGroups(user.SU(), id, []group.ID{group.System}); err != nil {
			t.Fatal(err)
		}
	}

	// the name of alpha sorts first, thus it is the default
	addProvider("Alpha", alpha.URL)
	addProvider("Beta", beta.URL)

	a := mgmt.Assistant
	deadline := time.Now().Add(10 * time.Second)
	var candidates []Candidate
	for len(candidates) != 2 {
		if time.Now().After(deadline) {
			t.Fatalf("providers not loaded, got %d", len(candidates))
		}
		time.Sleep(10 * time.Millisecond)
		candidates, _ = a.Providers(user.SU())
	}

	// the picker offers the models of both providers, named by their provider
	var labels []string
	var betaTwo ModelChoice
	for _, c := range candidates {
		models, err := a.Models(user.SU(), c)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range models {
			o := modelOption{provider: c.Provider, model: m, qualified: true}
			labels = append(labels, o.label())
			if m.ID == "beta-2" {
				betaTwo = NewModelChoice(c.Provider.Identity(), m.ID)
			}
		}
	}

	if !slices.Equal(labels, []string{"Alpha · alpha-1", "Beta · beta-1", "Beta · beta-2"}) {
		t.Fatalf("unexpected picker %v", labels)
	}

	ask := func(settings AssistantSettings) {
		t.Helper()
		chosen, mdl, err := a.Resolve(user.SU(), settings)
		if err != nil {
			t.Fatal(err)
		}

		res, err := chosen.Completions.Complete(context.Background(), user.SU(), completion.Options{
			Model:    mdl,
			Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Hallo"}}}},
		})
		if err != nil {
			t.Fatal(err)
		}

		if completion.FinalAnswer([]completion.Message{res.Message}) == "" {
			t.Fatalf("no answer")
		}
	}

	// without a choice, the first provider answers with its first model
	ask(AssistantSettings{})
	if got := alpha.requested(); !slices.Equal(got, []string{"alpha-1"}) {
		t.Fatalf("alpha requested %v", got)
	}

	// the operator chooses the second provider
	ask(AssistantSettings{Model: betaTwo})
	if got := beta.requested(); !slices.Equal(got, []string{"beta-2"}) {
		t.Fatalf("beta requested %v", got)
	}

	// a model id stored before providers could be chosen still reaches the provider which offers it
	ask(AssistantSettings{Model: "beta-1"})
	if got := beta.requested(); !slices.Equal(got, []string{"beta-2", "beta-1"}) {
		t.Fatalf("beta requested %v", got)
	}
}
