---
title: Group Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/user_management/galleries/admin_center.png"
  - src: "/images/systems/group_management/galleries/overview.png"
galleryCreateGroup:
  - src: "/images/systems/group_management/galleries/create.png"
---
The Group Management system provides functionality for creating and managing user groups.  
Groups are used to organize users and assign shared permissions, making it easier to control access to features, workflows, or pages across the platform.  
Groups can be created, edited, and deleted either through the UI or programmatically in the code.

A special **System group** is created automatically.  
It is reserved for internal services such as initializing the SMTP server.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Group Management offers the following key functions:

### Group creation and editing
- Create new groups with a name, description, and assigned rights
- Edit existing groups to update metadata and permissions
- Delete groups that are no longer needed

{{< swiper name="galleryCreateGroup" loop="false" >}}

### Group membership
- Users can be assigned to groups or removed from them
- Membership is managed via [User Management](../user_management/)
- Groups appear in the User Management interface once created

### Permissions and access control
- Groups act as containers for rights and roles
- Useful for restricting access to features, pages, or workflows
- Common use case: making certain sections of the platform visible only to specific groups

### System group
- Automatically created when Group Management is initialized
- Used by system-relevant services such as the SMTP server
- Not intended for real users

## Dependencies
**Requires:**
- None

**Is required by:**
- [User Management](../user_management/)
- [Secret Management](../secret_management/)
- [Token Management](../token_management/)
- [UserCircle Management](../usercircle_management/)

If these are not already active, they will be enabled automatically when Group Management is activated.

## Activation
This system is activated via:
```go
option.Must(cfg.GroupManagement())
```

```go
groupManagement := option.Must(cfg.GroupManagement())
```