// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package stdvalues uses functions and methods of the standard packages as values, e.g. s := list.SortFunc[int32].
// The checker either accepts such a program, then it translates and runs natively, in goja and in the vm alike, or
// rejects it with a finding of its own. Each other file is a program of its own together with this one. The other
// files are excluded from the subset by their build constraint, so that the package as a whole is checked without
// them, whichever way the checker decides; the tests load each with the constraint removed.
package stdvalues
