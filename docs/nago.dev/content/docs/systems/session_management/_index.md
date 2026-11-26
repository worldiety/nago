---
title: Session Management
---

Session Management handles login, logout and the authentication state of a client. A session is identified by
a cookie and survives tabs and restarts of the device. It supports login with mail and password and single
sign-on through the Nago Login Service (NLS).

![Login page](login.webp)

## Enable

```go
std.Must(cfg.Authentication())
// or
sessions := std.Must(cfg.SessionManagement()) // application.SessionManagement
```

Session Management is mandatory: the application panics at startup if it cannot be enabled. It enables
[Permission](../permission_management/), [Admin](../admin_management/), [Mail](../mail_management/),
[User](../user_management/) and [Settings](../settings_management/) Management, and through them
[Role](../role_management/), [Group](../group_management/), [Secret](../secret_management/),
[Template](../template_management/) and [Image](../image_management/) Management.

`application.SessionManagement` has the fields `UseCases session.UseCases` and `Pages uisession.Pages` with
the paths `Login`, `Logout` and `Authentication`.

## Use cases

| Use case              | Description                                                                  |
|-----------------------|------------------------------------------------------------------------------|
| `FindSessionByID`     | Loads a persisted session.                                                   |
| `FindUserSessionByID` | Returns a session without necessarily persisting it, e.g. for anonymous users. |
| `Login`               | Authenticates a session with mail and password.                              |
| `LoginUser`           | Marks a session as authenticated for a user ID, without any checks.         |
| `Logout`              | Logs a session out.                                                          |
| `Clear`               | Removes all sessions, only for fixing session problems.                      |
| `StartNLSFlow`        | Starts a single sign-on flow and returns the URL of the Nago Login Service.  |
| `ExchangeNLS`         | Completes the single sign-on flow.                                           |

After a successful login the event `session.Authenticated` is published.

## Permissions

Session Management declares no permissions.

## Configuration

- Sessions are stored encrypted with the master key, see [Backup Management](../backup_management/).
- The session cookie is secure unless the application runs in debug mode or `NO_SSL=true` is set.
- The login page offers registration and password reset depending on the user settings
  (`SelfRegistration`, `SelfPasswordReset`), see [User Management](../user_management/).
- Single sign-on is configured in the same settings (`SSONLSServer` and related fields).

## Pages in a frame of another site

A browser drops the session cookie (`SameSite=Lax`) in a frame of another site, so such a page can neither
download nor connect again without loading anew. `cfg.EmbeddablePages("e/*")` gives the pages of these paths,
written like the paths of `cfg.RootView`, a cookie of their own there: `SameSite=None; Secure; Partitioned`, with
the prefix `__Host-` over https. It only counts within the embedding site, and each embedding site has a session of
its own, which never has a user: `Login`, `LoginUser` and `StartNLSFlow` fail with `session.ErrEmbedded`, so no
other site can show the page of a logged-in user and trick the user into clicks. The navigation of such a frame to
any other page issues no cookie; once issued, the cookie is sent for every page of the application in a frame of
that site. A request with both cookies uses the plain one. The cookie is issued over https and http://localhost
only, never with `NO_SSL=true`.

## UI

| Path                         | Page                          |
|------------------------------|-------------------------------|
| `account/login`              | login                         |
| `account/logout`             | logout                        |
| `account/nls/authentication` | callback of the SSO flow      |

## Related

- [Tutorial: built-in IAM](/docs/examples/tutorial-26-buildin-iam/)
- [Tutorial: SSO](/docs/examples/tutorial-75-sso/)
