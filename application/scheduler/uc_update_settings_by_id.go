package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewUpdateSettings(repo SettingsRepository) UpdateSettings {
	return func(subject auth.Subject, settings Settings) error {
		if err := subject.Audit(PermUpdateSettingsByID); err != nil {
			return err
		}

		return repo.Save(settings)
	}
}
