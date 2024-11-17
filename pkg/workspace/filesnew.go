package workspace

import (
	"context"
	"fmt"
	"go.wdy.de/nago/annotation"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"io"
	"slices"
)

var permSaveFile = annotation.Permission[SaveFile]("de.worldiety.nago.workspace.files.save")

type SaveFile func(subject auth.Subject, id ID, name Filename, mimetype Mimetype, data io.Reader) (File, error)

func NewCreateFile(repo Repository, files blob.Store) SaveFile {
	return func(subject auth.Subject, id ID, name Filename, mimetype Mimetype, dat io.Reader) (File, error) {
		if err := subject.Audit(permSaveFile.Identity()); err != nil {
			return File{}, err
		}

		optWS, err := repo.FindByID(id)
		if err != nil {
			return File{}, fmt.Errorf("cannot load workspace: %w", err)
		}

		if optWS.IsNone() {
			return File{}, fmt.Errorf("no such workspace: %w", err)
		}

		ws := optWS.Unwrap()

		file := File{
			Name:     name,
			Mimetype: mimetype,
			Ref:      data.RandIdent[BlobID](),
		}

		// save new blob
		n, err := blob.Write(files, string(file.Ref), dat)
		if err != nil {
			return File{}, err
		}

		file.Size = n

		// delete from slice
		var updatedFile File
		ws.Files = slices.DeleteFunc(ws.Files, func(file File) bool {
			if file.Name == name {
				updatedFile = file
				return true
			}

			return false
		})

		// cleanup blob store
		if updatedFile.Name != "" {
			if err := files.Delete(context.Background(), string(updatedFile.Ref)); err != nil {
				return File{}, err
			}
		}

		// update workspace
		ws.Files = append(ws.Files, file)

		if err := repo.Save(ws); err != nil {
			return file, err
		}

		return file, nil
	}
}
