package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xtable"
)

type UserView struct {
	ID             string
	Mail           string
	Firstname      string
	Lastname       string
	Groups         []auth.GID
	Roles          []auth.RID
	AllPermissions []string
}

func Users(subject auth.Subject, modals ui.ModalOwner, service *iam.Service) core.Component {
	opts := xtable.Options[UserView]{
		CanSearch: true,
	}

	if subject.HasPermission(iam.DeleteUser) {
		opts.AggregateActions = append(opts.AggregateActions, xtable.NewDeleteAction(func(t UserView) error {
			return service.DeleteUser(subject, auth.UID(t.ID))
		}))
	}

	if subject.HasPermission(iam.CreateUser) {
		opts.Actions = append(opts.Actions, ui.NewActionButton("Neuen Nutzer anlegen", func() {
			create(subject, modals, service)
		}))
	}

	return xtable.NewTable[UserView](modals,
		mapUser2view(service.AllUsers(subject)),
		xtable.NewBinding[UserView]().
			AddColumn(xtable.Column[UserView]{
				Caption:  "Vorname",
				Sortable: true,
				MapField: func(view UserView) string {
					return view.Firstname
				},
			}).
			AddColumn(xtable.Column[UserView]{
				Caption:  "Nachname",
				Sortable: true,
				MapField: func(view UserView) string {
					return view.Lastname
				},
			}).
			AddColumn(xtable.Column[UserView]{
				Caption:  "eMail",
				Sortable: true,
				MapField: func(view UserView) string {
					return view.Mail
				},
			}),
		opts,
	)
}

func mapUser2view(src iter.Seq2[iam.User, error]) iter.Seq2[UserView, error] {
	return iter.Map2(func(usr iam.User, err error) (UserView, error) {
		if err != nil {
			return UserView{}, err
		}

		usrv := UserView{
			ID:        string(usr.ID),
			Firstname: usr.Firstname,
			Lastname:  usr.Lastname,
			Mail:      string(usr.Email),
		}

		return usrv, nil
	}, src)
}
