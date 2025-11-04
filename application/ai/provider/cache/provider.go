// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
)

var _ provider.Provider = (*Provider)(nil)

// Provider introduces a local cache layer to wrap around the given provider. This isolates the given provider
// from a lot of queries and may decrease latencies and failure conditions, especially if the wrapped provider
// is actually a pure cloud provider, like [mistralai.NewProvider].
type Provider struct {
	prov                 provider.Provider
	repoModels           model.Repository
	repoLibraries        library.Repository
	repoAgents           agent.Repository
	repoDocuments        document.Repository
	repoConversations    conversation.Repository
	repoMessages         message.Repository
	docTextStore         blob.Store
	idxMsg               *data.CompositeIndex[conversation.ID, message.ID]
	idxProvModels        *data.CompositeIndex[provider.ID, model.ID]
	idxProvAgents        *data.CompositeIndex[provider.ID, agent.ID]
	idxProvLibraries     *data.CompositeIndex[provider.ID, library.ID]
	idxProvConversations *data.CompositeIndex[provider.ID, conversation.ID]
}

func NewProvider(
	other provider.Provider,
	repoModels model.Repository,
	repoLibraries library.Repository,
	repoAgents agent.Repository,
	repoDocuments document.Repository,
	repoConversations conversation.Repository,
	repoMessages message.Repository,
	docTextStore blob.Store,
	idxMsg *data.CompositeIndex[conversation.ID, message.ID],
	idxProvModels *data.CompositeIndex[provider.ID, model.ID],
	idxProvAgents *data.CompositeIndex[provider.ID, agent.ID],
	idxProvLibraries *data.CompositeIndex[provider.ID, library.ID],
	idxProvConversations *data.CompositeIndex[provider.ID, conversation.ID],
) *Provider {
	p := &Provider{
		prov:                 other,
		repoModels:           repoModels,
		repoLibraries:        repoLibraries,
		repoAgents:           repoAgents,
		repoDocuments:        repoDocuments,
		repoConversations:    repoConversations,
		repoMessages:         repoMessages,
		docTextStore:         docTextStore,
		idxMsg:               idxMsg,
		idxProvModels:        idxProvModels,
		idxProvAgents:        idxProvAgents,
		idxProvLibraries:     idxProvLibraries,
		idxProvConversations: idxProvConversations,
	}

	return p
}

// LoadAll downloads blindly all data into the given repositories. This may keep stale data and
// overwrite existing data. See also [Provider.Clear].
func (p *Provider) LoadAll() error {
	if p.prov.Conversations().IsSome() {
		for conv, err := range p.prov.Conversations().Unwrap().All(user.SU()) {
			if err != nil {
				return err
			}

			if err := p.repoConversations.Save(conv); err != nil {
				return err
			}

			if err := p.idxProvConversations.Put(p.Identity(), conv.ID); err != nil {
				return err
			}

			for msg, err := range p.prov.Conversations().Unwrap().Conversation(user.SU(), conv.ID).All(user.SU()) {
				if err != nil {
					return err
				}

				if err := p.repoMessages.Save(msg); err != nil {
					return err
				}

				if err := p.idxMsg.Put(conv.ID, msg.ID); err != nil {
					return err
				}

			}
		}

	}

	if p.prov.Agents().IsSome() {
		for ag, err := range p.prov.Agents().Unwrap().All(user.SU()) {
			if err != nil {
				return err
			}

			if err := p.repoAgents.Save(ag); err != nil {
				return err
			}

			if err := p.idxProvAgents.Put(p.Identity(), ag.ID); err != nil {
				return err
			}
		}

	}

	if p.prov.Libraries().IsSome() {
		for lib, err := range p.prov.Libraries().Unwrap().All(user.SU()) {
			if err != nil {
				return err
			}

			if err := p.repoLibraries.Save(lib); err != nil {
				return err
			}

			if err := p.idxProvLibraries.Put(p.Identity(), lib.ID); err != nil {
				return err
			}

			slog.Info("ai provider cache iterating library", "library", lib.Identity(), "provider", p.prov.Identity())
			for doc, err := range p.prov.Libraries().Unwrap().Library(lib.ID).All(user.SU()) {
				if err != nil {
					return err
				}

				if err := p.repoDocuments.Save(doc); err != nil {
					return err
				}
			}
		}

	}

	return nil
}

// Clear purges all caches from the used cache repositories. If the same repositories are used for different
// providers, these caches are purged as well.
func (p *Provider) Clear() error {
	if err := p.repoAgents.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear agents cache repository: %w", err)
	}

	if err := p.repoModels.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear models cache repository: %w", err)
	}

	if err := p.repoLibraries.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear libraries cache repository: %w", err)
	}

	if err := p.repoDocuments.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear documents cache repository: %w", err)
	}

	if err := p.repoConversations.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear conversations cache repository: %w", err)
	}

	if err := p.repoMessages.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear messages cache repository: %w", err)
	}

	if err := p.idxMsg.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear index message cache repository: %w", err)
	}

	if err := p.idxProvModels.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear index prov/model cache repository: %w", err)
	}

	if err := p.idxProvAgents.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear index prov/agent cache repository: %w", err)
	}

	if err := p.idxProvLibraries.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear index prov/libs cache repository: %w", err)
	}

	if err := p.idxProvConversations.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear index prov/conv cache repository: %w", err)
	}

	if err := blob.DeleteAll(p.docTextStore); err != nil {
		return fmt.Errorf("failed to clear documents text cache repository: %w", err)
	}

	return nil
}

func (p *Provider) Identity() provider.ID {
	return p.prov.Identity()
}

func (p *Provider) Name() string {
	return p.prov.Name()
}

func (p *Provider) Description() string {
	return p.prov.Description()
}

func (p *Provider) Models() provider.Models {
	return &cacheModels{parent: p}
}

func (p *Provider) Libraries() option.Opt[provider.Libraries] {
	if p.prov.Libraries().IsNone() {
		return option.None[provider.Libraries]()
	}

	return option.Some[provider.Libraries](&cacheLibraries{p})
}

func (p *Provider) Agents() option.Opt[provider.Agents] {
	if p.prov.Agents().IsNone() {
		return option.None[provider.Agents]()
	}

	return option.Some[provider.Agents](&cacheAgents{p})
}

func (p *Provider) Conversations() option.Opt[provider.Conversations] {
	if p.prov.Conversations().IsNone() {
		return option.None[provider.Conversations]()
	}

	return option.Some[provider.Conversations](&cacheConversations{p})
}
