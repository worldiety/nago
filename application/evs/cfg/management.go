// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgevs

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/evs"
	uievs "go.wdy.de/nago/application/evs/ui"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
)

type Module[Evt any] struct {
	Mutex       *sync.Mutex // Mutex used by the UseCases to protect critical write sections
	UseCases    evs.UseCases[Evt]
	Pages       uievs.Pages
	Permissions evs.Permissions
	Indexers    []evs.Indexer[Evt]
}

type AdminCenter struct {
	// Description is shown on the card. If empty, a default CRUD text is shown.
	Description string
}

type Options[Evt any] struct {
	// Mutex to protect the default critical sections. If nil, a new mutex is allocated as required.
	// If you don't know or care, just leave it nil.
	Mutex *sync.Mutex

	// Bus to be passed into all use cases. May be nil.
	Bus events.Bus

	// AdminCenter configuration for this entity type.
	AdminCenter AdminCenter

	// Schema maps discriminators to concrete types implementing Evt.
	Schema map[evs.Discriminator]reflect.Type

	// Indexer is evaluated on insertion into the event store and will return those strings which
	// are used as a composite key lookup.
	Indexer []evs.Indexer[Evt]

	// indexerFactories is an addition to the Indexer field which delays construction of the indexer
	// to avoid that the developer needs to declare the actual indexer implementation upfront.
	// This is just for convenience and the factories will append to the indexer slice.
	indexerFactories []func(cfg *application.Configurator, prefix permission.ID, entityName string) (evs.Indexer[Evt], error)

	// DecorateUseCases is invoked before the use cases are passed into all generated and dependent code fragments
	// thus you can customize, intercept or replace any standard use case here. For example, you can
	// apply custom validation and return [xerrors.WithFields].
	DecorateUseCases func(uc evs.UseCases[Evt]) evs.UseCases[Evt]
}

