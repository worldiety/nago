// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewFindStagingByID(repo StagingRepository) FindStagingByID {
	return func(subject auth.Subject, id SID) (option.Opt[Staging], error) {
		optStaging, err := repo.FindByID(id)
		if err != nil {
			return option.Opt[Staging]{}, err
		}

		if optStaging.IsNone() {
			return option.Opt[Staging]{}, nil
		}

		staging := optStaging.Unwrap()

		if !subject.HasResourcePermission(rebac.Namespace(repo.Name()), rebac.Instance(staging.ID), PermFindStaging) && (staging.CreatedBy != subject.ID() || staging.CreatedBy == "") {
			return option.Opt[Staging]{}, user.PermissionDeniedError(PermFindStaging)
		}

		return optStaging, nil
	}
}
