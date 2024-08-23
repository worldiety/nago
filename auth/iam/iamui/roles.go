package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
)

func Roles(wnd core.Window, service *iam.Service) core.View {

	subject := wnd.Subject()

	bnd := crud.NewBinding[iam.Role](wnd)
	bnd.Add(
		crud.Text("ID", func(e *iam.Role) *auth.RID {
			return &e.ID
		}).ReadOnly(true).WithoutTable(),
		crud.Text("Name", func(e *iam.Role) *string {
			return &e.Name
		}),
		crud.Text("Beschreibung", func(e *iam.Role) *string {
			return &e.Description
		}),
		crud.OneToMany("Berechtigungen", service.AllPermissions(subject), func(t iam.Permission) core.View {
			return ui.Text(t.Name())
		}, func(model *iam.Role) *[]iam.PID {
			return &model.Permissions
		}),
		crud.AggregateActions(
			"Optionen",
			crud.ButtonDelete(wnd, func(group iam.Role) error {
				return service.DeleteRole(subject, group.ID)
			}),
			crud.ButtonEdit(bnd, func(group iam.Role) (errorText string, infrastructureError error) {
				return "", service.UpdateRole(subject, group)
			}),
		),
	)

	createBnd := bnd.Inherit("create")
	createBnd.SetDisabledByLabel("ID", false)

	opts := crud.Options(bnd).
		Actions(crud.ButtonCreate(createBnd, iam.Role{}, func(group iam.Role) (errorText string, infrastructureError error) {
			return "", service.CreateRole(subject, group)
		})).Title("Rollen").
		FindAll(service.AllRoles(subject))

	return crud.View[iam.Role](opts)

}
