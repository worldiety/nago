// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"strings"
	"sync"
	"time"

	"github.com/worldiety/enum"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/settings"
	"golang.org/x/text/language"
)

// SourceAssistantModels is the context name the model picker of [AssistantSettings] resolves through. It is
// registered by [Enable] and lists the models all configured providers actually offer.
const SourceAssistantModels = "nago.ai.assistant.models"

// DefaultAssistantMaxTokens caps the generated output per answer when the operator did not choose a limit.
const DefaultAssistantMaxTokens = 32000

var (
	StrAssistantSettingsTitle = i18n.MustString("nago.ai.assistant.settings.title", i18n.Values{
		language.German:  "KI-Assistent",
		language.English: "AI Assistant",
	})
	StrAssistantSettingsDesc = i18n.MustString("nago.ai.assistant.settings.desc", i18n.Values{
		language.German:  "Modell und Verhalten des KI-Assistenten.",
		language.English: "Model and behaviour of the AI assistant.",
	})
)

// AssistantSettings is what an operator decides about the assistant.
//
// These are global settings rather than constants in an application's source because every field is an
// operational choice, not a design one: which model is appropriate, how long an answer may get, and whether
// the assistant may change anything at all. Those decisions belong to whoever runs the installation and have
// to be revisable without a deployment.
//
// nago renders the administration form for this type by itself; the struct tags below are the entire user
// interface. Applications do not declare their own copy of this - it ships with [Enable].
type AssistantSettings struct {
	_ any `title:"nago.ai.assistant.settings.title" description:"nago.ai.assistant.settings.desc"`

	// Model is picked from what the configured providers actually offer, so a model that has been retired
	// cannot be selected by accident. It names the provider as well, because with several providers a model
	// id alone does not tell which one to ask.
	Model ModelChoice `json:"model" label:"Standardmodell" source:"nago.ai.assistant.models" supportingText:"Provider und Modell für alle Unterhaltungen. Ist nichts gewählt, nimmt der Assistent den ersten Provider und dessen erstes Modell. Fehlt ein Provider in der Liste, ist er nicht erreichbar — meist ein falscher API-Token im Tresor."`

	MaxTokens int `json:"maxTokens" label:"Maximale Antwortlänge" supportingText:"Obergrenze der erzeugten Tokens je Antwort. Null bedeutet den eingebauten Standard von 32000. Die Denkphase des Modells zählt mit."`

	// ReadOnly is phrased positively rather than as "disable writes" because it describes a state the
	// operator can reason about, and because the zero value must be the behaviour the application was built
	// for. An application that ships no mutating tools is unaffected either way.
	ReadOnly bool `json:"readOnly" label:"Nur lesender Zugriff" section:"Sicherheit" supportingText:"Der Assistent kann dann nur noch lesen und erklären. Schreibende Werkzeuge werden gar nicht erst angeboten, das Modell erfährt nichts von ihnen."`

	// SkipConfirmation is phrased so that its zero value is the safe behaviour: unless an operator decides
	// otherwise, the user approves every change before it happens. An assistant button may still override this
	// through [AssistantOptions.Confirmation].
	SkipConfirmation bool `json:"skipConfirmation" label:"Änderungen ohne Rückfrage ausführen" section:"Sicherheit" supportingText:"Schreibende Werkzeuge laufen dann ohne Bestätigung. Ist die Option aus, muss jede Änderung einzeln freigegeben werden, oder nur kritische, siehe unten. Einzelne Assistenten einer Anwendung können davon abweichen."`

	// ConfirmMarkedOnly narrows the confirmation to the changes an application marked as requiring approval
	// ([completion.Tool.RequiresApproval], [completion.Tool.ApprovalFor]), e.g. deleting data or granting rights.
	// It is a separate flag rather than a third state of SkipConfirmation, so that stored settings keep their
	// meaning. SkipConfirmation wins.
	ConfirmMarkedOnly bool `json:"confirmMarkedOnly" label:"Nur bei kritischen Änderungen nachfragen" section:"Sicherheit" supportingText:"Fragt nur vor Änderungen, die die Anwendung als bestätigungspflichtig markiert hat, etwa Löschen oder Rechtevergabe. Andere Änderungen laufen ohne Rückfrage. Ohne Wirkung, wenn Änderungen ohne Rückfrage ausgeführt werden."`

	// Hidden removes the button without removing the provider, which is what an operator wants while
	// investigating something rather than while decommissioning it.
	Hidden bool `json:"hidden" label:"Assistent ausblenden" section:"Sicherheit" supportingText:"Blendet den Knopf auf allen Seiten aus, ohne den Provider zu entfernen."`
}

