// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewUpdateStagingTransformation(mutex *sync.Mutex, repo StagingRepository) UpdateStagingTransformation {
	return func(subject auth.Subject, stage SID, transform Transformation) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(stage), PermUpdateStagingTransformation); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optStage, err := repo.FindByID(stage)
		if err != nil {
			return err
		}

		if optStage.IsNone() {
			return fmt.Errorf("stage %s not found: %w", stage, os.ErrNotExist)
		}

		staging := optStage.Unwrap()
		staging.Transformation = transform
		return repo.Save(staging)
	}
}
