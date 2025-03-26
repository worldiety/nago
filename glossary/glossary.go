// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package glossary

import (
	"fmt"
	"go.wdy.de/nago/glossary/docm"
	"go.wdy.de/nago/glossary/docm/markdown"
	"go.wdy.de/nago/pkg/xreflect"
	"log/slog"
	"runtime/debug"
	"strings"
)

// Auto creates a glossary based on conventions and any embedded source code using the [xreflect.Import] package.
func Auto() *docm.Document {
	info, ok := debug.ReadBuildInfo()
	var mainPkg *xreflect.Package
	if ok {
		pkg, _ := xreflect.PackageOf(info.Path)
		mainPkg = pkg
	}

	doc := &docm.Document{}
	var content docm.Sequence

	if mainPkg != nil {
		content = append(content, introBasedOnLauncherPkg(mainPkg))
	}

	for _, pkg := range xreflect.Packages() {
		if pkg.Name() == "main" {
			continue
		}

		pkgAlias, pkgDoc := nicePkgDoc(pkg)

		content.Add(&docm.Heading{Level: 2, Body: &docm.Text{Value: pkgAlias}})
		content.Add(markdown.Parse(pkgDoc))
		types := pkg.Types()
		for _, t := range types {
			alias, doc := niceDoc(t)
			content.Add(&docm.Heading{Level: 3, Body: &docm.Text{Value: alias}})
			content.Add(markdown.Parse(doc))

			switch t := t.(type) {
			case *xreflect.Struct:
				withComments, withoutComments := splitFields(t.Fields())
				for _, field := range withComments {
					alias, doc := niceDoc(field)

					content.Add(&docm.Heading{Level: 4, Body: &docm.Text{Value: alias}})
					content.Add(markdown.Parse(doc))
				}

				if len(withoutComments) > 0 {
					if len(withComments) == 1 {
						content.Add(&docm.Text{Value: "Eine weitere Eigenschaft ist " + withoutComments[0] + "."})
					} else {
						content.Add(&docm.Text{Value: "Weitere Eigenschaften sind " + strings.Join(withoutComments, ", ") + "."})
					}
				}
			}
		}
	}
	doc.Body = content
	return doc
}

func splitFields(fields []*xreflect.Field) (withComment []*xreflect.Field, withoutComments []string) {
	for _, field := range fields {
		alias, doc := niceDoc(field)
		if doc == "" {
			withoutComments = append(withoutComments, alias)
		} else {
			withComment = append(withComment, field)
		}
	}

	return
}

func introBasedOnLauncherPkg(mainPkg *xreflect.Package) docm.Element {
	return markdown.Parse(mainPkg.Doc())
}

func niceDocGeneric(pre string, typ interface {
	Name() string
	Doc() string
}) (alias, doc string) {
	doc = typ.Doc()
	prefix := fmt.Sprintf("%s%s oder ", pre, typ.Name())
	if strings.HasPrefix(doc, prefix) {
		idx := strings.Index(doc, ".")
		if idx < 0 {
			slog.Error("invalid alias definition for type", "type", typ.Name())
			return typ.Name(), doc
		}
		return doc[len(prefix):idx], doc[idx+1:]
	}

	return typ.Name(), doc
}

func niceDoc(typ interface {
	Name() string
	Doc() string
}) (alias, doc string) {
	return niceDocGeneric("", typ)
}

func nicePkgDoc(typ *xreflect.Package) (alias, doc string) {
	return niceDocGeneric("Package ", typ)
}
