// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
)

func NewCreateInstance(instanceStore blob.Store) CreateInstance {
	return func(subject user.Subject, id ID) (Instance, error) {
		if err := subject.AuditResource(rebac.Namespace(instanceStore.Name()), rebac.Instance(id), PermCreateInstance); err != nil {
			return "", err
		}

		if id == "" {
			return "", fmt.Errorf("empty id")
		}

		if strings.Contains(string(id), "/") {
			return "", fmt.Errorf("id must not contain '/'")
		}

		instance := data.RandIdent[Instance]()
		iid := NewInstanceKey(id, instance)
		exists, err := instanceStore.Exists(context.Background(), string(iid))
		if err != nil {
			return "", err
		}

		// unlikely anyway...
		if exists {
			return "", os.ErrExist
		}

		exists, err = instanceStore.Exists(context.Background(), string(instance))
		if err != nil {
			return "", err
		}

		// unlikely anyway...
		if exists {
			return "", os.ErrExist
		}

		if err := blob.Put(instanceStore, string(iid), nil); err != nil {
			return "", err
		}

		if err := blob.Put(instanceStore, string(instance), nil); err != nil {
			return "", err
		}

		return instance, nil
	}
}
