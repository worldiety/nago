package scheduler

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
)

func NewFindSettingsByID(repo SettingsRepository) FindSettingsByID {
	return func(subject auth.Subject, id ID) (std.Option[Settings], error) {
		if err := subject.Audit(PermFindSettingsByID); err != nil {
			return std.None[Settings](), err
		}

		return repo.FindByID(id)
	}
}
