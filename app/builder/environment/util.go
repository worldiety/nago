// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package environment

import (
	"fmt"
	"os"
	"slices"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func getEnv(subject auth.Subject, repoEnv Repository, env ID) (Environment, error) {
	var zero Environment
	optEnv, err := repoEnv.FindByID(env)
	if err != nil {
		return zero, fmt.Errorf("failed to find env by id: %w", err)
	}

	if optEnv.IsNone() {
		return zero, fmt.Errorf("env id %v does not exist: %w", env, os.ErrNotExist)
	}

	uEnv := optEnv.Unwrap()
	if !slices.Contains(uEnv.Owner, subject.ID()) {
		return zero, user.PermissionDeniedErr
	}

	return uEnv, nil
}
