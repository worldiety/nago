// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package inspector

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/magic"
	"io"
	"os"
)

func NewFilter() Filter {
	return func(subject auth.Subject, store blob.Store, opts FilterOptions) (PageResult, error) {
		if err := subject.Audit(PermDataInspector); err != nil {
			return PageResult{}, err
		}

		if opts.PageSize == 0 {
			opts.PageSize = 20
		}

		if opts.MaxDataSize == 0 {
			opts.MaxDataSize = 1024 * 64
		}

		if opts.Context == nil {
			opts.Context = context.Background()
		}

		var res PageResult
		idx := 0
		offset := opts.PageSize * opts.PageNo
		maxIdx := offset + opts.PageSize

		for key, err := range store.List(opts.Context, blob.ListOptions{}) {
			isPageResult := idx >= offset && idx < maxIdx

			if isPageResult {
				if err != nil {
					res.Entries = append(res.Entries, PageEntry{
						Key:   key,
						Error: err,
					})
					continue
				}

				entry := PageEntry{
					Key: key,
				}

				if !opts.OnlyKeys {
					buf, err := readAtMost(opts.Context, store, key, opts.MaxDataSize)
					if entry.Error == nil && err != nil {
						entry.Error = err
					}

					entry.Data = buf
				}

				if opts.DetectMimeTypes {
					mt, err := detectMimeType(opts.Context, store, key)
					if entry.Error == nil && err != nil {
						entry.Error = err
					}

					entry.MimeType = mt
				}

				res.Entries = append(res.Entries, entry)
			}

			idx++
		}

		res.Count = idx
		res.PageNo = opts.PageNo
		res.PageSize = opts.PageSize
		res.Pages = res.Count / res.PageSize
		if res.Count%res.PageSize != 0 {
			res.Pages++
		}

		return res, nil
	}
}

func readAtMost(ctx context.Context, store blob.Store, key string, atMost int) ([]byte, error) {
	optReader, err := store.NewReader(ctx, key)
	if err != nil {
		return nil, err
	}

	if optReader.IsNone() {
		return nil, os.ErrNotExist
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	tmp := make([]byte, atMost)
	n, err := reader.Read(tmp)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return tmp[:n], nil
}

func detectMimeType(ctx context.Context, store blob.Store, key string) (string, error) {
	optReader, err := store.NewReader(ctx, key)
	if err != nil {
		return "", err
	}

	if optReader.IsNone() {
		return "", os.ErrNotExist
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	tmp := make([]byte, 1024)
	_, err = reader.Read(tmp)
	if err != nil && err != io.EOF {
		return "", err
	}

	return magic.Detect(tmp), nil
}
