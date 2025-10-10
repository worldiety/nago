---
title: Role Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries/admin_center.png"
  - src: "/images/systems/role_management/galleries/overview.png"
galleryCreateRole:
  - src: "/images/systems/role_management/galleries/create.png"
galleryAssignRole:
  - src: "/images/systems/role_management/galleries/assign_role.png"
---
The Role Management system provides functionality to create, manage, and assign roles.  
Roles are collections of permissions and are the recommended way to assign multiple permissions to individual users.  
Roles can be created, edited, and deleted either through the UI or programmatically in the code.

Roles can only be assigned to individual users. Assigning roles to groups is not supported. Users can hold multiple roles simultaneously.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Role Management offers the following key functions:

### Role creation and editing
- Create new roles with a title, description, and assigned permissions
- Edit existing roles to update metadata or permissions
- Delete roles that are no longer needed

{{< swiper name="galleryCreateRole" loop="false" >}}

### Role assignment
- Assign one or more roles to individual users
- Remove roles from users as needed
- Membership is managed via [User Management](../user_management/)

{{< swiper name="galleryAssignRole" loop="false" >}}

## Code usage
```go
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

const (
	adminRole = role.ID("nago.admin")
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.create.role")

		roleManagement := std.Must(cfg.RoleManagement())

		std.Must(roleManagement.UseCases.Upsert(user.SU(), role.Role{
			ID:          adminRole,
			Name:        "Administrator",
			Description: "Full access to all features.",
			Permissions: []permission.ID{"user.PermCreate", "user.PermDelte", "user.PermFindAll"},
		}))

		cfg.RootViewWithDecoration("admin_secrets", func(wnd core.Window) core.View {
			if !wnd.Subject().HasRole(adminRole) {
				return alert.Banner("Error", "Access only for admins")
			}

			return ui.Text("Welcome, Admin! You have full access to the system.")
		})
	
	}).Run()
}
```

## Dependencies
**Requires:**
- None

**Is required by:**
- [User Management](../user_management/)
- [Token Management](../token_management/)
- [UserCircle Management](../usercircle_management/)

## Activation
This system is activated via:
```go
std.Must(cfg.RoleManagement())
```

```go
roleManagement := std.Must(cfg.RoleManagement())
```