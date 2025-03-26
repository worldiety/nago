// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package backup

import (
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob/crypto"
)

func NewImportMasterKey(setCryptoKey func(crypto.EncryptionKey)) ReplaceMasterKey {
	return func(subject auth.Subject, key string) error {
		if err := subject.Audit(PermReplaceMasterKey); err != nil {
			return err
		}

		buf, err := hex.DecodeString(key)
		if err != nil {
			return err
		}

		k := crypto.NewEncryptionKey()
		if len(buf) != len(k) {
			return fmt.Errorf("invalid key length, expected %d, got %d", len(k), len(buf))
		}

		copy(k[:], buf)
		setCryptoKey(k)

		return nil
	}
}
