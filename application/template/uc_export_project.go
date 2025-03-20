package template

import (
	"archive/zip"
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"io"
	"os"
)

func NewExportZip(files blob.Store, repo Repository) ExportZip {
	return func(subject auth.Subject, pid ID, dst io.Writer) (err error) {
		if err := subject.AuditResource(repo.Name(), string(pid), PermExportZip); err != nil {
			return err
		}

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return err
		}

		if optPrj.IsNone() {
			return os.ErrNotExist
		}

		prj := optPrj.Unwrap()
		var srcFiles []File = prj.Files

		zipWriter := zip.NewWriter(dst)
		defer func() {
			if e := zipWriter.Close(); e != nil && err == nil {
				err = e
			}
		}()

		for _, f := range srcFiles {
			if err := addFileToZip(zipWriter, f, files); err != nil {
				return err
			}
		}

		return nil
	}
}

func addFileToZip(zipWriter *zip.Writer, file File, files blob.Store) error {
	optReader, err := files.NewReader(context.Background(), file.Blob)
	if err != nil {
		return fmt.Errorf("cannot open file from reader for %s: %w", file.Blob, err)
	}

	if optReader.IsNone() {
		// not sure, perhaps a kind of race on the blob level. lets do what we can and just ignore that
		// but we don't want a lock here, because a slow download or timeout of the writer
		// may block everything else
		return nil
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	writer, err := zipWriter.Create(file.Filename)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("cannot copy file entry for %s: %w", file.Filename, err)
	}

	return nil
}
