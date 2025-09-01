---
title: Settings Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/settings_management/galleries/overview/overview.png"
  - src: "/images/systems/settings_management/galleries/overview/edit.png"
galleryFreeRegistration:
  - src: "/images/systems/settings_management/galleries/free_registration/edit.png"
galleryNewSettings:
  - src: "/images/systems/settings_management/galleries/new_settings/overview.png"
  - src: "/images/systems/settings_management/galleries/new_settings/edit.png"
galleryNewUserConsent:
  - src: "/images/systems/settings_management/galleries/new_user_consent/overview.png"
galleryThemeSettings:
  - src: "/images/systems/settings_management/galleries/theme_settings/overview.png"
  - src: "/images/systems/settings_management/galleries/theme_settings/edit.png"
---

The Settings Management system provides a central way to configure application-wide settings.
It controls **global or user-specific settings** for other systems such as **general user registration rules** for [User Management](../user_management/)
or **app icons & private policy pages** for [Theme Management](../theme_management/).

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas

### General user settings
- Enable/disable free user registration
- Enable/disable "forgot password" functionality
- Define a **domain whitelist** (each entry represents an allowed domain suffix, e.g. `@nago-dev.com`).  
  If the list is empty, any domain is allowed.
- Define default **roles** and **groups** for new users
- Define default **roles** and **groups** for anonymous users (not inherited to valid, invalid or logged-in users)

### Free registration form rules
- Configure which contact information is required during free registration
- Use **regular expressions** for validation:
    - `^.*$` → optional
    - `^.+$` → required
    - `^(OptionA|OptionB)$` → restrict values to a fixed set
- Define specific user consent options for the registration process

{{< swiper name="galleryFreeRegistration" loop="false" >}}

{{< callout type="warning" >}}
If you restrict values to a fixed set, ensure that you add supporting text for the field, otherwise users will not know which values are valid!
{{< /callout >}}

{{< callout type="info" >}}
Use the `reflect` package to add supporting text e.g. for the salutation field in **user.Settings**.
{{< /callout >}}

```go
xreflect.SetFieldTagFor[user.Settings]("Salutation", "supportingText", "'Mrs' or 'Mr'")
```

### Example: User Management – GDPR consents
```go
import (
    "go.wdy.de/nago/application"
    "go.wdy.de/nago/application/settings"
    "go.wdy.de/nago/application/user"
	
    "github.com/worldiety/option"
)


func configureGDPRConsents(cfg *application.Configurator) {
    usrSettings := settings.ReadGlobal[user.Settings](std.Must(cfg.SettingsManagement()).UseCases.LoadGlobal)
    
    // do not append, just clear it
    usrSettings.Consents = []user.ConsentOption{
        {
            ID: consent.DataProtectionProvision,
            Register: user.ConsentText{Label: "Yes, I have read and accepted the [Privacy Policy](https://www.nago-dv.com/private_policy"},
            Required: true,
        },
        {
            ID: consent.SMS,
            Profile: user.ConsentText{Label: "Transfer of my data to the project sponsor"},
            Required: false,
        },
        {
            ID: consent.Newsletter,
            Register: user.ConsentText{
                Label:          "Yes, I would like to receive news from the Nago community via email.",
                SupportingText: "You can unsubscribe at any time in your account settings or via the unsubscribe link in the emails.",
        },
            Profile: user.ConsentText{
                Label:          "Newsletter",
                SupportingText: "Receive regular email updates",
			},
            Required: false,
        },
	}
    
    // apply settings
    option.MustZero(option.Must(cfg.SettingsManagement()).UseCases.StoreGlobal(user.SU(), usrSettings))
}
```

{{< swiper name="galleryNewUserConsent" loop="false" >}}

### Global settings
- Each system can register its own global settings
- These settings are automatically exposed via the **Admin Center UI**
- Examples:
  - **Schedule Management**: periodic jobs with custom parameters
  - **Theme Management**: system-wide design configuration

