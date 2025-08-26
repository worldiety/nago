---
title: Permission Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries/admin_center.png"
  - src: "/images/systems/permission_management/galleries/overview.png"
galleryAssignPermissions:
  - src: "/images/systems/permission_management/galleries/assign_permissions.png"
  - src: "/images/systems/permission_management/galleries/assign_permissions_to_role.png"
---

The Permission Management system manages all available permissions across the platform.  
Permissions are the most fine-grained unit of access control and can be used for specific use cases in the domain.

Each application use case defines its own permissions **at development time**.  
Permissions cannot be created or modified at runtime, but they can be **viewed** and **assigned** to users.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Permission Management offers the following key functions:

### Permission lifecycle
- Permissions are defined in code at development time
- Default permissions are automatically created when a system is activated
- New permissions can be declared programmatically via the `permission.Declare` function
- Permissions **cannot** be created or edited in the UI

### Permission assignment
- Permissions **cannot** be assigned directly to groups
- Instead, permissions can be grouped into [Roles](../role_management/) which then can be assigned to users
- Permissions can also be assigned directly to users via [User Management](../user_management/)

{{< swiper name="galleryAssignPermissions" loop="false" >}}

### Default permissions
- Each system provides its own default permissions when activated
- Example (from Mail Management):
```go
var (
	PermSendMail             = permission.Declare[SendMail]("nago.mail.send", "Mail Senden", "Träger dieser Berechtigung können Mails versenden.")
	PermInitDefaultTemplates = permission.Declare[SendMail]("nago.mail.init_default_templates", "Standard Templates setzen", "Träger dieser Berechtigung können die Standard Mail templates aktivieren.")
)
```

## Code usage

```go
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"

	"fmt"
	"time"
)

var PermGetCurrentTime = permission.Declare[GetCurrentTime]("nago.current.time", "Get current time", "Holders of this permissions can retrieve the current time.")

type GetCurrentTime func(subject auth.Subject) time.Time

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.create.permission")
		
		std.Must(cfg.PermissionManagement())

		cfg.RootViewWithDecoration("current_time", func(wnd core.Window) core.View {
			if err := wnd.Subject().Audit(PermGetCurrentTime); err != nil {
				return alert.Banner("Error", "You are not allowed to see the current time")
			}

			return ui.Text(fmt.Sprintf("%s", time.Now()))
		})

	}).Run()
}
```

## Dependencies
**Requires:**
- None

**Is required by:**
- [User Management](../user_management/)
- [Session Management](../session_management/)

## Activation
This system is activated via:
```go
option.Must(cfg.PermissionManagement())
```

```go
permissionManagement := option.Must(cfg.PermissionManagement())
```