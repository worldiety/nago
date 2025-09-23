---
title: Admin Management
galleryNewAdminCenterGroup:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/admin_management/galleries/admin_center_group/create.png"
---

The Admin Management system provides the **Admin Center** – a central entry point for managing and configuring all available subsystems.  
It aggregates the administration pages of all connected systems (e.g., User, Role, Session, Permission, Billing, Backup, Secret, Template) and makes them accessible in a unified interface.

## Functional areas
Admin Management provides the following key functions:

### Central administration hub
- Provides the **Admin Center** as a unified entry point for administrators
- Integrates all connected subsystems automatically
- Groups systems into categories and displays them as cards

### Access control
- Each card can be associated with a **Role** or **Permission**
- Ensures that only authorized users can access specific admin functions
- Enforces subject validation (no access without valid user context)

### Extensibility
- Developers can register their own admin groups and cards
- Allows integration of custom systems into the Admin Center

### Example: Registering a custom admin group - with restricted access

```go
import (
	"go.wdy.de/nago/application/admin"
    "go.wdy.de/nago/application/permission"
    "go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

var PermFindDashboards = permission.Declare[FindDashboards]("nago.analytics.dashboards", "Get access to all analytic dashboards", "Holders of this permissions can find all analytic dashboards.")

type FindDashboards func(subject auth.Subject) error

// Example: add a custom "Analytics" section to the Admin Center
cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
	if !subject.Valid() {
		return admin.Group{}
	}

	return admin.Group{
		Title: "Analytics",
		Entries: []admin.Card{
			{
				Title:      "Reports",
				Text:       "View and generate custom reports",
				Target:     "analytics/reports",
				Permission: PermFindDashboards,
			},
			{
				Title:  "Dashboards",
				Text:   "Manage and customize dashboards",
				Target: "analytics/dashboards",
				Role:   "nago.admin",
			},
		},
	}
})
```

{{< swiper name="galleryNewAdminCenterGroup" loop="false" >}}

{{< callout type="info" >}}
By adding groups via AddAdminCenterGroup, custom systems can seamlessly integrate into the Admin Center alongside built-in ones.
{{< /callout >}}

## Dependencies
**Requires:**
- [Billing Management](../billing_management/)

**Optionally integrates with (if present):**
- [Backup Management](../backup_management/)
- [Group Management](../group_management/)
- [License Management](../license_management/)
- [Mail Management](../mail_management/)
- [Permission Management](../permission_management/)
- [Role Management](../role_management/)
- [Secret Management](../secret_management/)
- [Session Management](../session_management/)
- [Template Management](../template_management/)
- [User Management](../user_management/)

**Is required by:**
- None directly, but it serves as the **UI entry point** for most other systems.

## Activation
This system is activated via:
```go
std.Must(cfg.AdminManagement())
```

```go
adminManagement := std.Must(cfg.AdminManagement())
```