package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
)

func Groups(subject auth.Subject, modals ui.ModalOwner, service *iam.Service) core.Component {

	return crud.NewView(modals, crud.NewOptions[iam.Group](func(opts *crud.Options[iam.Group]) {
		opts.Title = "Gruppen"
		opts.FindAll = service.AllGroups(subject)
		opts.Create = func(group iam.Group) error {
			return service.CreateGroup(subject, group)
		}
		opts.OnDelete(func(group iam.Group) error {
			return service.DeleteGroup(subject, group.ID)
		})
		opts.OnUpdate(func(group iam.Group) error {
			return service.UpdateGroup(subject, group)
		})
		opts.Binding = crud.NewBinding[iam.Group](func(bnd *crud.Binding[iam.Group]) {

			crud.Text(bnd, crud.FromPtr("ID", func(model *iam.Group) *auth.GID {
				return &model.ID
			}, crud.RenderHints{
				crud.Overview: crud.Hidden,
				crud.Update:   crud.ReadOnly,
				crud.Create:   crud.Visible,
			}))

			crud.Text(bnd, crud.FromPtr("Name", func(model *iam.Group) *string {
				return &model.Name
			}))
			crud.Text(bnd, crud.FromPtr("Beschreibung", func(model *iam.Group) *string {
				return &model.Description
			}))
		})
	}))
}
