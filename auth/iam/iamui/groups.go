package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func Groups(wnd core.Window, service *iam.Service) core.View {
	subject := wnd.Subject()

	bnd := crud.NewBinding[iam.Group](wnd)
	bnd.Add(
		crud.Text("ID", func(e *iam.Group) *auth.GID {
			return &e.ID
		}).ReadOnly(true).WithoutTable(),
		crud.Text("Name", func(e *iam.Group) *string {
			return &e.Name
		}),
		crud.Text("Beschreibung", func(e *iam.Group) *string {
			return &e.Description
		}),
		crud.AggregateActions(
			"Optionen",
			crud.ButtonDelete(wnd, func(group iam.Group) error {
				return service.DeleteGroup(subject, group.ID)
			}),
			crud.ButtonEdit(bnd, func(group iam.Group) (errorText string, infrastructureError error) {
				return "", service.UpdateGroup(subject, group)
			}),
		),
	)

	createBnd := bnd.Inherit("create")
	createBnd.SetDisabledByLabel("ID", false)

	opts := crud.Options(bnd).
		Actions(crud.ButtonCreate(createBnd, iam.Group{}, func(group iam.Group) (errorText string, infrastructureError error) {
			return "", service.CreateGroup(subject, group)
		})).Title("Gruppen").
		FindAll(service.AllGroups(subject))

	return crud.View[iam.Group](opts)
}
