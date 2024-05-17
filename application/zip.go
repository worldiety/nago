package application

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
)

func makeZip(dstFile string, src fs.FS) error {
	zipfile, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("cannot create zip file: %w", err)
	}
	defer zipfile.Close()

	zipWriter := zip.NewWriter(zipfile)
	defer zipWriter.Close()

	err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := src.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot walk src fs: %w", err)
	}

	return nil
}
