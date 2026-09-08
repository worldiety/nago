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

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/agent"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/libsync"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/provider/cache"
	"go.wdy.de/nago/application/ai/rest"
	"go.wdy.de/nago/application/ai/session"
	uiai "go.wdy.de/nago/application/ai/ui"
	cfgdrive "go.wdy.de/nago/application/drive/cfg"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/layout"
	"golang.org/x/text/language"
)

var (
	StrMaintenanceAdminCardDesc = i18n.MustString("nago.ai.admin.maintenance_desc", i18n.Values{language.English: "Apply some maintenance tasks to the AI subsystem.", language.German: "Wartungsarbeiten am KI Subsystem durchführen."})

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
	LibSyncUseCases libsync.UseCases
	SessionUseCases session.UseCases
	Pages           uiai.Pages

	// Assistant is the ready-to-use chat assistant: provider lookup, model resolution, operator settings and
	// the floating button. See [Assistant.Decorate].
	Assistant *Assistant
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

	cacheEnabled := true
	repoAgents, err := application.JSONRepository[agent.Agent](cfg, "nago.ai.cache.agent")
	if err != nil {
		return Management{}, err
	}

	repoConversations, err := application.JSONRepository[conversation.Conversation](cfg, "nago.ai.cache.conversation")
	if err != nil {
		return Management{}, err
	}

	repoMessages, err := application.JSONRepository[message.Message](cfg, "nago.ai.cache.message")
	if err != nil {
		return Management{}, err
	}

	repoLibraries, err := application.JSONRepository[library.Library](cfg, "nago.ai.cache.library")
	if err != nil {
		return Management{}, err
	}

	repoDocuments, err := application.JSONRepository[document.Document](cfg, "nago.ai.cache.document")
	if err != nil {
		return Management{}, err
	}

	repoModels, err := application.JSONRepository[model.Model](cfg, "nago.ai.cache.model")
	if err != nil {
		return Management{}, err
	}

	repoFiles, err := application.JSONRepository[file.File](cfg, "nago.ai.cache.file")
	if err != nil {
		return Management{}, err
	}

	blobTextStore, err := cfg.FileStore("nago.ai.cache.document_text")
	if err != nil {
		return Management{}, err
	}

	fileStore, err := cfg.FileStore("nago.ai.cache.file_data")
	if err != nil {
		return Management{}, err
	}

	idxConvStore, err := cfg.EntityStore("nago.ai.cache.idx_conversation_message")
	if err != nil {
		return Management{}, err
	}
	idxConvMsg := data.NewCompositeIndex[conversation.ID, message.ID](idxConvStore)

	idxProvModStore, err := cfg.EntityStore("nago.ai.cache.idx_provider_model")
	if err != nil {
		return Management{}, err
	}
	idxProvMod := data.NewCompositeIndex[provider.ID, model.ID](idxProvModStore)

	idxProvAgentsStore, err := cfg.EntityStore("nago.ai.cache.idx_provider_agent")
	if err != nil {
		return Management{}, err
	}
	idxProvAgents := data.NewCompositeIndex[provider.ID, agent.ID](idxProvAgentsStore)

	idxProvLibrariesStore, err := cfg.EntityStore("nago.ai.cache.idx_provider_library")
	if err != nil {
		return Management{}, err
	}
	idxProvLibraries := data.NewCompositeIndex[provider.ID, library.ID](idxProvLibrariesStore)

	idxProvConvStore, err := cfg.EntityStore("nago.ai.cache.idx_provider_conversation")
	if err != nil {
		return Management{}, err
	}
	idxProvConv := data.NewCompositeIndex[provider.ID, conversation.ID](idxProvConvStore)

	idxProvFileStore, err := cfg.EntityStore("nago.ai.cache.idx_provider_file")
	if err != nil {
		return Management{}, err
	}
	idxProvFile := data.NewCompositeIndex[provider.ID, file.ID](idxProvFileStore)

	ucAI := ai.NewUseCases(cfg.EventBus(), secrets.UseCases.FindGroupSecrets, func(provider provider.Provider) (provider.Provider, error) {
		if !cacheEnabled {
			return provider, nil
		}

		prov := cache.NewProvider(
			provider,
			repoModels,
			repoLibraries,
			repoAgents,
			repoDocuments,
			repoConversations,
			repoMessages,
			repoFiles,
			blobTextStore,
			fileStore,
			idxConvMsg,
			idxProvMod,
			idxProvAgents,
			idxProvLibraries,
			idxProvConv,
			idxProvFile,
		)

		return prov, nil
	})

	modDrive, err := cfgdrive.Enable(cfg)
	if err != nil {
		return Management{}, err
	}

	stores, err := cfg.Stores()
	if err != nil {
		return Management{}, err
	}

	jobRepo, err := application.JSONRepository[libsync.Job](cfg, "nago.ai.libsync.job")
	syncRepo, err := application.JSONRepository[libsync.SyncInfo](cfg, "nago.ai.libsync.sync_info")

	ucLibSync := libsync.NewUseCases(
		cfg.EventBus(),
		ucAI.FindProviderByID,
		jobRepo,
		syncRepo,
		stores,
		modDrive.UseCases.WalkDir,
		modDrive.UseCases.Get,
		modDrive.UseCases.Stat,
	)

	// Sessions are provider-independent, locally persisted chats on top of the stateless completion API.
	// Unlike provider conversations they are not wrapped by the cache decorator - the whole (lossless)
	// history lives in this repository.
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

	ucSession := session.NewUseCases(repoSessions, rdb)

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
		LibSyncUseCases: ucLibSync,
		UseCases:        ucAI,
		SessionUseCases: ucSession,
		Assistant:       assistant,
		Pages: uiai.Pages{
			Maintenance:  "admin/ai/maintenance",
			Provider:     "admin/ai/provider",
			Library:      "admin/ai/library",
			Conversation: "admin/ai/provider/conversation",
			Document:     "admin/ai/library/document",
			Chat:         "admin/ai/chat",
			Agent:        "admin/ai/agent",
		},
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
		func(subject user.Subject) ([]model.Model, error) {
			_, comps, err := assistant.Provider(subject)
			if err != nil {
				// An empty picker with a log line beats an error banner on the settings page: the operator is
				// most likely on their way to configure the very provider that is missing.
				assistant.complain("models", fmt.Sprintf("the assistant model picker stays empty: %v", err))
				return nil, nil
			}

			models, err := assistant.Models(subject, comps)
			if err != nil {
				assistant.complain("models", fmt.Sprintf("the assistant model picker stays empty: %v", err))
				return nil, nil
			}

			return models, nil
		},
		func(m model.Model) string { return string(m.ID) },
		func(m model.Model) string { return m.Name },
	)))

	cfg.RootViewWithDecoration(management.Pages.Provider, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageProvider(wnd, management.UseCases))
	})
	cfg.RootViewWithDecoration(management.Pages.Library, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageLibrary(wnd, stores, modDrive.UseCases.ReadDrives, modDrive.UseCases.Stat, management.UseCases, management.LibSyncUseCases))
	})

	cfg.RootViewWithDecoration(management.Pages.Conversation, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageConversation(wnd, management.UseCases))
	})

	cfg.NoFooter(management.Pages.Chat)
	cfg.RootViewWithDecoration(management.Pages.Chat, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageChat(wnd, management.UseCases))
	})

	cfg.RootViewWithDecoration(management.Pages.Maintenance, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageMaintenance(wnd, management.UseCases))
	})

	cfg.RootViewWithDecoration(management.Pages.Agent, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageAgent(wnd, management.UseCases))
	})

	cfg.RootViewWithDecoration(management.Pages.Document, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uiai.PageDocument(wnd, management.UseCases))
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {

		grp := admin.Group{
			Title: "AI",
		}

		grp.Entries = append(grp.Entries, admin.Card{
			Title:      rstring.LabelMaintenance.Get(subject),
			Text:       StrMaintenanceAdminCardDesc.Get(subject),
			Target:     management.Pages.Maintenance,
			Permission: ai.PermClearCache,
		})

		for provider, err := range ucAI.FindAllProvider(subject) {
			if err != nil {
				slog.Error("failed to find provider", "err", err.Error())
				continue
			}

			grp.Entries = append(grp.Entries, admin.Card{
				Title:        provider.Name(),
				Text:         provider.Description(),
				Target:       management.Pages.Provider,
				TargetParams: core.Values{"provider": string(provider.Identity())},
			})
		}

		return grp
	})

	cfg.AddContextValue(core.ContextValue("nago.ai", management))

	cfg.AddContextValue(core.ContextValue("", management.UseCases.FindProviderByID))
	cfg.AddContextValue(core.ContextValue("", management.UseCases.FindProviderByName))

	cfg.HandleFunc(rest.Endpoint, rest.NewFileEndpoint(ucAI.FindProviderByID))

	slog.Info("installed AI module")
	return management, nil
}
