// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"os"
	"reflect"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewAnalyze(declarations *concurrent.RWMap[ID, *workflow]) Analyze {
	return func(subject user.Subject, id ID) (Analyzed, error) {
		if err := subject.AuditResource(RepositoryNameDeclaredWorkflows, rebac.Instance(id), PermAnalyze); err != nil {
			return Analyzed{}, err
		}

		wf, ok := declarations.Get(id)
		if !ok {
			return Analyzed{}, os.ErrNotExist
		}

		res := Analyzed{
			EventTypesByName: map[Typename]reflect.Type{},
		}

		for _, r := range wf.uniqueEventTypes() {
			res.EventTypesByName[NewTypename(r)] = r
			if wf.globalEvent(r) {
				res.StartEvents = append(res.StartEvents, NewTypename(r))
			}
		}

		return res, nil
	}
}
