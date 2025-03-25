package theme

import (
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
)

func NewUpdateColors(bus events.Bus, loadGlobal settings.LoadGlobal, storeGlobal settings.StoreGlobal) UpdateColors {
	return func(subject auth.Subject, colors Colors) error {
		if err := subject.Audit(PermUpdateColors); err != nil {
			return err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)
		cfg.Colors = colors
		err := storeGlobal(user.SU(), cfg)
		if err != nil {
			return err
		}

		bus.Publish(SettingsUpdated{
			Settings: cfg,
		})
		
		return nil
	}
}
