// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rest

import (
	"archive/zip"
	"fmt"
	"io"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
)

// ImportFromZip reads a ZIP archive from zipFile and writes every contained entry into the
// named target store.  The zip entry filenames are expected to be encoded with [EncodeZipKey];
// if decoding fails the raw filename is used as-is so that archives with arbitrary names are
// also accepted.
func ImportFromZip(store blob.Store, zipFile core.File) error {

	// Open the uploaded file for reading.
	rc, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("cannot open zip file: %w", err)
	}
	defer rc.Close()

	// archive/zip requires an io.ReaderAt; assert the reader accordingly.
	readerAt, ok := rc.(io.ReaderAt)
	if !ok {
		return fmt.Errorf("zip file reader does not implement io.ReaderAt; provide a seekable source such as a multipart upload or an os.File")
	}

	// archive/zip also requires the total size.
	size, hasSize := zipFile.Size()
	if !hasSize {
		return fmt.Errorf("zip file size is unknown; cannot open zip archive without a known size")
	}

	zr, err := zip.NewReader(readerAt, size)
	if err != nil {
		return fmt.Errorf("cannot parse zip archive: %w", err)
	}

	for _, f := range zr.File {
		// Skip directory entries – only blobs matter.
		if f.FileInfo().IsDir() {
			continue
		}

		// Try to reverse the percent-encoding applied by EncodeZipKey.
		// If the name was never encoded (or was encoded differently) use it as-is.
		key, decodeErr := DecodeZipKey(f.Name)
		if decodeErr != nil {
			key = f.Name
		}

		fr, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open zip entry %q: %w", f.Name, err)
		}

		_, writeErr := blob.Write(store, key, fr)
		_ = fr.Close()

		if writeErr != nil {
			return fmt.Errorf("cannot write blob %q to store %q: %w", key, store.Name(), writeErr)
		}
	}

	return nil
}
