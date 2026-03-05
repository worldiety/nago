/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * bool2Str converts the given bool into a Go backend-string-parseable event value representation.
 */
export function bool2Str(b: boolean): string {
	return b ? 'true' : 'false';
}

export function randomStr(len: number): string {
	const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
	let result = '';
	for (let i = 0; i < len; i++) {
		result += chars.charAt(Math.floor(Math.random() * chars.length));
	}
	return result;
}
