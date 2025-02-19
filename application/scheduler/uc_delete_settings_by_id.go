package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewDeleteSettingsByID(repo SettingsRepository) DeleteSettingsByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteSettingsByID); err != nil {
			return err
		}

		return repo.DeleteByID(id)
	}
}
