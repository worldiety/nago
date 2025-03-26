/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

export function gapSize2Tailwind(s: string): string {
	if (s == null || s == '') {
		return '';
	}

	if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
		return 'gap-[' + s + ']';
	}

	return s;
}
