package uigroup

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
	"iter"
)

type Pages struct {
	Groups core.NavigationPath
}

func Groups(wnd core.Window, useCases group.UseCases) core.View {
	uc := crud.UseCasesFromFuncs[group.Group, group.ID](
		func(subject auth.Subject, id group.ID) (std.Option[group.Group], error) {
			return useCases.FindByID(subject, id)
		},
		func(subject auth.Subject) iter.Seq2[group.Group, error] {
			return useCases.FindAll(subject)
		},
		func(subject auth.Subject, id group.ID) error {
			return useCases.Delete(subject, id)
		},
		func(subject auth.Subject, entity group.Group) (group.ID, error) {
			return useCases.Upsert(subject, entity)
		},
	)
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Gruppen"}, uc)(wnd)
}
