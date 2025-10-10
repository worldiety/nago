---
title: User Circle Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_circle_management/galleries/overview/admin_center.png"
  - src: "/images/systems/user_circle_management/galleries/overview/create_1.png"
  - src: "/images/systems/user_circle_management/galleries/overview/create_2.png"
  - src: "/images/systems/user_circle_management/galleries/overview/create_3.png"
  - src: "/images/systems/user_circle_management/galleries/overview/create_4.png"
galleryManageCircle:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/admin_center.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/user.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/user_overview.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/user_actions.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/roles.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/roles_overview.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/groups.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/groups_overview.png"
  - src: "/images/systems/user_circle_management/galleries/manage_circle/groups_actions.png"
---

User Circle Management allows you to create and manage **user circles** — subsets of users that can be administered independently by delegated users.  
This enables decentralized administration, e.g., department heads or customer administrators can manage users, assign licenses, or adjust roles and groups without requiring backend access.

{{< callout type="info" >}}
User circles are ideal when parts of the user base should be managed independently, such as in multi-tenant, customer, or department-based environments.
{{< /callout >}}

## Functional areas
User Circle Management provides the following key functions:

### Circle creation and management
- Create, edit, and delete user circles
- Define administrators who can manage users within the circle
- Specify which roles, groups, and licenses can be managed within each circle
- Set membership rules (e.g., by user, email domain, group, or role)

{{< swiper name="galleryOverview" loop="false" >}}

### Circle administration by delegated users
For each created circle, the system automatically provides a dedicated administration section in the Admin Center.  
Administrators of a circle can:
- View and manage users within their circle
- Assign or revoke roles, groups, or licenses (as permitted by the circle configuration)
- Activate, deactivate, or verify users
- Remove users from the circle
  
{{< swiper name="galleryManageCircle" loop="false" >}}

{{< callout type="info" >}}
Only users designated as **administrators** of a circle will see its card in the Admin Center.
{{< /callout >}}

## Example: Delegated license administration
A company administrator could create a circle for "Department A" and assign the department head as an administrator.  
That person can then:
- See all members of Department A
- Assign or remove licenses
- Deactivate users who leave the department
- Manage access without involving system-wide administrators 
 
This allows operational teams or customers to self-manage users within defined boundaries while preserving global control and auditability.

## Dependencies
**Requires:**
- [User Management](../user_management/)
- [Group Management](../group_management/)
- [Role Management](../role_management/)
- [License Management](../license_management/)

If these systems are not already active, they will be enabled automatically when User Circle Management is activated.

**Is required by:**
- none

## Activation
This system is activated via:
```go
std.Must(cfgusercircle.Enable(cfg))
```

```go
userCircleManagement := std.Must(cfgusercircle.Enable(cfg))
```