Custom settings must be defined in a struct and registered as a variant of **settings.GlobalSettings**.
They are then automatically available in the **Admin Center** UI.

### Example: Scheduler with custom settings
```go
type Settings struct {
    _        any           `title:"Events" description:"Configure events."`
    Lifetime time.Duration `json:"lifetime" label:"Lifetime of events" supportingText:"Events will be deleted after the defined lifetime."`
}

func (t Settings) GlobalSettings() bool {
    return true
}

var _ = enum.Variant[settings.GlobalSettings, Settings]()
```

Usage example in a scheduler configuration:
```go
import (
	"context"
	"time"
	
    "go.wdy.de/nago/application/scheduler"
    cfgscheduler "go.wdy.de/nago/application/scheduler/cfg"
    "go.wdy.de/nago/application/settings"
    "go.wdy.de/nago/application/user"
	
    "github.com/worldiety/option"
)

type DeleteOldest func(subject auth.Subject, settings Settings) error

schedulers := std.Must(cfgscheduler.Enable(cfg))

option.MustZero(schedulers.UseCases.Configure(user.SU(), scheduler.Options{
    ID:          "nago.dev.events.remove",
    Name:        "Remove events",
    Description: "Events that are not used must be removed.",
    Kind:        scheduler.Schedule,
    Defaults: scheduler.Settings{
        StartDelay: time.Second * 10,
        PauseTime:  time.Hour * 24,
    },
    Runner: func(ctx context.Context) error {
        myEventSettings := settings.ReadGlobal[Settings](settingsManagement.UseCases.LoadGlobal)
        return DeleteOldest(user.SU(), myEventSettings) // security note: cron job
    },
}))
```

For more information have a look at [Scheduler Management](../scheduler_management/).

{{< swiper name="galleryNewSettings" loop="false" >}}

### Example: custom theme settings via code configuration
```go
import (
    "context"
    _ "embed"
    "time"
    
    "go.wdy.de/nago/application"
    "go.wdy.de/nago/application/image"
    "go.wdy.de/nago/application/settings"
    "go.wdy.de/nago/application/user"
    "go.wdy.de/nago/presentation/core"
    
    "github.com/worldiety/option"
)

//go:embed nago_icon.png
var nagoIcon application.StaticBytes

func LoadMySettings(settingsManagement application.SettingsManagement, imageManagement application.ImageManagement) {
	userSettings := settings.ReadGlobal[user.Settings](settingsManagement.UseCases.LoadGlobal)
	storeSettings := settingsManagement.UseCases.StoreGlobal

	userSettings.SelfRegistration = true
	userSettings.SelfPasswordReset = true
	userSettings.AllowedDomains = []string{
		"@nago-dev.com",
		"@worldiety.de",
	}
	userSettings.AnonRoles = []role.ID{
		"nago.dev",
	}

	option.MustZero(storeSettings(user.SU(), userSettings))

	themeSettings := settings.ReadGlobal[theme.Settings](settingsManagement.UseCases.LoadGlobal)
	srcSet := option.Must(imageManagement.UseCases.CreateSrcSet(user.SU(), image.Options{}, core.MemFile{
		Filename:     "nago_icon.png",
		MimeTypeHint: "png",
		Bytes:        nagoIcon,
	}))

	themeSettings.PageLogoLight = srcSet.ID
	themeSettings.PageLogoDark = srcSet.ID
	themeSettings.Impress = "page/privacy"

	option.MustZero(storeSettings(user.SU(), themeSettings))
}
```

Configuration via the admin center:

{{< swiper name="galleryThemeSettings" loop="false" >}}

## Dependencies
**Requires:**
- No other systems

**Is required by:**
- [API Management](../api_management/)
- [Session Management](../session_management/)
- [Signature Management](../signature_management/)
- [Theme Management](../theme_management/)
- [User Management](../user_management/)

## Activation
This system is activated via the configurator:
```go
std.Must(cfg.SettingsManagement())
```
```go
settingsManagement := std.Must(cfg.SettingsManagement())
```