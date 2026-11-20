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
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// NewRestore returns a [Restore] into dst. Before it clears the first store, it invalidates every derived store of dst,
// see [blob.Invalidate], and it restores none of them, should the backup contain one. At its end, whether it succeeded or
// failed, it publishes [Restored] on bus, which may be nil, and lets the owners of the derived stores rebuild them.
func NewRestore(dst blob.Stores, bus events.Bus) Restore {
	return func(ctx context.Context, subject auth.Subject, src io.Reader) error {
		if err := subject.Audit(PermRestore); err != nil {
			return err
		}

		if bus != nil {
			defer bus.Publish(Restored{})
		}

		slog.Info("restore", "action", "started", "userID", subject.ID(), "userEMail", subject.Email())

		// for security, we need to consume and unpack the entire zip file at first, otherwise we
		// may get interrupted in the middle of a restore leaving the system in a broken state, e.g.
		// without a valid user table. This happens often for large backups when uploaded.
		tempFilename := filepath.Join(os.TempDir(), data.RandIdent[string]()+".restore.zip")
		tmpFile, err := os.OpenFile(tempFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("could not create temp file: %w", err)
		}

		defer func() {
			_ = tmpFile.Close()
			if err := os.Remove(tempFilename); err != nil {
				slog.Error("cannot remove temp zip file", "err", err)
			}
		}()

		slog.Info("restore", "action", "receiving zip file")
		srcSize, err := io.Copy(tmpFile, src)
		if err != nil {
			return fmt.Errorf("could not copy to temp file: %w", err)
		}

		if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("could not seek temp file: %w", err)
		}

		slog.Info("restore", "action", "copied zip file into temporary file", "file", tempFilename, "size", srcSize)
		// copy is complete, now let us verify the zip file for any obvious corruptions
		zipReader, err := zip.NewReader(tmpFile, srcSize)
		if err != nil {
			return fmt.Errorf("could not open zip reader: %w", err)
		}

		// read index
		index, err := decodeIndex(zipReader)
		if err != nil {
			return fmt.Errorf("could not decode index: %w", err)
		}

		// validate that we have all files available and that they have not seen bitrot
		slog.Info("restore", "action", "validation started")

		totalBytes := int64(0)
		for _, store := range index.Stores {
			for _, file := range store.Blobs {
				r, err := zipReader.Open(file.Path)
				if err != nil {
					return fmt.Errorf("could not open zip file entry %s: %w", file.Path, err)
				}

				hasher := sha256.New()

				n, err := io.Copy(hasher, r)
				if err != nil {
					return fmt.Errorf("reading zip file entry %s failed: %w", file.Path, err)
				}

				if n != file.Size {
					return fmt.Errorf("checking zip file entry %s size failed: expected %d, got %d file size", file.Path, file.Size, n)
				}

				hash := hex.EncodeToString(hasher.Sum(nil))
				if hash != file.Sha256 {
					return fmt.Errorf("invalid checksum %s", file.Path)
				}

				totalBytes += file.Size
			}
		}

		slog.Info("restore", "action", "validation successful", "totalBytes", totalBytes, "size", index.Size(), "count", index.Count())

		// The derived stores do not fit the restored data, and a crash at any later point must leave them invalid. Their
		// owners rebuild them, after the restore has ended.
		derived, endInvalidation, err := invalidateDerived(dst)
		defer endInvalidation()
		if err != nil {
			return err
		}

		// check is complete now, let us destroy some bytes...
		for _, idxStore := range index.Stores {
			if derived[idxStore.Name] {
				slog.Info("restore", "action", "derived store left out", "name", idxStore.Name)
				continue
			}

			slog.Info("restore", "action", "clearing store", "name", idxStore.Name, "stereotype", idxStore.Stereotype)
			var store blob.Store
			switch idxStore.Stereotype {
			case StereotypeDocument:
				store, err = dst.Open(idxStore.Name, blob.OpenStoreOptions{Type: blob.EntityStore})
				if err != nil {
					return fmt.Errorf("could not create entity store %s: %w", idxStore.Name, err)
				}

			case StereotypeBlob:
				store, err = dst.Open(idxStore.Name, blob.OpenStoreOptions{Type: blob.FileStore})
				if err != nil {
					return fmt.Errorf("could not create blob store %s: %w", idxStore.Name, err)
				}

			default:
				return fmt.Errorf("unknown stereotype %s", idxStore.Stereotype)
			}

			if err := blob.DeleteAll(store); err != nil {
				return fmt.Errorf("could not clear store %s: %w", idxStore.Name, err)
			}

			slog.Info("restore", "action", "clearing complete", "name", idxStore.Name)

			slog.Info("restore", "action", "copy new data", "name", idxStore.Name)

			for _, b := range idxStore.Blobs {
				cancelWriteCtx, cancelWrite := context.WithCancel(ctx)
				writer, err := store.NewWriter(cancelWriteCtx, b.ID)
				if err != nil {
					cancelWrite()
					return fmt.Errorf("could not create new blob writer in store %s: %w", idxStore.Name, err)
				}

				reader, err := zipReader.Open(b.Path)
				if err != nil {
					cancelWrite()
					_ = writer.Close()
					return fmt.Errorf("could not open zip file entry %s: %w", b.Path, err)
				}

				if _, err := io.Copy(writer, reader); err != nil {
					cancelWrite()
					_ = writer.Close()
					_ = reader.Close()
					return fmt.Errorf("could not restore blob %s: %w", b.Path, err)
				}

				_ = reader.Close()

				if err := writer.Close(); err != nil {
					cancelWrite()
					return fmt.Errorf("could not close writer: %w", err)
				}

				// need to free resources, even though writer has closed successfully
				cancelWrite()
			}

			slog.Info("restore", "action", "complete", "name", idxStore.Name, "blobs", len(idxStore.Blobs))
		}

		slog.Info("restore", "action", "complete", "totalBytes", totalBytes, "size", index.Size())

		return nil
	}
}

// invalidateDerived invalidates every derived store of dst and returns their names. end ends the invalidations, also
// after an error.
func invalidateDerived(dst blob.Stores) (derived map[string]bool, end func(), err error) {
	derived = map[string]bool{}
	var ends []func()
	end = func() {
		for _, fn := range ends {
			fn()
		}
	}

	for name, err := range dst.All() {
		if err != nil {
			return derived, end, err
		}

		optStat, err := dst.Stat(name)
		if err != nil {
			return derived, end, err
		}

		if optStat.IsNone() || !optStat.Unwrap().Derived {
			continue
		}

		stat := optStat.Unwrap()
		store, err := dst.Open(name, blob.OpenStoreOptions{Type: stat.Type, Derived: true})
		if err != nil {
			return derived, end, fmt.Errorf("cannot open derived store %s: %w", name, err)
		}

		endStore, err := blob.Invalidate(store)
		if err != nil {
			return derived, end, fmt.Errorf("cannot invalidate derived store %s: %w", name, err)
		}

		ends = append(ends, endStore)
		derived[name] = true
		slog.Info("restore", "action", "derived store invalidated", "name", name)
	}

	return derived, end, nil
}

func decodeIndex(zipReader *zip.Reader) (Index, error) {
	var index Index
	reader, err := zipReader.Open("index.json")
	if err != nil {
		return index, fmt.Errorf("could not open index.json: %w", err)
	}

	defer reader.Close()

	if err := json.NewDecoder(reader).Decode(&index); err != nil {
		return index, fmt.Errorf("could not decode index.json: %w", err)
	}

	return index, nil
}
