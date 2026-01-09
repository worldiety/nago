// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package builder

type Form struct {
	Elements []Element `json:"elements"`
}

func (f *Form) AddTitle(t string) {
	f.Elements = append(f.Elements, Title{t})
}

type Title struct {
	Title string `json:"title"`
}

type Element interface {
}
