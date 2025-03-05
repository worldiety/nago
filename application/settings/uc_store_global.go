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
		return repo.Save(StoreBox[GlobalSettings]{
			ID:       id,
			Settings: settings,
		})
	}
}
