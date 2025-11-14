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
| `Admin Management` | It provides the Admin Center, a central hub that aggregates and displays the administration pages of all connected subsystems. Key features include:<ul><li>Centralized access to management systems</li><li>Automatic integration of built-in systems (User, Role, Session, Permission, etc. )</li><li>Role- and permission-based access control</li><li>Extensibility by allowing developers to register custom groups and cards</li></ul> | [Details »](admin_management) |
| `Backup Management` | It provides access to backup and restore functionality in Nago. It includes use cases for creating backups, restoring from backups, exporting the master key, and replacing the master key. UseCases:<ul><li>Backup: Creates a full backup of the application. Encrypted stores remain encrypted; the master key is not included. - Restore: Restores the application state from a backup file. This overwrites existing data. Encrypted stores require the master key. - ExportMasterKey: Returns the current Nago master key, required to decrypt encrypted stores in backups. - ReplaceMasterKey: Replaces the current master key. Encrypted stores can only be decrypted after restart with the new key.</li></ul> | [Details »](backup_management) |
| `Drive Management` | It provides a file storage and file management subsystem that can be used from both frontend UI components and backend use-cases. The system implements an owner/group/permission model. Files contain Owner, Group and FileMode fields, and permission checks are implemented in File. CanRead, File. CanWrite, File. CanDelete, and File. CanRename. Shares and resource-level permissions are also considered by the permission checks. Use the provided use-cases (drive. OpenRoot, drive. Put, drive. MkDir, drive. Delete, drive. Stat, drive. Zip, drive. Get, drive. Rename) to integrate Drive into your application logic or to expose it through custom APIs. | [Details »](drive_management) |
| `Group Management` | It provides UseCases for creating and managing user groups. Groups are used to bundle users together and control access to certain pages or resources. They can be created, edited, and deleted via UI or code. Group membership is managed through UserManagement. A special "System" group is created automatically for internal services and is not intended for real users. | [Details »](group_management) |
| `Image Management` | It provides backend functionality for storing, processing, and serving images. Images are represented as SrcSets, which contain multiple scaled variants (e. g. , thumbnails, previews, high-resolution versions) of the same source. Features:<ul><li>Validation of uploaded images (file size, supported format, dimensions)</li><li>Creation of SrcSets with automatically generated downscaled variants</li><li>Deduplication of image data in the underlying blob store</li><li>Best-fit selection of images for given dimensions and object-fit strategies</li><li>Secure opening of image readers with permission-aware access control</li><li>Default HTTP endpoint (/api/nago/v1/image) for image delivery, including caching</li></ul> | [Details »](image_management) |
| `Permission Management` | It is responsible for managing permission. Permission. These are the most fine-grained access control unit in the system. They are always defined at development time in code, and cannot be created or modified at runtime. Permissions can be granted to individual users or roles. | [Details »](permission_management) |
| `Role Management` | It provides UseCases for creating, editing, deleting roles. Roles are used to grant users bundled permissions. They can be created, edited, and deleted via UI or code. Roles assignment is managed through UserManagement. | [Details »](role_management) |
| `Scheduler Management` | It provides centralized management of background processes (schedulers). Developers can register recurring or one-shot jobs at application startup, monitor their execution, and interact with them through the Admin Center UI. | [Details »](scheduler_management) |
| `Secret Management` | It is responsible for storing, managing, and controlling access to sensitive data such as passwords, API keys, and external system configurations. It stores all data in an encrypted blob store and allows secure sharing with other users or groups. Typical workflows include:<ul><li>Creating, updating, and deleting secrets</li><li>Sharing secrets with users or groups</li><li>Defining new secret types in the source code by implementing the secret. Credentials interface.</li></ul> | [Details »](secret_management) |
| `Session Management` | It provides functionality for handling user sessions, including login, logout, authentication state, and Single Sign-On (SSO) via the Nago Login Service (NLS). A session is identified by a unique cookie-based ID and represents the persistent state of a client. This ID is stable across tabs and device restarts. Key features include:<ul><li>Session lifecycle management (create, find, clear, timeout handling)</li><li>Authentication via email/password or direct user ID</li><li>Single Sign-On support (start, exchange, refresh NLS flows)</li><li>Logout and session invalidation</li><li>Tracking of creation and authentication timestamps</li><li>Storing small key-value pairs in session context</li></ul> | [Details »](session_management) |
| `Settings Management` | It provides centralized configuration for global and per-user settings and has two main responsibilities:  1. Manage general application-level user settings such as:<ul><li>free registration (enable/disable)</li><li>forgot password functionality</li><li>domain whitelist for registration</li><li>default roles and groups for new and anonymous users</li><li>User Management: GDPR consent texts</li><li>Theme Management: global theme configuration</li><li>Schedule Management: job lifetime and cleanup rules</li></ul> | [Details »](settings_management) |
| `Template Management` | It is responsible for creating, editing, and managing reusable templates. It provides a centralized way to separate content from code and supports multiple output formats, including Go HTML templates for emails, plain text, and various text-to-PDF workflows (Typst, LaTeX, AsciiDoc). The system is primarily used by other modules such as Mail Management for standardized emails, but can also support document generation (e. g. , reports, certificates, invoices). By centralizing template logic, it increases flexibility, maintainability, and consistency across the platform. | [Details »](template_management) |
| `Theme Management` | Theme Management handles the configuration of theme and corporate identity settings. It allows to define logos and app icons (for dark and light mode), configure legal information (e. g. Impressum, Privacy Policy, Terms, User Agreement), and set provider contact details such as responsible entity, contact email, and API documentation URL. Additionally, developers can define fonts and base colors (main, interactive, accent) directly via code. Colors can differ between dark and light mode. | [Details »](theme_management) |
| `Token Management` | It configures and provides the backend for managing API access tokens. Tokens are used to authenticate requests against REST APIs. They can carry groups, roles, permissions, and licenses, similar to regular users. This enables external applications or services to act as authenticated subjects. It is typically used together with cfghapi. Management to secure API endpoints with bearer tokens. | [Details »](token_management) |
| `User Management` | It is responsible for creating, managing, and maintaining user accounts within the platform. It provides both administrative and self-service features, allowing administrators to manage users and permissions, while enabling end users to maintain their own profiles. Typical workflows include:<ul><li>Creating and deleting user accounts</li><li>Assigning roles, groups, and permissions</li><li>Managing user profile data and contact details</li><li>Password management (self-service and administrative)</li><li>Email verification and account activation notifications.</li></ul> | [Details »](user_management) |

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

{{< callout type="info" >}}
Each of these systems becomes available with its default UI pages and functionality, including integration into the Admin Center.
{{< /callout >}}

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
