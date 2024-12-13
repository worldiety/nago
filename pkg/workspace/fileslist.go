package workspace

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"slices"
	"strings"
)

var permListFiles = annotation.Permission[ListFiles]("de.worldiety.nago.workspace.files.list")

type ListFiles func(subject auth.Subject, id ID) ([]File, error)

func NewListFiles(repo Repository) ListFiles {
	return func(subject auth.Subject, id ID) ([]File, error) {
		if err := subject.Audit(permListFiles.Identity()); err != nil {
			return nil, err
		}

		optWS, err := repo.FindByID(id)
		if err != nil {
			return nil, fmt.Errorf("cannot load workspace: %w", err)
		}

		if optWS.IsNone() {
			return nil, nil
		}

		files := optWS.Unwrap().Files
		slices.SortedFunc(slices.Values(files), func(e File, e2 File) int {
			return strings.Compare(string(e.Name), string(e2.Name))
		})

		return files, nil
	}
}
