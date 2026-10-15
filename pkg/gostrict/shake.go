// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"regexp"
	"strings"
	"sync"
)

// The runtime is split into its top level declarations, so that a bundle contains only those, which it uses.

type runtimeDecl struct {
	name string
	code string
	deps []string
}

var (
	runtimeOnce  sync.Once
	runtimeDecls []runtimeDecl
	runtimeIndex map[string]int
)

var (
	declStart = regexp.MustCompile(`^(?:function|const|class) (\$[A-Za-z0-9_]*)`)
	dollarRef = regexp.MustCompile(`(?:^|[^A-Za-z0-9_$])(\$[A-Za-z_][A-Za-z0-9_]*)`)
)

func parseRuntime() {
	runtimeIndex = map[string]int{}
	var cur *runtimeDecl
	var comment []string
	for _, line := range strings.Split(runtimeJS, "\n") {
		if m := declStart.FindStringSubmatch(line); m != nil {
			runtimeDecls = append(runtimeDecls, runtimeDecl{name: m[1]})
			cur = &runtimeDecls[len(runtimeDecls)-1]
			runtimeIndex[cur.name] = len(runtimeDecls) - 1
			comment = nil
		}
		if cur == nil {
			continue
		}
		if strings.HasPrefix(line, "//") {
			comment = append(comment, line)
			continue
		}
		if len(comment) > 0 {
			cur.code += strings.Join(comment, "\n") + "\n"
			comment = nil
		}
		cur.code += line + "\n"
	}

	for i := range runtimeDecls {
		for _, m := range dollarRef.FindAllStringSubmatch(runtimeDecls[i].code, -1) {
			if m[1] != runtimeDecls[i].name {
				runtimeDecls[i].deps = append(runtimeDecls[i].deps, m[1])
			}
		}
	}
}

// usedRuntime returns the declarations of the runtime, which the code refers to, with their dependencies, in the
// order of the runtime, so that a constant is defined before another one uses it.
func usedRuntime(code ...string) string {
	// the variants of unions at the boundary are boxed
	code = append(code, "$box")
	runtimeOnce.Do(parseRuntime)

	used := map[int]bool{}
	var visit func(name string)
	visit = func(name string) {
		i, ok := runtimeIndex[name]
		if !ok || used[i] {
			return
		}
		used[i] = true
		for _, d := range runtimeDecls[i].deps {
			visit(d)
		}
	}
	for _, c := range code {
		for _, m := range dollarRef.FindAllStringSubmatch(c, -1) {
			visit(m[1])
		}
	}

	var sb strings.Builder
	for i, d := range runtimeDecls {
		if used[i] {
			sb.WriteString(d.code)
		}
	}
	return sb.String()
}
