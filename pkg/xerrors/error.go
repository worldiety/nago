// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xerrors

func WithFields(msg string, args ...string) error {
	if len(args)%2 != 0 {
		args = append(args, "!MISSING_VALUE")
	}

	f := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		f[args[i]] = args[i+1]
	}

	return errorWithFields{msg, f}
}

type errorWithFields struct {
	msg    string
	fields map[string]string
}

func (e errorWithFields) UnwrapFields() any {
	return e.fields
}

func (e errorWithFields) Error() string {
	return e.msg
}
