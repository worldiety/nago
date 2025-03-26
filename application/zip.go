// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"archive/zip"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"iter"
	"os"
)

func makeZip(dstFile string, it iter.Seq2[core.File, error]) error {
	zipfile, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("cannot create zip file: %w", err)
	}
	defer zipfile.Close()

	zipWriter := zip.NewWriter(zipfile)
	defer zipWriter.Close()

	it(func(file core.File, e error) bool {
		if e != nil {
			err = e
			return false
		}

		header := &zip.FileHeader{
			Name:   file.Name(),
			Method: zip.Deflate,
			//UncompressedSize64: uint64(file.Size()),
		}

		writer, e := zipWriter.CreateHeader(header)
		if e != nil {
			err = e
			return false
		}

		_, e = file.Transfer(writer)

		if e != nil {
			err = e
		}

		return true
	})

	if err != nil {
		return fmt.Errorf("cannot walk src fs: %w", err)
	}

	return nil
}
