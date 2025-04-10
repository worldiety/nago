// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package settings

// GlobalSettingsUpdated is triggered whenever new settings have been stored. Note, that this is a rather uncommon
// case and happens usually ever in the setup phase of a project or when tweaking something at runtime.
type GlobalSettingsUpdated struct {
	Settings any
}
