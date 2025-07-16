// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"context"
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"io"
	"os"
	"sync"
	"time"
)

func signUnqualified(mutex *sync.Mutex, openImgReader image.OpenReader, repo Repository, stores blob.Stores, idx *inMemoryIndex, uid user.ID, cdata AnonSignData) (ID, error) {
	if cdata.Firstname == "" {
		return "", fmt.Errorf("firstname must not be empty: %w", user.InvalidSubjectErr)
	}

	if cdata.Lastname == "" {
		return "", fmt.Errorf("lastname must not be empty: %w", user.InvalidSubjectErr)
	}

	if cdata.Email == "" {
		return "", fmt.Errorf("email must not be empty: %w", user.InvalidSubjectErr)
	}

	mutex.Lock()
	defer mutex.Unlock()

	ctx := context.Background()

	var docs []Document
	for _, resource := range cdata.Resources {
		optStore, err := stores.Get(resource.Name)
		if err != nil {
			return "", fmt.Errorf("cannot get the store '%s': %w", resource.Name, err)
		}

		if optStore.IsNone() {
			return "", fmt.Errorf("the store was not found '%s': %w", resource.Name, os.ErrNotExist)
		}

		store := optStore.Unwrap()
		size, hash, err := hashBlob(ctx, store, resource.ID)
		if err != nil {
			return "", fmt.Errorf("cannot hash blob '%s.%s': %w", resource.Name, resource.ID, err)
		}

		docs = append(docs, Document{
			Resource: resource,
			Size:     size,
			Hash:     hash,
		})
	}

	var imgHash Sha3H256
	if cdata.SignatureImage != "" {
		optReader, err := openImgReader(user.SU(), cdata.SignatureImage)
		if err != nil {
			return "", fmt.Errorf("cannot open signature image reader '%s': %w", cdata.SignatureImage, err)
		}

		_, hash, err := hashOptReader(optReader)
		if err != nil {
			return "", fmt.Errorf("cannot hash signature image reader '%s': %w", cdata.SignatureImage, err)
		}

		imgHash = hash
	}

	var prevSigHash Sha3H256
	myNumber := 1
	if optSig := idx.LastSignature(); optSig.IsSome() {
		prevSigHash = optSig.Unwrap().Hash
		myNumber = optSig.Unwrap().Number + 1
	}

	sig := Signature{
		ID:                    data.RandIdent[ID](),
		PreviousSignatureHash: prevSigHash,
		Number:                myNumber,
		Timestamp:             xtime.UnixMilliseconds(time.Now().UnixMilli()),
		Timezone:              time.Local.String(), // TODO should subject provide its timezone?
		Firstname:             cdata.Firstname,
		Lastname:              cdata.Lastname,
		User:                  uid,
		Email:                 cdata.Email,
		Image:                 cdata.SignatureImage,
		ImageHash:             imgHash,
		Documents:             xslices.Wrap(docs...),
	}

	sig.Hash = sig.CalcHash()

	// proof that there is no collision
	optSig, err := repo.FindByID(sig.ID)
	if err != nil {
		return "", err
	}

	if optSig.IsSome() {
		return sig.ID, fmt.Errorf("signature already exists: %w", os.ErrExist)
	}

	if err := repo.Save(sig); err != nil {
		return "", fmt.Errorf("cannot save signature '%s': %w", sig.ID, err)
	}

	idx.Index(sig)

	return sig.ID, nil
}

func hashBlob(ctx context.Context, store blob.Store, key string) (int64, Sha3H256, error) {
	optReader, err := store.NewReader(ctx, key)
	if err != nil {
		return 0, "", fmt.Errorf("cannot create reader for '%s': %w", key, err)
	}

	return hashOptReader(optReader)
}

func hashOptReader(optReader option.Opt[io.ReadCloser]) (int64, Sha3H256, error) {
	if optReader.IsNone() {
		return 0, "", fmt.Errorf("missing blob: %w", os.ErrNotExist)
	}

	reader := optReader.Unwrap()
	defer reader.Close()

	hasher := sha3.New256()

	n, err := io.Copy(hasher, reader)
	if err != nil {
		return 0, "", fmt.Errorf("cannot hash blob: %w", err)
	}

	sum := hasher.Sum(nil)
	return n, Sha3H256(hex.EncodeToString(sum)), nil
}
