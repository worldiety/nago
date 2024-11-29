package application

import (
	"fmt"
	"go.wdy.de/nago/annotation"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/auth/iam/iamui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

type IAMSettings struct {
	Decorator   func(wnd core.Window, view core.View) core.View
	Permissions Permissions
	Users       Users
	Sessions    Sessions
	Roles       Roles
	Groups      Groups
	Login       Login
	Logout      Logout
	Dashboard   Dashboard
	Profile     Profile
	Service     *iam.Service
}

func (settings IAMSettings) DecorateRootView(factory func(wnd core.Window) core.View) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		if settings.Decorator == nil {
			panic("settings has not been initialized correctly")
		}
		view := factory(wnd)
		return settings.Decorator(wnd, view)
	}
}

func (settings IAMSettings) LogoutMenuEntry(wnd core.Window) ui.ScaffoldMenuEntry {
	return ui.ForwardScaffoldMenuEntry(wnd, heroSolid.ArrowLeftStartOnRectangle, "Abmelden", settings.Logout.ID)
}

// AdminMenu returns a user management menu entry.
func (settings IAMSettings) AdminMenu(wnd core.Window) ui.ScaffoldMenuEntry {
	return ui.ParentScaffoldMenuEntry(wnd, heroSolid.Users, "Nutzer",
		ui.ForwardScaffoldMenuEntry(wnd, heroSolid.Users, "Berechtigungen", settings.Permissions.ID),
		ui.ForwardScaffoldMenuEntry(wnd, heroSolid.Cog6Tooth, "Rollen", settings.Roles.ID),
		ui.ForwardScaffoldMenuEntry(wnd, heroSolid.Cog6Tooth, "Gruppen", settings.Groups.ID),
		ui.ForwardScaffoldMenuEntry(wnd, heroSolid.Users, "Konten", settings.Users.ID),
	)
}

func (settings IAMSettings) LoginStatusMenu(wnd core.Window) ui.ScaffoldMenuEntry {
	if wnd.Subject().Valid() {
		return ui.ForwardScaffoldMenuEntry(wnd, heroSolid.UserPlus, "Mit einem anderen Konto anmelden", settings.Login.ID)
	} else {
		return ui.ForwardScaffoldMenuEntry(wnd, heroSolid.UserPlus, "Anmelden", settings.Login.ID)
	}
}

type Profile struct {
	// default to iam/profile
	ID core.NavigationPath
}

func (p Profile) Path() core.NavigationPath {
	if p.ID == "" {
		return "iam/profile"
	}

	return p.ID
}

type Groups struct {
	// default to iam/groups
	ID         core.NavigationPath
	Repository iam.GroupRepository
}

type Roles struct {
	// default to iam/roles
	ID         core.NavigationPath
	Repository iam.RoleRepository
}

type Login struct {
	// default to iam/login
	ID core.NavigationPath
}

type Logout struct {
	// default to iam/logout
	ID core.NavigationPath
}

type Dashboard struct {
	// default to iam/dashboard
	ID core.NavigationPath
}

type Users struct {
	// defaults to iam/users
	ID         core.NavigationPath
	Repository iam.UserRepository
}

type Permissions struct {
	// defaults to iam/permissions
	ID core.NavigationPath
	// If nil [iam.DefaultPermissions] is used.
	Permissions *iam.Permissions
}

type Sessions struct {
	Repository iam.SessionRepository
}

func (c *Configurator) IAM(settings IAMSettings) IAMSettings {
	if settings.Permissions.ID == "" {
		settings.Permissions.ID = "iam/permissions"
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

	if settings.Dashboard.ID == "" {
		settings.Dashboard.ID = "iam/dashboard"
	}

	if settings.Decorator == nil {
		settings.Decorator = func(wnd core.Window, content core.View) core.View {
			var items []ui.ScaffoldMenuEntry
			items = append(items, settings.LoginStatusMenu(wnd))
			if wnd.Subject().Valid() {
				items = append(items, settings.LogoutMenuEntry(wnd))
			}

			if auth.OneOf(wnd.Subject(), iam.ReadPermission, iam.ReadUser) {
				items = append(items, settings.AdminMenu(wnd))
			}

			return ui.Scaffold(ui.ScaffoldAlignmentLeading).Menu(items...).Body(content)
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

	if settings.Permissions.Permissions == nil {
		settings.Permissions.Permissions = iam.PermissionsFrom(annotation.Permissions())
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
	c.RootView(settings.Permissions.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Permissions(wnd, service))
	})

	c.RootView(settings.Login.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Login(wnd, service))
	})

	c.RootView(settings.Logout.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Logout(wnd, service))
	})

	c.RootView(settings.Users.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Users(wnd, service))
	})

	c.RootView(settings.Roles.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Roles(wnd, service))
	})

	c.RootView(settings.Groups.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Groups(wnd, service))
	})

	c.RootView(settings.Dashboard.ID, func(wnd core.Window) core.View {
		return settings.Decorator(wnd, iamui.Dashboard(wnd, iamui.DashboardModel{
			Accounts:    settings.Users.ID,
			Permissions: settings.Permissions.ID,
			Groups:      settings.Groups.ID,
			Roles:       settings.Roles.ID,
		}))
	})

	c.AddOnWindowCreatedObserver(func(wnd core.Window) {
		wnd.UpdateSubject(service.Subject(wnd.SessionID()))
	})

	return settings
}
