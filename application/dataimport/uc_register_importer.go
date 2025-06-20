// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"fmt"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewRegisterImporter(importers *concurrent.RWMap[importer.ID, importer.Importer]) RegisterImporter {
	return func(subject auth.Subject, imp importer.Importer) error {
		if err := subject.Audit(PermRegisterParser); err != nil {
			return err
		}

		if imp.Identity() == "" {
			return fmt.Errorf("invalid importer identity")
		}

		if imp.Configuration().Name == "" {
			return fmt.Errorf("invalid importer name")
		}

		importers.Put(imp.Identity(), imp)

		return nil
	}
}
