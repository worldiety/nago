package settings

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"reflect"
)

func NewLoadGlobal(repo data.Repository[StoreBox[GlobalSettings], ID]) LoadGlobal {
	return func(subject permission.Auditable, t reflect.Type) (GlobalSettings, error) {
		var zero GlobalSettings
		if t.Kind() != reflect.Struct {
			return zero, fmt.Errorf(`type must be struct type`)
		}

		zero = reflect.New(t).Elem().Interface().(GlobalSettings)

		if err := subject.Audit(PermLoadGlobal); err != nil {
			return zero, err
		}

		id := makeGlobalID(t)
		optBox, err := repo.FindByID(id)
		if err != nil {
			return zero, err
		}

		if optBox.IsNone() {
			return zero, nil
		}

		return optBox.Unwrap().Settings, nil
	}
}

func makeGlobalID(t reflect.Type) ID {
	return ID(fmt.Sprintf("%s.%s", t.PkgPath(), t.Name()))
}
