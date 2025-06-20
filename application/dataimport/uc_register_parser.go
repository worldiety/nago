// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"fmt"
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewRegisterParser(parsers *concurrent.RWMap[parser.ID, parser.Parser]) RegisterParser {
	return func(subject auth.Subject, p parser.Parser) error {
		if err := subject.Audit(PermRegisterParser); err != nil {
			return err
		}

		if p.Identity() == "" {
			return fmt.Errorf("invalid parser identity")
		}

		if p.Configuration().Name == "" {
			return fmt.Errorf("invalid parser name")
		}

		parsers.Put(p.Identity(), p)

		return nil
	}
}
