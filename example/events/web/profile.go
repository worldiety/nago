package web

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/presentation/ui"
)

type ProfileModel struct {
	UserID string
	Name   string
	Email  string
}

func Profile() ui.PageHandler {
	return ui.Page(
		"profile",
		RenderProfile,
		ui.OnAuthRequest(func(user auth.User, model ProfileModel) ProfileModel {
			model.Name = user.Name()
			model.UserID = user.UserID()
			model.Email = user.Email()
			return model
		}),
	)
}

func RenderProfile(model ProfileModel) ui.View {
	return ui.Grid{
		Columns: 2,
		Cells: slice.Of(
			ui.GridCell{
				Child: ui.Text("Your Profile"),
			},
			ui.GridCell{Child: ui.Text("Name")},
			ui.GridCell{Child: ui.Text(model.Name)},
			ui.GridCell{Child: ui.Text("Email")},
			ui.GridCell{Child: ui.Text(model.Email)},
			ui.GridCell{Child: ui.Text("ID")},
			ui.GridCell{Child: ui.Text(model.UserID)},
		),
	}
}
