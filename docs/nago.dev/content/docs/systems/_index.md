---
title: Systems
weight: 6
sidebar:
  open: false
---

A system is a ready-made functional unit of Nago, such as user management, mail delivery or a file drive. It
brings its use cases, its permissions, its persistence and usually its admin pages, which appear as cards in
the [admin center](admin_management/).

## Enable a system

Each system is enabled by a method of the configurator or by an `Enable` function in its `cfg` package. Both
return a struct with the use cases and the page paths of the system:

```go
users := std.Must(cfg.UserManagement())       // application.UserManagement
drives := std.Must(cfgdrive.Enable(cfg))      // cfgdrive.Management
```

Keep the returned value to call the use cases of the system later, e.g. `users.UseCases.FindByID`.

Enabling is idempotent: the first call creates the system, later calls return the same one. A system enables
the systems it depends on, so you never have to care about the order.

### Standard systems

```go
option.MustZero(cfg.StandardSystems())
```

`StandardSystems` enables Image, Admin, User, Backup, Mail, Secret, Template and Session Management. Session
Management is mandatory anyway and pulls in most of them, so every application has User, Role, Group,
Permission, Settings, Admin, Mail, Secret, Template, Image and Theme Management. `StandardSystems`
additionally enables Backup Management. All other systems must be enabled explicitly.

{{< callout type="info" >}}
A user only sees the admin cards of a system if they have the matching permissions. During development, the
[bootstrap admin](user_management/#bootstrap-admin) has all `nago.*` permissions.
{{< /callout >}}

## Available systems

### Identity and access

| System                                       | Enable                       | Description                                                        |
|----------------------------------------------|------------------------------|--------------------------------------------------------------------|
| [User Management](user_management/)          | `cfg.UserManagement()`       | user accounts, registration, profile, passwords, consents         |
| [Session Management](session_management/)    | `cfg.SessionManagement()`    | login, logout and single sign-on (always enabled)                 |
| [Role Management](role_management/)          | `cfg.RoleManagement()`       | roles which bundle permissions, system roles                      |
| [Group Management](group_management/)        | `cfg.GroupManagement()`      | groups of users for sharing resources                             |
| [Permission Management](permission_management/) | `cfg.PermissionManagement()` | declaring and listing permissions                              |
| [ReBAC](rebac_management/)                   | `cfg.RDB()`, `cfgrebac.Enable` | relations database for all grants, editor for per-resource grants |
| [User Circles](user_circle_management/)      | `cfgusercircle.Enable`       | delegated administration of subsets of users                      |
| [Token Management](token_management/)        | `cfg.TokenManagement()`      | API access tokens                                                 |

### Administration and operations

| System                                       | Enable                       | Description                                                        |
|----------------------------------------------|------------------------------|--------------------------------------------------------------------|
| [Admin Management](admin_management/)        | `cfg.AdminManagement()`      | the admin center (always enabled)                                 |
| [Settings Management](settings_management/)  | `cfg.SettingsManagement()`   | global settings with generated forms                              |
| [Theme Management](theme_management/)        | `cfg.ThemeManagement()`      | colors, fonts, logos and legal links (always enabled)             |
| [Backup Management](backup_management/)      | `cfg.BackupManagement()`     | backup, restore and master key                                    |
| [Secret Management](secret_management/)      | `cfg.SecretManagement()`     | encrypted vault for credentials                                   |
| [Scheduler Management](scheduler_management/) | `cfgscheduler.Enable`       | background jobs                                                    |
| [Migration Management](migration_management/) | `cfgmigration.Enable`       | status of data migrations                                          |
| [Inspector Management](inspector_management/) | `cfginspector.Enable`       | browse and edit all stores                                         |
| [Localization Management](localization_management/) | `cfglocalization.Enable` | translate texts at runtime                                     |

### Content and communication

| System                                       | Enable                       | Description                                                        |
|----------------------------------------------|------------------------------|--------------------------------------------------------------------|
| [Mail Management](mail_management/)          | `cfg.MailManagement()`       | outgoing mail queue via SMTP                                      |
| [Template Management](template_management/)  | `cfg.TemplateManagement()`   | editable text, HTML and PDF templates                             |
| [Image Management](image_management/)        | `cfg.ImageManagement()`      | image upload and delivery in fitting sizes (always enabled)       |
| [Drive Management](drive_management/)        | `cfgdrive.Enable`            | file storage with folders, versions and access control            |
| [Data Import Management](data_import_management/) | `cfgdataimport.Enable`  | import CSV, JSON and PDF forms with review                         |
| [Signature Management](signature_management/) | `cfgsignature.Enable`       | unqualified electronic signatures                                  |
| [SMS Management](sms_management/)            | `cfgsms.Enable`              | send SMS                                                           |
| [Chatbot Management](chatbot_management/)    | `cfgchatbot.Enable`          | send chat messages through a bot                                   |
| [REST APIs (HAPI)](hapi_management/)         | `cfghapi.Enable`             | REST endpoints with OpenAPI                                        |
| [AI Management](ai_management/)              | `cfgai.Enable`               | LLM providers, chat sessions and an assistant                      |
| [Requirements (Speclink)](speclink_management/) | `cfgspeclink.Enable`      | browse the requirements compiled into the application             |

The `cfg` packages are below `go.wdy.de/nago/application/<system>/cfg`, for example
`go.wdy.de/nago/application/drive/cfg`.

### Experimental

Nago contains further modules which are still under development and not documented here yet: Flow
(`application/flow/cfg`, data modeling and forms), CMS (`application/cms/cfg`, static pages) and Workflow
(`application/workflow/cfg`, an event-driven workflow engine).
