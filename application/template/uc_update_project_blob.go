package template

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"io"
	"os"
)

func NewUpdateProjectBlob(files blob.Store, repo Repository) UpdateProjectBlob {
	return func(subject auth.Subject, pid ID, file BlobID, value io.Reader) error {
		if err := subject.AuditResource(repo.Name(), string(pid), PermUpdateProjectBlob); err != nil {
			return err
		}

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return err
		}

		if optPrj.IsNone() {
			return nil
		}

		prj := optPrj.Unwrap()
		for _, f := range prj.Files {
			if f.Blob == file {
				wr, err := files.NewWriter(context.Background(), file)
				if err != nil {
					return err
				}

				defer wr.Close()

				_, err = io.Copy(wr, value)
				if err != nil {
					return err
				}

				return nil
			}
		}

		return os.ErrNotExist
	}
}
