// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package backup

import (
	"encoding/hex"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob/crypto"
)

func NewExportMasterKey(getCryptoKey func() crypto.EncryptionKey) ExportMasterKey {
	return func(subject auth.Subject) (string, error) {
		if err := subject.Audit(PermExportMasterKey); err != nil {
			return "", err
		}

		key := getCryptoKey()
		return hex.EncodeToString(key[:]), nil
	}
}
