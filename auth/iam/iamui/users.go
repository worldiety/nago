package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func Users(wnd core.Window, service *iam.Service) core.View {
	subject := wnd.Subject()

	bnd := crud.NewBinding[iam.User](wnd)
	bnd.Add(
		crud.Text("ID", func(e *iam.User) *auth.UID {
			return &e.ID // TODO 	crud.Update:   crud.ReadOnly,
		}).WithoutTable(),
		crud.Text("Vorname", func(e *iam.User) *string {
			return &e.Firstname
		}),
		crud.Text("Nachname", func(e *iam.User) *string {
			return &e.Lastname
		}),
		// TODO oneToMany is missing
		//crud.OneToMany(bnd, service.AllPermissions(subject), func(permission iam.Permission) string {
		//				return permission.Name()
		//			}, crud.Field[iam.User, []iam.PID]{
		//				Caption: "Berechtigungen",
		//				Stringer: func(user iam.User) string {
		//					// this is by intention: we show the entire inherited list of all permission, not just the customized ones
		//					perms, err := service.FindAllUserPermissions(subject, user.ID)
		//					if err != nil {
		//						slog.Error("failed to find all user permissions", "err", err, "uid", user.ID)
		//					}
		//
		//					actualPerms := service.AllPermissionsByIDs(subject, perms...)
		//					var tmp []string
		//					actualPerms(func(permission iam.Permission, e error) bool {
		//						if e != nil {
		//							err = e
		//							return false
		//						}
		//
		//						tmp = append(tmp, permission.Name())
		//						return true
		//					})
		//
		//					return strings.Join(tmp, ", ")
		//				},
		//				FromModel: func(user iam.User) []iam.PID {
		//					return user.Permissions // this is by intention, only allow editing the custom ones
		//				},
		//				IntoModel: func(model iam.User, value []iam.PID) (iam.User, error) {
		//					model.Permissions = value // this is by intention, only allow editing the custom ones
		//					return model, nil
		//				},
		//			})
		//
		//			crud.OneToMany(bnd,
		//				service.AllGroups(subject),
		//				func(usr iam.Group) string {
		//					return usr.Name
		//				},
		//				crud.FromPtr("Gruppen", func(model *iam.User) *[]auth.GID {
		//					return &model.Groups
		//				}),
		//			)
		//
		//			crud.OneToMany(bnd,
		//				service.AllRoles(subject),
		//				func(usr iam.Role) string {
		//					return usr.Name
		//				},
		//				crud.FromPtr("Rollen", func(model *iam.User) *[]auth.RID {
		//					return &model.Roles
		//				}),
		//			)
		crud.AggregateActions(
			"Optionen",
			crud.ButtonDelete(wnd, func(group iam.User) error {
				return service.DeleteUser(subject, group.ID)
			}),
			crud.ButtonEdit(bnd, func(user iam.User) (errorText string, infrastructureError error) {
				return "", service.UpdateUser(subject, user.ID, string(user.Email), user.Firstname, user.Lastname, user.Permissions, user.Roles, user.Groups)
			}),
		),
	)

	bndCrUsr := crud.NewBinding[createUser](wnd).Add(
		crud.Text("Vorname", func(e *createUser) *string {
			return &e.Firstname
		}),
		crud.Text("Nachname", func(e *createUser) *string {
			return &e.Lastname
		}),
		crud.Text("eMail", func(e *createUser) *string {
			return &e.Lastname
		}),
		crud.Text("Kennwort", func(e *createUser) *string {
			return &e.Lastname
		}),
		crud.Text("Kennwort wiederholen", func(e *createUser) *string {
			return &e.Lastname
		}),
	)
	opts := crud.Options(bnd).
		Actions(crud.ButtonCreate[createUser](bndCrUsr, createUser{}, func(model createUser) (errorText string, infrastructureError error) {
			if !iam.Email(model.EMail).Valid() {
				return "Die eMail-Adresse ist ungültig.", nil
			}

			if model.Password1 != model.Password2 {
				return "Die Kennwörter stimmen nicht überein.", nil
			}

			_, err := service.NewUser(subject, model.EMail, model.Firstname, model.Lastname, model.Password1)
			return "", err
		})).Title("Nutzerkonten").
		FindAll(service.AllUsers(subject))

	return crud.View[iam.User](opts)

}

type createUser struct {
	Firstname string
	Lastname  string
	EMail     string
	Password1 string
	Password2 string
}

func (createUser) Identity() string {
	return ""
}
