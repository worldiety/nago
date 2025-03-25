package settings

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"reflect"
)

func NewStoreGlobal(repo data.Repository[StoreBox[GlobalSettings], ID]) StoreGlobal {
	return func(subject permission.Auditable, settings GlobalSettings) error {
		if err := subject.Audit(PermStoreGlobal); err != nil {
			return err
		}

		if settings == nil {
			return fmt.Errorf("settings must not be nil")
		}

		id := makeGlobalID(reflect.TypeOf(settings))

		// only issue a write, if the settings are actually different
		// permanent writes are costly for us and this may simplify things for developers
		optBox, _ := repo.FindByID(id)
		if optBox.IsSome() {
			s := optBox.Unwrap()
			if reflect.DeepEqual(s.Settings, settings) {
				return nil
			}
		}

		return repo.Save(StoreBox[GlobalSettings]{
			ID:       id,
			Settings: settings,
		})
	}
}
