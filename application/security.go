// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/pkg/blob/crypto"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	envNagoMasterKey  = "NAGO_MASTER_KEY"
	fileNagoMasterKey = ".masterkey"
)

// MasterKey returns a 32 byte key. Use this master key to encrypt or decrypt all other secrets or authenticate
// against other vaults. To resolve the master key, the current implementation does the following steps:
//   - lookup a hex encoded 32 byte sequence from the environment variable NAGO_MASTER_KEY
//   - if lookup fails, a random key is generated and written/read into the applications private home dir into .masterkey
func (c *Configurator) MasterKey() (crypto.EncryptionKey, error) {
	hexKey := os.Getenv(envNagoMasterKey)
	if hexKey := strings.TrimSpace(hexKey); hexKey != "" {
		buf, err := hex.DecodeString(hexKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode NAGO_MASTER_KEY")
		}

		var key [32]byte
		copy(key[:], buf)

		return &key, nil
	}

	slog.Warn("missing environment variable NAGO_MASTER_KEY, using fallback file")

	// fallback to file write
	secFile := c.masterKeyFile()
	if _, err := os.Stat(secFile); os.IsNotExist(err) {
		slog.Info("generating local NAGO_MASTER_KEY")
		// file does not exist, thus init with random key
		key := crypto.NewEncryptionKey()

		if err := c.WriteMasterKey(key); err != nil {
			return nil, err
		}

		return key, nil
	}

	// fallback to file read, already exists
	hexBuf, err := os.ReadFile(secFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read local NAGO_MASTER_KEY from %s: %w", secFile, err)
	}

	var key [32]byte
	if _, err := hex.Decode(key[:], hexBuf); err != nil {
		return nil, fmt.Errorf("failed to decode NAGO_MASTER_KEY from file %s: %w", secFile, err)
	}

	return &key, nil
}

func (c *Configurator) masterKeyFile() string {
	return filepath.Join(c.DataDir(), fileNagoMasterKey)
}

func (c *Configurator) WriteMasterKey(key crypto.EncryptionKey) error {
	// disallow all other users to read the key
	if err := os.WriteFile(c.masterKeyFile(), []byte(hex.EncodeToString((*key)[:])), 0600); err != nil {
		return fmt.Errorf("failed to write local NAGO_MASTER_KEY into %s: %w", c.masterKeyFile(), err)
	}

	return nil
}
