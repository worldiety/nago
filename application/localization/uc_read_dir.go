// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"maps"
	"slices"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
)

func NewReadDir(resources *i18n.Resources) ReadDir {
	return func(subject auth.Subject, path Path) (Directory, error) {
		if err := subject.Audit(PermReadDir); err != nil {
			return Directory{}, err
		}

		if path == "." {
			path = ""
		}

		keys := resources.SortedKeys()
		if path != "" {
			keys = xslices.PrefixSearch(keys, i18n.Key(path))
		}

		tmp := map[string]DirInfo{}
		var res Directory
		// keys now contains the subset of all recursively contained children.
		// this gets o(log(n)) cheaper, the deeper the path query is.
		for _, key := range keys {
			if key.StringKey() {
				continue
			}
			trimmedKey := i18n.Key(strings.TrimPrefix(string(key), string(path)))
			trimmedKeyDirs := trimmedKey.Directories()
			leafEntry := len(trimmedKeyDirs) == 0
			if leafEntry {
				// so this is like 1.leaf
				res.Strings = append(res.Strings, key)
				continue
			}

			// this is like 1.2.leaf or 1.2.3.leaf
			dirInfo := tmp[trimmedKeyDirs[0]]
			if len(dirInfo.Name) == 0 {
				dirInfo.Name = NormalizeAndTitle(trimmedKeyDirs[0])
				dirInfo.Path = xstrings.Join2(".", path, Path(trimmedKeyDirs[0]))
			}

			dirInfo.TotalKeys++

			for _, bnd := range resources.All() {
				if bnd.MessageTypeByKey(key) == i18n.MessageUndefined {
					dirInfo.TotalMissingKeys++
				}
			}

			tmp[trimmedKeyDirs[0]] = dirInfo
		}

		for _, key := range slices.Sorted(maps.Keys(tmp)) {
			res.Directories = append(res.Directories, tmp[key])
		}

		return res, nil
	}
}
