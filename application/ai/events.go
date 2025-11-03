// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

// LibraryUpdateRequested should be issued by anyone who has modified the denoted Store and Key combination.
// If a synchronization process has been defined between the Store and a [library.Library], the insert, update or
// delete mechanics are triggered faster and more specifically. However, you can never expect this to happen
// immediately.
type LibraryUpdateRequested struct {
	Store string // name of the affected store
	Key   string
}
