// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// InferModulePath derives the module path of a tree of sources, e.g. an [embed.FS], from the imports of its
// packages: a program in the directory scene, which imports example.com/app/scene/messages, while the tree has the
// directory scene/messages, belongs to the module example.com/app. It returns false, if no package imports another
// one of the tree, and an error, if the imports contradict each other.
func InferModulePath(fsys fs.FS) (string, bool, error) {
	dirs := map[string]bool{}
	var files []string
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && path.Ext(p) == ".go" && !strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
			dirs[path.Dir(p)] = true
		}
		return nil
	})
	if err != nil {
		return "", false, err
	}

	module := ""
	for _, name := range files {
		src, err := fs.ReadFile(fsys, name)
		if err != nil {
			return "", false, err
		}
		f, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			for dir := range dirs {
				prefix, ok := strings.CutSuffix(p, "/"+dir)
				if !ok || dir == "." {
					continue
				}
				if module != "" && module != prefix {
					return "", false, fmt.Errorf("the imports of %s and of another file disagree about the module path: %s or %s", name, module, prefix)
				}
				module = prefix
			}
		}
	}
	return module, module != "", nil
}
