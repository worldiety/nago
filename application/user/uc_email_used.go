// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

func NewEMailUsed(mailIdx *MailIndex) EMailUsed {
	return func(email Email) (bool, error) {
		return mailIdx.Used(email)
	}
}
