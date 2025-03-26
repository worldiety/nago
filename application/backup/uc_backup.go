// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package backup

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"io"
	"log/slog"
	"time"
)

func NewBackup(src Persistence) Backup {
	return func(ctx context.Context, subject auth.Subject, dst io.Writer) (err error) {
		if err := subject.Audit(PermBackup); err != nil {
			return err
		}

		slog.Info("backup", "action", "started", "userID", subject.ID(), "userEMail", subject.Email())

		zipWriter := zip.NewWriter(dst)
		defer func() {
			if e := zipWriter.Close(); e != nil && err == nil {
				err = e
			}
		}()

		index := &Index{
			CreatedAt: time.Now(),
		}

		fileCounter := 0
		for fileStoreName, err := range src.FileStores() {
			if err != nil {
				return err
			}

			slog.Info("backup", "action", "large blob store started", "store", fileStoreName)

			storeIdx := Store{
				Name:       fileStoreName,
				Stereotype: StereotypeBlob,
			}

			store, err := src.FileStore(fileStoreName)
			if err != nil {
				return fmt.Errorf("cannot open file store %s: %w", fileStoreName, err)
			}

			if err := addFilesFromStoreToZipStream(&storeIdx, &fileCounter, ctx, zipWriter, store); err != nil {
				return fmt.Errorf("cannot add file store files to zip stream: %w", err)
			}

			index.Stores = append(index.Stores, storeIdx)

			slog.Info("backup", "action", "large blob store completed", "store", fileStoreName, "blobs", len(storeIdx.Blobs))
		}

		for entityStoreName, err := range src.EntityStores() {
			if err != nil {
				return err
			}

			storeIdx := Store{
				Name:       entityStoreName,
				Stereotype: StereotypeDocument,
			}

			slog.Info("backup", "action", "document store started", "store", entityStoreName)

			store, err := src.EntityStore(entityStoreName)
			if err != nil {
				return fmt.Errorf("cannot open entity store %s: %w", entityStoreName, err)
			}

			if err := addFilesFromStoreToZipStream(&storeIdx, &fileCounter, ctx, zipWriter, store); err != nil {
				return fmt.Errorf("cannot add file store files to zip stream: %w", err)
			}

			index.Stores = append(index.Stores, storeIdx)

			slog.Info("backup", "action", "document store completed", "store", entityStoreName, "blobs", len(storeIdx.Blobs))
		}

		// write backup index
		bufIdx, err := json.Marshal(index)
		if err != nil {
			return fmt.Errorf("cannot marshal index: %w", err)
		}

		writer, err := zipWriter.Create("index.json")
		if err != nil {
			return fmt.Errorf("cannot create index file: %w", err)
		}

		if _, err := writer.Write(bufIdx); err != nil {
			return fmt.Errorf("cannot write index: %w", err)
		}

		slog.Info("backup", "action", "complete", "objects", fileCounter, "size", index.Size())

		return nil
	}
}

func addFilesFromStoreToZipStream(idx *Store, fileCounter *int, ctx context.Context, zipWriter *zip.Writer, store blob.Store) error {
	for name, err := range store.List(ctx, blob.ListOptions{}) {
		if err != nil {
			return fmt.Errorf("cannot list %s: %w", name, err)
		}

		optR, err := store.NewReader(ctx, name)
		if err != nil {
			return fmt.Errorf("cannot read %s.%s: %w", name, name, err)
		}

		if optR.IsNone() {
			// may be normal, due to concurrency
			continue
		}

		*fileCounter++

		reader := optR.Unwrap()
		if err := addFileToZip(idx, *fileCounter, zipWriter, name, reader); err != nil {
			_ = reader.Close()
			return fmt.Errorf("cannot add %s.%s to zip: %w", name, name, err)
		}

		if err := reader.Close(); err != nil {
			return fmt.Errorf("cannot close %s: %w", name, err)
		}
	}

	return nil
}

func addFileToZip(idx *Store, fileCounter int, zipWriter *zip.Writer, name string, reader io.Reader) error {
	path := fmt.Sprintf("data/%d", fileCounter)
	writer, err := zipWriter.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create file entry for %s: %w", name, err)
	}

	hash := sha256.New()
	tee := io.TeeReader(reader, hash)

	n, err := io.Copy(writer, tee)
	if err != nil {
		return fmt.Errorf("cannot copy file entry for %s: %w", name, err)
	}

	hStr := hex.EncodeToString(hash.Sum(nil))

	// we will not deduplicate stored data here, because this means at first we need to read everything twice
	// and more importantly that hash sum is not stable, because we cannot backup a consistent state.
	idx.Blobs = append(idx.Blobs, Blob{
		ID:     name,
		Size:   n,
		Sha256: hStr,
		Path:   path,
	})

	return nil
}