// ModelChoice identifies a model of a specific provider as "<provider id>/<model id>", see [NewModelChoice].
//
// A value which does not start with a provider id is a bare model id, as stored before a provider could be
// chosen. It is used unchanged by whichever provider offers that model, see [Assistant.Resolve].
type ModelChoice string

// NewModelChoice combines a provider and one of its models. Provider ids are secret ids and never contain a
// slash, while model ids may, e.g. "openai/gpt-4o" at OpenRouter.
func NewModelChoice(p provider.ID, m model.ID) ModelChoice {
	return ModelChoice(string(p) + "/" + string(m))
}

// split separates the provider and the model, if the prefix is a provider for which known returns true.
func (c ModelChoice) split(known func(provider.ID) bool) (provider.ID, model.ID, bool) {
	p, m, ok := strings.Cut(string(c), "/")
	if !ok || p == "" || m == "" || !known(provider.ID(p)) {
		return "", "", false
	}

	return provider.ID(p), model.ID(m), true
}

// namesProvider reports whether the choice starts with something formatted like a provider id, which is a
// secret id of 32 hex digits. It tells a choice whose provider has been deleted apart from a bare model id
// like "openai/gpt-4o", which never starts like that.
func (c ModelChoice) namesProvider() bool {
	p, _, ok := strings.Cut(string(c), "/")
	if !ok || len(p) != 32 {
		return false
	}

	for _, r := range p {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}

	return true
}

// GlobalSettings marks this as an application-wide setting.
func (AssistantSettings) GlobalSettings() bool { return true }

var _ = enum.Variant[settings.GlobalSettings, AssistantSettings](
	enum.Rename[AssistantSettings]("nago.ai.assistant.settings"),
)

// MaxTokensOr applies the built-in default.
func (s AssistantSettings) MaxTokensOr(fallback int) int {
	if s.MaxTokens > 0 {
		return s.MaxTokens
	}

	return fallback
}

const (
	// modelCacheHit is how long a successful listing is trusted. A provider does not retire a model between
	// two clicks; an administration that has just added one waits at most this long to see it.
	modelCacheHit = 10 * time.Minute

	// modelCacheMiss is how long a failure is remembered. Short, because the usual cause is a wrong token and
	// somebody is fixing it right now - but not zero, because a failing network call on every render is
	// exactly what this cache exists to prevent.
	modelCacheMiss = 30 * time.Second
)

// modelCache remembers which models the provider offers.
//
// It exists because asking is a network call and rendering is not, and two places ask often: the assistant
// button sits in a decorator and would otherwise resolve its model on every single page view, and the model
// picker on the settings page is walked by the form renderer as one listing plus one lookup per entry.
//
// Measured in a production application before it existed: 190 to 400 ms per page render for the assistant
// against 60 µs for the whole rest of the page frame, and two and a half to three seconds for the settings
// page. Neither is an answer that changes between two clicks.
type modelCache struct {
	mutex  sync.Mutex
	models []model.Model
	err    error
	until  time.Time
}

// list returns the provider's models, fetching at most once per interval.
func (c *modelCache) list(fetch func() ([]model.Model, error)) ([]model.Model, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if time.Now().Before(c.until) {
		return c.models, c.err
	}

	c.models, c.err = fetch()
	if c.err != nil {
		c.until = time.Now().Add(modelCacheMiss)
	} else {
		c.until = time.Now().Add(modelCacheHit)
	}

	return c.models, c.err
}

// forget drops the cache so the next caller asks again.
//
// It is what makes the ten minute lifetime acceptable: an administrator who has just fixed the token does not
// have to wait it out, because saving the settings or the secret clears it.
func (c *modelCache) forget() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.models, c.err, c.until = nil, nil, time.Time{}
}
