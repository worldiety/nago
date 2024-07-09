package application

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/auth/iam/iamui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
)

type IAMSettings struct {
	Decorator   func(wnd core.Window, page *ui.Page, content core.Component)
	Permissions Permissions
	Users       Users
	Sessions    Sessions
	Roles       Roles
	Groups      Groups
	Login       Login
	Logout      Logout
	Service     *iam.Service
}

func (settings IAMSettings) LogoutMenuEntry(wnd core.Window) *ui.MenuEntry {
	return ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
		menuEntry.Link(settings.Logout.ID, wnd, nil)
		menuEntry.Title().Set("Abmelden")
		menuEntry.Icon().Set(icon.ArrowLeftStartOnRectangle)
	})
}

// DefaultMenuEntry returns a user management menu entry.
func (settings IAMSettings) DefaultMenuEntry(wnd core.Window) *ui.MenuEntry {
	return ui.NewMenuEntry(func(usm *ui.MenuEntry) {
		usm.Title().Set("Nutzerverwaltung")
		usm.Icon().Set(icon.Users)
		usm.Menu().Append(
			ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
				menuEntry.Link(settings.Permissions.ID, wnd, nil)
				menuEntry.Title().Set("Berechtigungen")
				menuEntry.Icon().Set(icon.Users)
			}),
			ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
				menuEntry.Link(settings.Roles.ID, wnd, nil)
				menuEntry.Title().Set("Rollen")
				menuEntry.Icon().Set(icon.Cog6Tooth)
			}),
			ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
				menuEntry.Link(settings.Groups.ID, wnd, nil)
				menuEntry.Title().Set("Gruppen")
				menuEntry.Icon().Set(icon.Cog6Tooth)
			}),
			ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
				menuEntry.Link(settings.Users.ID, wnd, nil)
				menuEntry.Title().Set("Nutzerkonten")
				menuEntry.Icon().Set(icon.Users)
			}),
		)
	})
}

type Groups struct {
	// default to iam/groups
	ID         ora.ComponentFactoryId
	Repository iam.GroupRepository
}

type Roles struct {
	// default to iam/roles
	ID         ora.ComponentFactoryId
	Repository iam.RoleRepository
}

type Login struct {
	// default to iam/login
	ID ora.ComponentFactoryId
}

type Logout struct {
	// default to iam/logout
	ID ora.ComponentFactoryId
}

type Users struct {
	// defaults to iam/users
	ID         ora.ComponentFactoryId
	Repository iam.UserRepository
}

type Permissions struct {
	// defaults to iam/permissions
	ID          ora.ComponentFactoryId
	Permissions *iam.Permissions
}

type Sessions struct {
	Repository iam.SessionRepository
}

func (c *Configurator) IAM(settings IAMSettings) IAMSettings {
	if settings.Permissions.ID == "" {
		settings.Permissions.ID = "iam/permissions"
	}

	if settings.Permissions.Permissions == nil {
		settings.Permissions.Permissions = iam.PermissionsFrom[iam.Permission](nil)
	}

	if settings.Users.ID == "" {
		settings.Users.ID = "iam/users"
	}

	if settings.Login.ID == "" {
		settings.Login.ID = "iam/login"
	}

	if settings.Logout.ID == "" {
		settings.Logout.ID = "iam/logout"
	}

	if settings.Roles.ID == "" {
		settings.Roles.ID = "iam/roles"
	}

	if settings.Groups.ID == "" {
		settings.Groups.ID = "iam/groups"
	}

	if settings.Decorator == nil {
		settings.Decorator = func(wnd core.Window, page *ui.Page, content core.Component) {
			page.Body().Set(ui.NewScaffold(func(scaffold *ui.Scaffold) {
				subject := wnd.Subject()
				scaffold.NavigationComponent().Set(ui.NewNavigationComponent(func(navigationComponent *ui.NavigationComponent) {
					navigationComponent.Alignment().Set(ora.Leading)
					navigationComponent.Menu().Append(
						ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
							menuEntry.Link(settings.Login.ID, wnd, nil)
							if subject.Valid() {
								menuEntry.Title().Set("Mit einem anderen Konto anmelden")
								menuEntry.Icon().Set(icon.UserPlus)
							} else {
								menuEntry.Title().Set("Anmelden")
								menuEntry.Icon().Set(icon.UserPlus)
							}
						}),
					)

					if subject.Valid() {
						navigationComponent.Menu().Append(
							settings.LogoutMenuEntry(wnd),
						)
					}

					if auth.OneOf(subject, iam.ReadPermission, iam.ReadUser) {
						navigationComponent.Menu().Append(
							settings.DefaultMenuEntry(wnd),
						)
					}

				}))
				scaffold.Body().Set(content)
			}))
		}
	}

	if settings.Users.Repository == nil {
		settings.Users.Repository = json.NewSloppyJSONRepository[iam.User](c.EntityStore("iam.users"))
	}

	if settings.Sessions.Repository == nil {
		settings.Sessions.Repository = json.NewSloppyJSONRepository[iam.Session](c.EntityStore("iam.sessions"))
	}

	if settings.Roles.Repository == nil {
		settings.Roles.Repository = json.NewSloppyJSONRepository[iam.Role](c.EntityStore("iam.roles"))
	}

	if settings.Groups.Repository == nil {
		settings.Groups.Repository = json.NewSloppyJSONRepository[iam.Group](c.EntityStore("iam.groups"))
	}

	service := settings.Service
	if settings.Service == nil {
		service = iam.NewService(
			settings.Permissions.Permissions,
			settings.Users.Repository,
			settings.Sessions.Repository,
			settings.Roles.Repository,
			settings.Groups.Repository,
		)
		settings.Service = service
	}

	c.iamSettings = settings

	if err := service.Bootstrap(); err != nil {
		panic(fmt.Errorf("cannot bootstrap IAM service: %v", err))
	}
	c.Component(settings.Permissions.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Permissions(wnd, page, service))
		return page
	})

	c.Component(settings.Login.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Login(wnd, page, service))
		return page
	})

	c.Component(settings.Logout.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Logout(wnd, service))
		return page
	})

	c.Component(settings.Users.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Users(wnd, page, service))
		return page
	})

	c.Component(settings.Roles.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Roles(wnd, page, service))
		return page
	})

	c.Component(settings.Groups.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Groups(wnd, page, service))
		return page
	})

	c.AddOnWindowCreatedObserver(func(wnd core.Window) {
		wnd.UpdateSubject(service.Subject(wnd.SessionID()))
	})

	return settings
}
