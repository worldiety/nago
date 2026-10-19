// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !vmstats

package vmcode

// statsOn counts the executed instructions, see stats_on.go.
const statsOn = false

func countOp(prev, cur op) {}
