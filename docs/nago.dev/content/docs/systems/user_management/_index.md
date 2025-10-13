---
title: User Management
galleryLifecycle:
  - src: "/images/systems/user_management/galleries/lifecycle/registration_01.png"
  - src: "/images/systems/user_management/galleries/lifecycle/registration_02.png"
  - src: "/images/systems/user_management/galleries/lifecycle/registration_03.png"
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries/admin_center.png"
  - src: "/images/systems/user_management/galleries/user_table.png"
  - src: "/images/systems/user_management/galleries/lifecycle/create_user.png"
  - src: "/images/systems/user_management/galleries/lifecycle/edit_user.png"
  - src: "/images/systems/user_management/galleries/lifecycle/edit_user_2.png"
galleryProfileInfos:
  - src: "/images/systems/user_management/galleries/user_account.png"
  - src: "/images/systems/user_management/galleries/user_account_2.png"
  - src: "/images/systems/user_management/galleries/profile_infos/edit_account_details.png"
galleryPasswordManagement:
  - src: "/images/systems/user_management/galleries/password_management/reset_pw_2.png"
  - src: "/images/systems/user_management/galleries/password_management/reset_pw_1.png"
  - src: "/images/systems/user_management/galleries/user_account.png"
  - src: "/images/systems/user_management/galleries/user_account_2.png"
  - src: "/images/systems/user_management/galleries/password_management/change_pw_user.png"
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries/admin_center.png"
  - src: "/images/systems/user_management/galleries/user_table.png"
  - src: "/images/systems/user_management/galleries/password_management/change_pw_admin.png"
galleryRolesPermsGroups:
    - src: "/images/systems/shared/admin_center.png"
    - src: "/images/systems/user_management/galleries/admin_center.png"
    - src: "/images/systems/user_management/galleries/roles_perms_groups/edit_roles.png"
    - src: "/images/systems/user_management/galleries/roles_perms_groups/edit_groups.png"
    - src: "/images/systems/user_management/galleries/roles_perms_groups/edit_perms.png"
galleryRelatedSettings:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries//related_settings/admin_center.png"
  - src: "/images/systems/user_management/galleries//related_settings/edit_registration_settings.png"
---
The User Management system is responsible for creating, managing, and maintaining user accounts within the platform.
It provides both administrative and self-service features, allowing administrators to manage users and permissions, while enabling end users to maintain their own profiles.
Typical workflows include:
- Creating and deleting user accounts
- Assigning roles, groups, and permissions
- Managing user profile data and contact details
- Password management (self-service and administrative)
- Email verification and account activation notifications

{{< callout type="info" >}}
Ensure that an SMTP server is configured if your workflows need to send emails in any case.
For more information how to configure an SMTP server have a look at [Secret Management](../secret_management/)
{{< /callout >}}

## Functional areas
User Management offers a wide range of features, grouped into the following main areas:

### Account lifecycle
- Create, update, deactivate, or delete user accounts
- Register a new account (self-service)
- Send notification when a new account is created

{{< swiper name="galleryLifecycle" loop="false" >}}

### Profile and contact information
- View and edit personal profile information
- Update contact details
- Manage consent options

{{< swiper name="galleryProfileInfos" loop="false" >}}

### Authentication & password management
- Change own password
- Reset password via confirmation code
- Force password change (admin)

{{< swiper name="galleryPasswordManagement" loop="false" >}}

### Roles, permissions, and groups
- Assign or revoke roles
- Assign or revoke group memberships
- Grant or revoke permissions

{{< swiper name="galleryRolesPermsGroups" loop="false" >}}

## Related settings
While not strictly part of the core User Management functionality, certain configuration options in Settings Management influence how User Management behaves, such as:
- Whitelisted email domains
- Automatic role and group assignments for anonymous users

These dependencies are handled by the Settings Management system but are relevant for User Management workflows.

{{< swiper name="galleryRelatedSettings" loop="false" >}}

## Dependencies
User Management only works properly if the [Session Management](../session_management/) is activated.

**Requires:**
- [License Management](../license_management/)
- [Role Management](../role_management/)
- [Permission Management](../permission_management/)
- [Group Management](../group_management/)
- [Settings Management](../settings_management/)

If these are not already active, they will be enabled automatically when User Management is activated.

**Is required by:**
- [Billing Management](../license_management/)
- [Mail Management](../mail_management/)
- [Secret Management](../secret_management/)
- [Token Management](../token_management/)
- [Session Management](../session_management/)

## Activation
This system is activated via the configurator:
```go
std.Must(cfg.UserManagement())
```

```go
userManagement := std.Must(cfg.UserManagement())
```
