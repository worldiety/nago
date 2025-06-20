// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"sync"
	"time"
)

func NewCreateStaging(mutex *sync.Mutex, repoStaging StagingRepository) CreateStaging {
	return func(subject auth.Subject, cdata StagingCreationData) (Staging, error) {
		if err := subject.Audit(PermCreateStaging); err != nil {
			return Staging{}, err
		}

		mutex.Lock()
		defer mutex.Unlock()

		stage := Staging{
			ID:        data.RandIdent[SID](),
			CreatedAt: time.Now(),
			CreatedBy: subject.ID(),
			Name:      cdata.Name,
			Comment:   cdata.Comment,
			Importer:  cdata.Importer,
		}

		optStage, err := repoStaging.FindByID(stage.ID)
		if err != nil {
			return Staging{}, err
		}

		if optStage.IsSome() {
			return stage, fmt.Errorf("stage already exists")
		}

		return stage, repoStaging.Save(stage)
	}
}
