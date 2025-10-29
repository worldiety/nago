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
	"iter"
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
	"go.wdy.de/nago/pkg/data"
)

var _ provider.Provider = (*Provider)(nil)

// Provider introduces a local cache layer to wrap around the given provider. This isolates the given provider
// from a lot of queries and may decrease latencies and failure conditions, especially if the wrapped provider
// is actually a pure cloud provider, like [mistralai.NewProvider].
type Provider struct {
	prov              provider.Provider
	repoModels        model.Repository
	repoLibraries     library.Repository
	repoAgents        agent.Repository
	repoDocuments     document.Repository
	repoConversations conversation.Repository
	repoMessages      message.Repository
	idxMsg            *data.CompositeIndex[conversation.ID, message.ID]
}

func NewProvider(
	other provider.Provider,
	repoModels model.Repository,
	repoLibraries library.Repository,
	repoAgents agent.Repository,
	repoDocuments document.Repository,
	repoConversations conversation.Repository,
	repoMessages message.Repository,
	idxMsg *data.CompositeIndex[conversation.ID, message.ID],
) *Provider {
	p := &Provider{
		prov:              other,
		repoModels:        repoModels,
		repoLibraries:     repoLibraries,
		repoAgents:        repoAgents,
		repoDocuments:     repoDocuments,
		repoConversations: repoConversations,
		repoMessages:      repoMessages,
		idxMsg:            idxMsg,
	}

	return p
}

// LoadIfEmpty checks for each repository, if it is empty and tries to load from the wrapped provider and
// stores the result in the cache. If at least one entry is found within a cache repository, the remote data
// will not be loaded again. See also [Provider.Clear].
func (p *Provider) LoadIfEmpty() error {
	if err := loadIfEmpty(p.Identity(), p.repoModels, func() iter.Seq2[model.Model, error] {
		return p.prov.Models().All(user.SU())
	}); err != nil {
		return err
	}

	if p.prov.Conversations().IsSome() {
		if err := p.idxMsg.Clear(context.Background()); err != nil {
			return err
		}

		if err := p.repoMessages.DeleteAll(); err != nil {
			return err
		}

		if err := loadIfEmpty(p.Identity(), p.repoConversations, func() iter.Seq2[conversation.Conversation, error] {
			return p.prov.Conversations().Unwrap().All(user.SU())
		}); err != nil {
			return err
		}

		for conv, err := range p.repoConversations.All() {
			if err != nil {
				return err
			}

			slog.Info("ai provider cache iterating conversation", "conversation", conv.Identity(), "provider", p.prov.Identity())
			lib := p.prov.Conversations().Unwrap().Conversation(user.SU(), conv.ID)
			for m, err := range lib.All(user.SU()) {
				if err != nil {
					return err
				}

				if err := p.repoMessages.Save(m); err != nil {
					return err
				}

				if err := p.idxMsg.Put(conv.ID, m.ID); err != nil {
					return err
				}

			}

		}
	}

	if p.prov.Agents().IsSome() {
		if err := loadIfEmpty(p.Identity(), p.repoAgents, func() iter.Seq2[agent.Agent, error] {
			return p.prov.Agents().Unwrap().All(user.SU())
		}); err != nil {
			return err
		}
	}

	if p.prov.Libraries().IsSome() {
		if err := loadIfEmpty(p.Identity(), p.repoLibraries, func() iter.Seq2[library.Library, error] {
			return p.prov.Libraries().Unwrap().All(user.SU())
		}); err != nil {
			return err
		}

		for cLib, err := range p.repoLibraries.All() {
			if err != nil {
				return err
			}

			slog.Info("ai provider cache iterating library", "library", cLib.Identity(), "provider", p.prov.Identity())
			lib := p.prov.Libraries().Unwrap().Library(cLib.ID)
			if err := loadIfEmpty(p.Identity(), p.repoDocuments, func() iter.Seq2[document.Document, error] {
				return lib.All(user.SU())
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadIfEmpty[E data.Aggregate[ID], ID data.IDType](prov provider.ID, repo data.Repository[E, ID], src func() iter.Seq2[E, error]) (err error) {
	defer func() {
		if err != nil {
			slog.Error("ai provider cache failed to load from source: removing entries from cache", "repository", repo.Name(), "provider", prov)
			if err := repo.DeleteAll(); err != nil {
				slog.Error("failed to delete all entries from cache", "repository", repo.Name(), "provider", prov, "err", err.Error())
			}
		}
	}()

	count, err := repo.Count()
	if err != nil {
		return fmt.Errorf("failed to count repository %s: %w", repo.Name(), err)
	}

	if count == 0 {
		slog.Info("filling ai provider cache from remote...", "repository", repo.Name(), "provider", prov)

		num := 0
		for e, err := range src() {
			if err != nil {
				return err
			}

			if err := repo.Save(e); err != nil {
				return err
			}
			num++
		}

		slog.Info("ai provider cache loaded from source", "repository", repo.Name(), "count", num, "provider", prov)
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
