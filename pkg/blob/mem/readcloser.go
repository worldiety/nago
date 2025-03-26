// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mem

import "bytes"

type readerCloser struct {
	*bytes.Reader
}

func (readerCloser) Close() error {
	return nil
}
