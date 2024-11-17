package workspace

import (
	"context"
	"fmt"
	"go.wdy.de/nago/annotation"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"slices"
)

var permDeleteFiles = annotation.Permission[DeleteFiles]("de.worldiety.nago.workspace.files.delete")

// DeleteFiles removes all given files at once.
type DeleteFiles func(subject auth.Subject, id ID, files ...Filename) error

func NewDeleteFiles(repo Repository, blobs blob.Store) DeleteFiles {
	return func(subject auth.Subject, id ID, files ...Filename) error {
		if err := subject.Audit(permDeleteFiles.Identity()); err != nil {
			return err
		}

		optWS, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot load workspace: %w", err)
		}

		if optWS.IsNone() {
			return nil
		}

		var removed []File
		ws := optWS.Unwrap()
		ws.Files = slices.DeleteFunc(ws.Files, func(file File) bool {
			for _, filename := range files {
				if file.Name == filename {
					removed = append(removed, file)
					return true
				}
			}

			return false
		})

		for _, file := range removed {
			if err := blobs.Delete(context.Background(), string(file.Ref)); err != nil {
				return fmt.Errorf("cannot delete file %s@%s: %w", file.Name, file.Ref, err)
			}
		}

		if err := repo.Save(ws); err != nil {
			return fmt.Errorf("cannot save workspace: %w", err)
		}

		return nil
	}
}
