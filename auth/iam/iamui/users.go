package iamui

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xtable"
	"strings"
	"time"
)

type UserView struct {
	ID             string
	Mail           string
	Firstname      string
	Lastname       string
	Groups         []auth.GID
	Roles          []auth.RID
	AllPermissions []string
	Status         iam.AccountStatus
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

	if subject.HasPermission(iam.ReadUser) {
		opts.AggregateActions = append(opts.AggregateActions, xtable.NewEditAction(func(t UserView) error {
			editUser(subject, modals, auth.UID(t.ID), service)
			return nil
		}))
	}

	if subject.HasPermission(iam.CreateUser) {
		opts.Actions = append(opts.Actions, ui.NewButton(func(btn *ui.Button) {
			btn.Caption().Set("Neuen Nutzer anlegen")
			btn.PreIcon().Set(icon.UserPlus)
			btn.Action().Set(func() {
				create(subject, modals, service)
			})
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
				CompareField: func(a, b UserView) int {
					return strings.Compare(a.Firstname, b.Firstname)
				},
			}).
			AddColumn(xtable.Column[UserView]{
				Caption:  "Nachname",
				Sortable: true,
				MapField: func(view UserView) string {
					return view.Lastname
				},
				CompareField: func(a, b UserView) int {
					return strings.Compare(a.Lastname, b.Lastname)
				},
			}).
			AddColumn(xtable.Column[UserView]{
				Caption:  "eMail",
				Sortable: true,
				MapField: func(view UserView) string {
					return view.Mail
				},
				CompareField: func(a, b UserView) int {
					return strings.Compare(a.Mail, b.Mail)
				},
			}).
			AddColumn(xtable.Column[UserView]{
				Caption: "Status",
				MapField: func(view UserView) string {
					return iam.MatchAccountStatus(view.Status,
						func(enabled iam.Enabled) string {
							return "Zulässig"
						},
						func(disabled iam.Disabled) string {
							return "Blockiert"
						},
						func(until iam.EnabledUntil) string {
							return fmt.Sprintf("Zulässig bis %s", until.ValidUntil.Format(time.DateTime))
						},
						func(a any) string {
							return "unbekannt"
						},
					)
				},
				CompareField: func(a, b UserView) int {
					return strings.Compare(a.Mail, b.Mail)
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
			Status:    usr.Status,
		}

		return usrv, nil
	}, src)
}
