---
# Static content
title: Systems
weight: 3
prev: /docs/components
next: /docs/examples
sidebar:
  open: false
---

# What is a “System”?

A system in the NAGO framework is an optional, self-contained functional unit that can be activated via the Configurator. Each system encapsulates:
- Business logic (UseCases)
- UI components
- Configuration
- Often also data persistence and admin interfaces

Systems are designed to be modular and can be integrated independently, depending on the needs of the application.

## List of Systems

| System | Description | Details |
|--------|------------|---------|
| `Admin Management` | The Admin Management system enables the Admin Center - a central interface that provides access to all available settings and features of other activated systems. Once enabled, it automatically integrates:<ul><li>Navigation entries for all systems with UI pages</li><li>A consistent layout and theming for admin functionality</li><li>Role-based access to systems (in combination with Permission Management).</li></ul> | [Details »](admin_management) |
| `CMS Management` |  | [Details »](cms_management) |
| `Group Management` | Macht tolle Dinge. | [Details »](group_management) |
| `User Management` | The User Management system is responsible for creating, managing, and maintaining user accounts within the platform. It provides both administrative and self-service features, allowing administrators to manage users and permissions, while enabling end users to maintain their own profiles. Typical workflows include:<ul><li>Creating and deleting user accounts</li><li>Assigning roles, groups, and permissions</li><li>Managing user profile data and contact details</li><li>Password management (self-service and administrative)</li><li>Email verification and account activation notifications.</li></ul> | [Details »](user_management) |

## How to activate a System

Systems can be activated via the Configurator.
Each system provides a method that returns a structured type, such as SettingsManagement, containing the system’s UseCases and UI configuration.

```go
settingsManagement, err := configurator.SettingsManagement()
```
### With option & std Package

To simplify activation, the option and std package can be used.
Depending on the return type, you can use option.Must, std.Must or option.MustZero to safely unwrap the result or panic on error.
```go
import 	(
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/std"
)

// Use when the method returns nothing
option.MustZero(configurator.StandardSystems())

// Use when the method returns something
settingsManagement := option.Must(configurator.SettingisManagement())

// Or
settingsManagement := std.Must(configurator.SettingisManagement())
```

### Enabling all Standard Systems

To quickly get started with the default NAGO functionality, you can use the built-in method:
```go
option.MustZero(configurator.StandardSystems())
```

This method enables a pre-defined set of systems that are commonly used in most applications.
The following systems are enabled when calling StandardSystems():
- Admin Management
- Backup Management
- Billing Management
- Group Management
- License Management
- Mail Management
- Permission Management
- Role Management
- Secret Management
- Session Management
- Settings Management
- Template Management
- User Management

Each of these systems becomes available with its default UI pages and functionality, including integration into the Admin Center.

### Custom Configuration
If you want full control over which systems are active in your application, you can skip StandardSystems() and activate each system manually.
This approach is especially useful when you only need a subset of systems.

## Note on System dependencies
Some systems internally activate other systems they depend on.
For example, SessionManagement() will automatically enable UserManagement(), MailManagement() and some more.
You don't need to handle dependencies manually, but be aware that activating one system may implicitly initialize others.

{{< callout type="warning" >}}
Ensure that a user has all necessary permissions for the system so that it appears in the admin center.
{{< /callout >}}
