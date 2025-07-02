// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func Parse(str string) (Version, bool) {
	if !strings.HasPrefix(str, "v") {
		str = "v" + str
	}

	v, ok := parse(str)
	if !ok {
		return Version{}, false
	}

	var res Version
	if iv, err := strconv.Atoi(v.major); err == nil {
		res.Major = iv
	} else {
		return res, false
	}

	if iv, err := strconv.Atoi(v.minor); err == nil {
		res.Minor = iv
	} else {
		return res, false
	}

	if iv, err := strconv.Atoi(v.patch); err == nil {
		res.Patch = iv
	} else {
		return res, false
	}

	return res, true
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) Newer(other Version) bool {
	return Compare(v, other) > 0
}

func Compare(a, b Version) int {
	if a.Major != b.Major {
		return a.Major - b.Major
	}

	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}

	if a.Patch != b.Patch {
		return a.Patch - b.Patch
	}

	return 0
}