func (o Options[Evt]) WithOptions(opts ...Opt[Evt]) Options[Evt] {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

type Opt[Evt any] func(*Options[Evt])

func Schema[T any, Evt any](d evs.Discriminator) Opt[Evt] {
	return func(o *Options[Evt]) {
		if o.Schema == nil {
			o.Schema = make(map[evs.Discriminator]reflect.Type)
		}

		if _, ok := o.Schema[d]; ok {
			panic(fmt.Errorf("duplicate discriminator: %s", d))
		}

		o.Schema[d] = reflect.TypeFor[T]()
	}
}

func Index[Primary ~string, Evt any](reader evs.PrimaryReader[Primary, Evt]) Opt[Evt] {
	return func(o *Options[Evt]) {
		o.indexerFactories = append(o.indexerFactories, func(cfg *application.Configurator, prefix permission.ID, entityName string) (evs.Indexer[Evt], error) {
			name := strings.ToLower(reflect.TypeFor[Primary]().Name())
			store, err := cfg.EntityStore(string(prefix.WithName(name)) + ".idx")
			if err != nil {
				return nil, err
			}

			rType := reflect.TypeFor[Primary]()
			idx := evs.NewStoreIndex[Primary, Evt](store, reader)
			idx.SetInfo(evs.IndexerInfo{
				ID:          evs.IdxID(makeFactoryID(permission.ID(rType.String()))),
				Name:        rType.Name(),
				Description: uievs.StrIndexManagement.String(),
			})

			return idx, nil
		})
	}
}

// Enable configures an event sourcing module instance. See also [evs.UseCases] and [evs.DeclarePermissions] for details.
func Enable[Evt any](cfg *application.Configurator, prefix permission.ID, entityName string, opts Options[Evt]) (Module[Evt], error) {
	mod, ok := core.FromContext[Module[Evt]](cfg.Context(), "")
	if ok {
		return mod, nil
	}

	if !prefix.Valid() {
		return Module[Evt]{}, fmt.Errorf("prefix is not valid")
	}

	eventsBucketName := string(prefix) + ".event"
	eventStore, err := cfg.EntityStore(eventsBucketName)
	if err != nil {
		return mod, fmt.Errorf("failed to open entity store: %w", err)
	}

	timesBucketName := string(prefix) + ".time.idx"
	timesStore, err := cfg.EntityStore(timesBucketName)
	if err != nil {
		return mod, fmt.Errorf("failed to open entity store: %w", err)
	}

	if opts.Mutex == nil {
		opts.Mutex = &sync.Mutex{}
	}

	for _, factory := range opts.indexerFactories {
		fac, err := factory(cfg, prefix, entityName)
		if err != nil {
			return mod, err
		}

		opts.Indexer = append(opts.Indexer, fac)
	}

	perms := evs.DeclarePermissions[Evt](prefix, entityName)
	uc := evs.NewUseCases[Evt](perms, eventStore, timesStore, evs.Options[Evt]{
		Mutex:   opts.Mutex,
		Bus:     opts.Bus,
		Indexer: opts.Indexer,
	})

	for discriminator, r := range opts.Schema {
		if err := uc.Register(r, discriminator); err != nil {
			return mod, err
		}
	}

	mod = configureMod(cfg, perms, uc, opts)
	return mod, nil
}

func configureMod[Evt any](cfg *application.Configurator, perms evs.Permissions, uc evs.UseCases[Evt], opts Options[Evt]) Module[Evt] {
	if opts.DecorateUseCases != nil {
		uc = opts.DecorateUseCases(uc)
	}

	mod := Module[Evt]{

		UseCases: uc,
		Pages: uievs.Pages{
			Audit:  "admin/events/" + makeFactoryID(perms.Prefix) + "/audit",
			Create: "admin/events/" + makeFactoryID(perms.Prefix) + "/create",
			Index:  "admin/events/" + makeFactoryID(perms.Prefix) + "/index",
		},
		Permissions: perms,
	}

	cfg.RootViewWithDecoration(mod.Pages.Audit, func(wnd core.Window) core.View {
		return uievs.PageAudit(wnd, mod.UseCases, uievs.PageAuditOptions[Evt]{
			EntityName: perms.EntityName,
			Perms:      perms,
			Pages:      mod.Pages,
			Prefix:     perms.Prefix,
			Indexer:    opts.Indexer,
		})
	})

	for rT := range uc.RegisteredTypes() {
		cfg.RootViewWithDecoration(mod.Pages.Create.Join(string(rT.Discriminator)), func(wnd core.Window) core.View {
			return uievs.PageCreate(wnd, mod.UseCases, uievs.PageCreateOptions[Evt]{
				EntityName: perms.EntityName,
				Perms:      perms,
				Pages:      mod.Pages,
				Prefix:     perms.Prefix,
			})
		})
	}

	for _, idxer := range opts.Indexer {
		info := idxer.Info()
		cfg.RootViewWithDecoration(mod.Pages.Index.Join(string(info.ID)), func(wnd core.Window) core.View {
			return uievs.PageIndex(wnd, mod.UseCases, uievs.PageIndexOptions[Evt]{
				EntityName: perms.EntityName,
				Perms:      perms,
				Pages:      mod.Pages,
				Prefix:     perms.Prefix,
				Indexer:    opts.Indexer,
			})
		})
	}

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var res admin.Group

		if !(subject.HasPermission(perms.ReadAll)) {
			return res
		}

		cardText := subject.Bundle().Resolve(opts.AdminCenter.Description)
		if cardText == "" {
			cardText = uievs.StrManageEntitiesX.Get(subject, i18n.String("name", perms.EntityName))
		}

		groupTitle := perms.EntityName

		res.Title = groupTitle
		res.Entries = append(res.Entries, admin.Card{
			Title:  perms.EntityName,
			Text:   cardText,
			Target: mod.Pages.Audit,
			ID:     string(perms.Prefix),
		})

		for _, idxer := range opts.Indexer {
			info := idxer.Info()
			desc := subject.Bundle().Resolve(info.Description, i18n.String("name", info.Name))
			res.Entries = append(res.Entries, admin.Card{
				Title:  info.Name,
				Text:   desc,
				Target: mod.Pages.Index.Join(string(info.ID)),
				ID:     string(info.ID),
			})
		}

		return res
	})

	cfg.AddContextValue(core.ContextValue(string("module-"+perms.Prefix), mod))
	//cfg.AddContextValue(core.ContextValue(string(perms.Prefix), form.AnyUseCaseList[T, ID](uc.FindAll)))

	mod.Indexers = opts.Indexer
	return mod
}

func makeFactoryID(prefix permission.ID) core.NavigationPath {
	return core.NavigationPath(strings.ReplaceAll(string(prefix), ".", "-"))
}
