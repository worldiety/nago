// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type Package struct {
	ID          PackageID
	Types       *Types
	ImportPath  ImportPath
	Name        Ident
	Description string
}

func NewPackage(id PackageID, path ImportPath, name Ident) *Package {
	return &Package{
		ID:         id,
		Types:      NewTypes(),
		ImportPath: path,
		Name:       name,
	}
}

func (p *Package) Clone() *Package {
	return &Package{
		ID:          p.ID,
		Types:       p.Types.Clone(),
		ImportPath:  p.ImportPath,
		Name:        p.Name,
		Description: p.Description,
	}
}

func (p *Package) String() string {
	return string(p.ImportPath)
}

func (p *Package) Identity() PackageID {
	return p.ID
}
