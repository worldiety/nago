// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/filecleanup"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrResSessions = i18n.MustString("nago.ai.session.resources.name", i18n.Values{language.English: "AI Sessions", language.German: "KI Sitzungen"})
	StrResSessDesc = i18n.MustString("nago.ai.session.resources.desc", i18n.Values{language.English: "Persisted, provider-independent AI chat sessions with their full message history.", language.German: "Persistierte, providerunabhängige KI-Chat-Sitzungen mit vollständigem Nachrichtenverlauf."})

	StrRoleAssistantUserName = i18n.MustString("nago.ai.role.assistant_user.name", i18n.Values{
		language.English: "AI Assistant User",
		language.German:  "KI-Assistent Nutzer",
	})
	StrRoleAssistantUserDesc = i18n.MustString("nago.ai.role.assistant_user.desc", i18n.Values{
		language.English: "Allows using the AI assistant. It grants no access to any business data: the assistant always acts with the permissions of the user operating it.",
		language.German:  "Erlaubt die Nutzung des KI-Assistenten. Sie gewährt keinerlei Zugriff auf Fachdaten: Der Assistent handelt immer mit den Berechtigungen des Nutzers, der ihn bedient.",
	})
)

// RoleAssistantUser is the system role that makes the AI assistant available to a user. Assign it and the
// chat button appears; the assistant can then use exactly those use cases the user could reach by hand.
//
// It is declared by [Enable] and protected against deletion and permission edits (see
// [application.Configurator.DeclareSystemRole]).
const RoleAssistantUser role.ID = "nago.ai.assistant.user"

type Management struct {
	UseCases        ai.UseCases
	SessionUseCases session.UseCases

	// Assistant is the ready-to-use chat assistant: provider lookup, model resolution, operator settings and
	// the floating button. See [Assistant.Decorate].
	Assistant *Assistant

	sessionDeleteHooks *sessionDeleteHooks
}

// sessionDeleteHooks are the hooks of [Management.OnSessionDelete].
type sessionDeleteHooks struct {
	mutex sync.RWMutex
	hooks []func(id session.ID) error
}

func (h *sessionDeleteHooks) run(id session.ID) error {
	h.mutex.RLock()
	hooks := slices.Clone(h.hooks)
	h.mutex.RUnlock()

	for _, fn := range hooks {
		if err := fn(id); err != nil {
			return err
		}
	}

	return nil
}

