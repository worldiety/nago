// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac_test

import (
	"testing"

	"go.wdy.de/nago/application/rebac"
)

func TestSelect(t *testing.T) {
	query := rebac.Select().
		Where().Source().IsNamespace("test").
		Where().Source().IsInstance("1234").
		Where().Relation().Has(rebac.Member).
		Where().Target().IsGlobal().
		Where().Target().IsAny()

	query = rebac.Select().
		Where().Source().Is("test", "1234").
		Where().Relation().IsAny().
		Where().Target().IsGlobal()

	_ = query
}
