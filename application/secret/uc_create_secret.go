// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"fmt"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"

	"sync"
	"time"
)

func NewCreateSecret(mutex *sync.Mutex, bus events.Bus, secrets Repository) CreateSecret {
	return func(subject auth.Subject, credentials Credentials) (ID, error) {
		if err := subject.Audit(PermCreateSecret); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		secret := Secret{
			ID:          data.RandIdent[ID](),
			Owners:      []user.ID{subject.ID()},
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
