// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package speclink

import (
	"fmt"
	"io/fs"
	"strings"

	"go.wdy.de/nago/auth"
)

// FindSourceDocument reads one of the documents the requirements were derived from.
//
// # Why this is worth its own use case
//
// Because it is often the better answer to "warum". A requirement is a sentence distilled by somebody from
// what was asked for; the source document is what the person asking actually wrote. Where the two differ,
// the second is the one that explains the first.
//
// The path comes from [Requirement.Sources], so a caller reaches this by having read a requirement rather
// than by knowing the repository layout.
type FindSourceDocument func(subject auth.Subject, path string) (string, error)

// NewFindSourceDocument builds the reader over the documents the application embedded.
//
// A nil filesystem is not an error: an application that embeds no sources simply has no documents, and the
// use case says so rather than being absent, so a caller does not have to test for a missing capability.
func NewFindSourceDocument(sources fs.FS) FindSourceDocument {
	return func(subject auth.Subject, path string) (string, error) {
		if err := subject.Audit(PermFindSourceDocument); err != nil {
			return "", err
		}

		if sources == nil {
			return "", fmt.Errorf("diese Anwendung hat keine Quelldokumente eingebettet")
		}

		// Sources are recorded as "path#anchor"; the anchor addresses a heading inside the document and is
		// not part of the file name.
		clean := path
		if i := strings.IndexByte(clean, '#'); i >= 0 {
			clean = clean[:i]
		}

		clean = strings.TrimPrefix(strings.TrimSpace(clean), "./")

		// fs.ValidPath rejects absolute paths, "..", and anything else that would escape the embedded tree.
		// The caller is frequently a language model repeating a path back, so this is not paranoia about an
		// attacker but about a plausible mistake.
		if !fs.ValidPath(clean) {
			return "", fmt.Errorf("ungültiger Pfad %q", path)
		}

		buf, err := fs.ReadFile(sources, clean)
		if err != nil {
			return "", fmt.Errorf("kein Quelldokument %q; die Pfade stehen im Feld sources einer Anforderung", clean)
		}

		return string(buf), nil
	}
}
