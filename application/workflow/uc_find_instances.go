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
	"iter"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
)

func NewFindInstances(instanceStore blob.Store) FindInstances {
	return func(subject user.Subject, id ID) iter.Seq2[Instance, error] {
		return func(yield func(Instance, error) bool) {
			for key, err := range instanceStore.List(context.Background(), blob.ListOptions{
				Prefix: string(id) + "/",
			}) {
				if err != nil {
					if !subject.HasResourcePermission(instanceStore.Name(), key, PermFindInstances) {
						continue
					}
				}

				_, id, ok := InstanceKey(key).Split()
				if !ok {
					if !yield("", fmt.Errorf("invalid key format: %s", key)) {
						return
					}
				}

				if !yield(id, err) {
					return
				}
			}
		}
	}
}
