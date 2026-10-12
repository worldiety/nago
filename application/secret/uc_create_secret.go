// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewCreateSecret(mutex *sync.Mutex, bus events.Bus, secrets Repository) CreateSecret {
	return func(subject auth.Subject, credentials Credentials, opts ...CreateOptions) (ID, error) {
		if err := subject.Audit(PermCreateSecret); err != nil {
			return "", err
		}

		owners := []user.ID{subject.ID()}
		var groups []group.ID
		for _, opt := range opts {
			if len(opt.Groups) > 0 {
				if err := subject.Audit(PermUpdateMySecretGroups); err != nil {
					return "", err
				}
			}

			for _, gid := range opt.Groups {
				if !subject.HasGroup(gid) {
					return "", std.NewLocalizedError("Secret nicht angelegt", fmt.Sprintf("Die Gruppe '%v' kann nicht hinzugefügt werden, da das Konto selbst nicht zu der Gruppe gehört.", gid))
				}

				if !slices.Contains(groups, gid) {
					groups = append(groups, gid)
				}
			}

			if len(opt.Owners) > 0 {
				if err := subject.Audit(PermUpdateMySecretOwners); err != nil {
					return "", err
				}
			}

			for _, uid := range opt.Owners {
				if !slices.Contains(owners, uid) {
					owners = append(owners, uid)
				}
			}
		}

		mutex.Lock()
		defer mutex.Unlock()

		secret := Secret{
			ID:          data.RandIdent[ID](),
			Owners:      owners,
			Groups:      groups,
			LastMod:     time.Now(),
			Credentials: credentials,
		}

		optSec, err := secrets.FindByID(secret.ID)
		if err != nil {
			return "", fmt.Errorf("cannot find secret: %v", err)
		}

		if optSec.IsSome() {
			return "", fmt.Errorf("secret already exists")
		}

		if err := secrets.Save(secret); err != nil {
			return "", fmt.Errorf("cannot save secret: %v", err)
		}

		bus.Publish(Created{Secret: secret.ID})

		return secret.ID, nil
	}
}
