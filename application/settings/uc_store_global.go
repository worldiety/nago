// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package settings

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"reflect"
)

func NewStoreGlobal(repo data.Repository[StoreBox[GlobalSettings], ID], bus events.Bus) StoreGlobal {
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

		err := repo.Save(StoreBox[GlobalSettings]{
			ID:       id,
			Settings: settings,
		})

		if err != nil {
			return err
		}

		bus.Publish(GlobalSettingsUpdated{Settings: settings})
		return nil
	}
}
