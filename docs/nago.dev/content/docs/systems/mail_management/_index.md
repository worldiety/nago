---
title: Mail Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/mail_management/galleries/overview/overview.png"
galleryTestMail:
  - src: "/images/systems/mail_management/galleries/mail_config/test_mail.png"
  - src: "/images/systems/mail_management/galleries/mail_config/test_mail_with_template.png"
galleryLogs:
  - src: "/images/systems/mail_management/galleries/logs/overview.png"
galleryTemplates:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/mail_management/galleries/templates/overview.png"
  - src: "/images/systems/mail_management/galleries/templates/projects.png"
  - src: "/images/systems/mail_management/galleries/templates/edit.png"
---
The Mail Management system handles the sending of emails within the platform.
It requires an SMTP secret created in [Secret Management](../secret_management/_index.md).
It also integrates with [Template Management](../template_management/_index.md) to use customizable email templates for common workflows.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Mail Management offers the following key functions:

### Outgoing email handling
- Test your email configurations
- Send default test emails
- Send emails that use a template

{{< swiper name="galleryTestMail" loop="false" >}}

### Mail logs
- Keep a record of sent emails
- View message status, timestamp, and recipient
- Search and filter logs for troubleshooting

{{< swiper name="galleryLogs" loop="false" >}}

### Email templates
- Uses predefined templates from Template Management for standard messages
- Templates include registration confirmation, password reset, and various notifications
- Templates can be edited in Template Management without code changes

{{< swiper name="galleryTemplates" loop="false" >}}

## Dependencies
**Requires:**
- [Secret Management](../secret_management/_index.md) for storing SMTP credentials
- [Template Management](../template_management/_index.md) for email templates
- [User Management](../user_management/_index.md) for workflows such as password resets

If these are not already active, they will be enabled automatically when Mail Management is activated.

**Is required by:**
- [Session Management](../session_management/_index.md)

## Activation
This system is activated via:
```go
option.Must(cfg.MailManagement())
```

```go
mailManagement := option.Must(cfg.MailManagement())
```