// OnSessionDelete registers a hook, which is invoked for every session right before it is deleted, also for each
// of its child sessions, e.g. to delete what an application keeps per conversation, like the attachments it took
// over with OnAttach. An error aborts the deletion, so that nothing is orphaned. Register it while configuring the
// application.
func (m Management) OnSessionDelete(fn func(id session.ID) error) {
	m.sessionDeleteHooks.mutex.Lock()
	defer m.sessionDeleteHooks.mutex.Unlock()

	m.sessionDeleteHooks.hooks = append(m.sessionDeleteHooks.hooks, fn)
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	secrets, err := cfg.SecretManagement()
	if err != nil {
		return Management{}, err
	}

	// Files a conversation uploaded to a provider are recorded and deleted together with the conversation, see
	// the filecleanup package. The ledger resolves the providers of ucAI, which in turn decorates them with it.
	repoProviderFiles, err := application.JSONRepository[filecleanup.Entry](cfg, "nago.ai.provider_file")
	if err != nil {
		return Management{}, err
	}

	var ucAI ai.UseCases
	providerFiles := filecleanup.New(repoProviderFiles, func(id provider.ID) (option.Opt[provider.Provider], error) {
		return ucAI.FindProviderByID(cfg.SysUser(), id)
	}, cfg.SysUser)
	ucAI = ai.NewUseCases(cfg.EventBus(), secrets.UseCases.FindGroupSecrets, providerFiles.Decorate)
	go providerFiles.Run(cfg.Context())

	// Sessions are provider-independent, locally persisted chats on top of the stateless completion API. The
	// whole (lossless) history lives in this repository.
	repoSessions, err := application.JSONRepository[session.Session](cfg, string(session.Namespace))
	if err != nil {
		return Management{}, err
	}

	rdb, err := cfg.RDB()
	if err != nil {
		return Management{}, err
	}

	// Register the ReBAC static rules that allow granting a user ownership and the per-instance permissions
	// on a specific session (required before rebac.DB.Put may write those triples). The global user ->
	// global rules for these permissions are already registered by the user management for every permission.
	rdb.RegisterStaticRule(rebac.StaticRule{
		Source:   user.Namespace,
		Relation: rebac.Owner,
		Target:   session.Namespace,
	})
	for _, pid := range session.InstancePermissions {
		rdb.RegisterStaticRule(rebac.StaticRule{
			Source:   user.Namespace,
			Relation: rebac.Relation(pid),
			Target:   session.Namespace,
		})
	}

	// Make sessions browsable in the general ReBAC editor. The default mapper uses Session.String() for a
	// recognizable instance label (title / first-message preview).
	rdb.RegisterResources(rebac.NewRepositoryResources(StrResSessions, StrResSessDesc, repoSessions))

	deleteHooks := &sessionDeleteHooks{}
	ucSession := session.NewUseCases(repoSessions, rdb, session.OnDelete(deleteHooks.run), session.OnDelete(func(id session.ID) error {
		return providerFiles.Release(file.Owner(id))
	}))

	// Ship the authorization for the assistant as one assignable role instead of a list of permissions an
	// operator has to reproduce by hand. Without this, the failure mode is silent: the chat button simply
	// does not render for anybody, which is exactly what happened in production before this existed.
	//
	// The wording is resolved against the system user (English) because it is written once at creation and
	// then belongs to the operator; see [application.Configurator.DeclareSystemRole].
	sys := cfg.SysUser()
	if err := cfg.DeclareSystemRole(role.Role{
		ID:          RoleAssistantUser,
		Name:        StrRoleAssistantUserName.Get(sys),
		Description: StrRoleAssistantUserDesc.Get(sys),
	}, uicompletion.RequiredPermissions()...); err != nil {
		return Management{}, fmt.Errorf("cannot declare the assistant user role: %w", err)
	}

	assistant := &Assistant{useCases: ucAI, sessions: ucSession}

	management = Management{
		UseCases:           ucAI,
		SessionUseCases:    ucSession,
		Assistant:          assistant,
		sessionDeleteHooks: deleteHooks,
	}

	// Make the assistant's model picker resolvable and keep it fresh. Both events that can invalidate the
	// cached list are subscribed to, because both are things an administrator does while wondering why the
	// list is empty: fixing the token, and saving the settings.
	if _, err := cfg.SettingsManagement(); err != nil {
		return Management{}, fmt.Errorf("cannot enable settings management for the assistant: %w", err)
	}

	events.SubscribeFor(cfg.EventBus(), func(evt settings.GlobalSettingsUpdated) {
		if _, ours := evt.Settings.(AssistantSettings); ours {
			assistant.Forget()
		}
	})

	events.SubscribeFor(cfg.EventBus(), func(secret.Updated) { assistant.Forget() })
	events.SubscribeFor(cfg.EventBus(), func(secret.Created) { assistant.Forget() })

	cfg.AddContextValue(core.ContextValue(SourceAssistantModels, form.NewQuerySource(
		func(subject user.Subject) ([]modelOption, error) {
			candidates, err := assistant.Providers(subject)
			if err != nil {
				// An empty picker with a log line beats an error banner on the settings page: the operator is
				// most likely on their way to configure the very provider that is missing.
				assistant.complain("models", fmt.Sprintf("the assistant model picker stays empty: %v", err))
				return nil, nil
			}

			var options []modelOption
			for _, c := range candidates {
				models, err := assistant.Models(subject, c)
				if err != nil {
					// one broken token must not hide the models of the other providers
					assistant.complain("models:"+string(c.Provider.Identity()), fmt.Sprintf("the assistant model picker misses a provider: %v", err))
					continue
				}

				for _, m := range models {
					options = append(options, modelOption{provider: c.Provider, model: m, qualified: len(candidates) > 1})
				}
			}

			return options, nil
		},
		func(o modelOption) string { return string(NewModelChoice(o.provider.Identity(), o.model.ID)) },
		func(o modelOption) string { return o.label() },
	)))
	cfg.AddContextValue(core.ContextValue("nago.ai", management))

	cfg.AddContextValue(core.ContextValue("", management.UseCases.FindProviderByID))
	cfg.AddContextValue(core.ContextValue("", management.UseCases.FindProviderByName))

	slog.Info("installed AI module")
	return management, nil
}

// modelOption is an entry of the model picker of [AssistantSettings].
type modelOption struct {
	provider provider.Provider
	model    model.Model
	// qualified prefixes the provider name, which is only helpful with several providers
	qualified bool
}

func (o modelOption) label() string {
	name := o.model.Name
	if name == "" {
		name = string(o.model.ID)
	}

	if o.qualified {
		return o.provider.Name() + " · " + name
	}

	return name
}
