---
title: Group Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries/admin_center.png"
  - src: "/images/systems/group_management/galleries/overview.png"
galleryCreateGroup:
  - src: "/images/systems/group_management/galleries/create.png"
galleryAssignGroup:
  - src: "/images/systems/group_management/galleries/assign_group.png"
---
The Group Management system provides functionality for creating and managing user groups.  
Groups are used to organize users and control access to certain pages or resources across the platform.  
Groups can be created, edited, and deleted either through the UI or programmatically in the code.

A special **System group** is created automatically.  
It is reserved for internal services such as initializing the SMTP server.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Group Management offers the following key functions:

### Group creation and editing
- Create new groups with a name and description
- Edit existing groups to update metadata
- Delete groups that are no longer needed

{{< swiper name="galleryCreateGroup" loop="false" >}}

### Group membership
- Users can be assigned to groups or removed from them
- Membership is managed via [User Management](../user_management/)
- Groups appear in the User Management interface once created

{{< swiper name="galleryAssignGroup" loop="false" >}}

### Permissions and access control
- Groups are containers for organizing users
- They can be used to control visibility of certain pages or resources
- Common use case: making certain sections of the platform visible only to specific groups

{{< callout type="info" >}}
Groups do not define permissions. Functional access rights are managed via [Role Management](../role_management/).
{{< /callout >}}

### System group
- Automatically created when Group Management is initialized
- Used by system-relevant services such as the SMTP server
- Not intended for real users

## Code usage
```go
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

const (
	nagoDevs = group.ID("nago.devs")
)

func main() {
    application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.create.group")

		groupManagement := std.Must(cfg.GroupManagement())
		std.Must(groupManagement.UseCases.Upsert(user.SU(), group.Group{
			ID:          nagoDevs,
			Name:        "Nago Developer",
			Description: "Devs in this project.",
		}))

		cfg.RootViewWithDecoration("dev_secrets", func(wnd core.Window) core.View {
			if !wnd.Subject().HasGroup(nagoDevs) {
				return alert.Banner("Error", "You are not a member of the group: nago devs")
			}

			return ui.Text("Welcome to the group of nago devs.")
		})
		
	}).Run()
}
```

## Dependencies
**Requires:**
- None

**Is required by:**
- [User Management](../user_management/)
- [Secret Management](../secret_management/)
- [Token Management](../token_management/)
- [UserCircle Management](../usercircle_management/)

## Activation
This system is activated via:
```go
option.Must(cfg.GroupManagement())
```

```go
groupManagement := option.Must(cfg.GroupManagement())
```