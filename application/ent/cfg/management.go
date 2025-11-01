package cfgent

import (
	"fmt"
	"strings"
	"sync"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/ent"
	uientities "go.wdy.de/nago/application/ent/ui"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
)

type Module[T ent.Aggregate[T, ID], ID data.IDType] struct {
	Mutex      *sync.Mutex // Mutex used by the UseCases to protect critical write sections
	Repository data.Repository[T, ID]
	UseCases   ent.UseCases[T, ID]
	Pages      uientities.Pages
}

type Options[T ent.Aggregate[T, ID], ID ~string] struct {
	// Mutex to protect the default critical sections. If nil, a new mutex is allocated as required.
	// If you don't know or care, just leave it nil.
	Mutex *sync.Mutex

	// AdminCenter indicates if the admin center shall show a card to view and manage entities.
	// The visibility and capabilities are automatically determined by the assigned permissions.
	AdminCenter bool

	// List may be nil to generate a default list view implementation, e.g. using the [dataview.FromData].
	// The page route can be read from [Module.Pages.Create]
	List func(wnd core.Window, uc ent.UseCases[T, ID]) core.View

	// Create may be nil to generate a default create form view, e.g. using [form.Auto]. The page route can be
	// read from [Module.Pages.Create]. Note, that your use case may provide validation information by returning
	// a custom error [xerrors.WithFields].
	Create func(wnd core.Window, uc ent.UseCases[T, ID]) core.View

	// Update may be nil to generate a default update form view, e.g. using [form.Auto]. The page route can be
	//	// read from [Module.Pages.Update]. Note, that your use case may provide validation information by returning
	//	// a custom error [xerrors.WithFields].
	Update func(wnd core.Window, uc ent.UseCases[T, ID], id ID) core.View

	// DecorateUseCases is invoked before the use cases are passed into all generated and dependent code fragments
	// thus you can customize, intercept or replace any standard use case here. For example, you can
	// apply custom validation and return [xerrors.WithFields].
	DecorateUseCases func(uc ent.UseCases[T, ID]) ent.UseCases[T, ID]
}

// Enable configures a crud module instance. See also [crud.UseCases] and [crud.DeclarePermissions] for details.
func Enable[T ent.Aggregate[T, ID], ID ~string](cfg *application.Configurator, prefix permission.ID, entityName string, opts Options[T, ID]) (Module[T, ID], error) {
	mod, ok := core.FromContext[Module[T, ID]](cfg.Context(), "")
	if ok {
		return mod, nil
	}

	if !prefix.Valid() {
		return Module[T, ID]{}, fmt.Errorf("prefix is not valid")
	}

	bucketName := string(prefix)

	store, err := cfg.EntityStore(bucketName)
	if err != nil {
		return mod, fmt.Errorf("failed to open entity store: %w", err)
	}

	if opts.Mutex == nil {
		opts.Mutex = &sync.Mutex{}
	}

	repo := json.NewSloppyJSONRepository[T, ID](store)
	perms := ent.DeclarePermissions[T, ID](prefix, entityName)
	uc := ent.NewUseCases[T, ID](perms, repo, ent.Options{
		Mutex: opts.Mutex,
	})

	if opts.DecorateUseCases != nil {
		uc = opts.DecorateUseCases(uc)
	}

	mod = Module[T, ID]{
		Repository: repo,
		UseCases:   uc,
		Pages: uientities.Pages{
			List:   "admin/entities/" + makeFactoryID(prefix) + "/list",
			Create: "admin/entities/" + makeFactoryID(prefix) + "/create",
			Update: "admin/entities/" + makeFactoryID(prefix) + "/update",
		},
	}

	cfg.RootViewWithDecoration(mod.Pages.List, func(wnd core.Window) core.View {
		return uientities.PageList(wnd, mod.UseCases, uientities.PageListOptions[T, ID]{
			EntityName: entityName,
			Perms:      perms,
			Pages:      mod.Pages,
			List:       opts.List,
			Prefix:     prefix,
		})
	})

	cfg.RootViewWithDecoration(mod.Pages.Create, func(wnd core.Window) core.View {
		return uientities.PageCreate(wnd, mod.UseCases, uientities.PageCreateOptions[T, ID]{
			EntityName: entityName,
			Perms:      perms,
			Pages:      mod.Pages,
			Prefix:     prefix,
			Create:     opts.Create,
		})
	})

	cfg.RootViewWithDecoration(mod.Pages.Update, func(wnd core.Window) core.View {
		return uientities.PageUpdate(wnd, mod.UseCases, uientities.PageUpdateOptions[T, ID]{
			EntityName: entityName,
			Perms:      perms,
			Pages:      mod.Pages,
			Prefix:     prefix,
			Update:     opts.Update,
		})
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var res admin.Group
		if !opts.AdminCenter {
			return res
		}

		if !(subject.HasPermission(perms.FindAll) || subject.HasPermission(perms.FindAllIdentifiers)) {
			return res
		}

		res.Title = entityName
		res.Entries = append(res.Entries, admin.Card{
			Title:  entityName,
			Text:   "",
			Target: mod.Pages.List,
		})

		return res
	})

	cfg.AddContextValue(core.ContextValue(string("module-"+prefix), mod))
	cfg.AddContextValue(core.ContextValue(string(prefix), form.AnyUseCaseList[T, ID](uc.FindAll)))

	return mod, nil
}

func makeFactoryID(prefix permission.ID) core.NavigationPath {
	return core.NavigationPath(strings.ReplaceAll(string(prefix), ".", "-"))
}
