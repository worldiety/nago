/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import type { Color } from '@/shared/proto/nprotoc_gen';

export function gapSize2Tailwind(s: string): string {
	if (s == null || s == '') {
		return '';
	}

	if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
		return 'gap-[' + s + ']';
	}

	return s;
}

export function colorToHexValue(color: Color): string {
	if (color.startsWith('#')) {
		return color;
	}

	if (color.startsWith('var(')) {
		color = color.replace('var(', '').replace(')', '');
	}

	if (!color.startsWith('--')) {
		color = '--'.concat(color);
	}

	return getComputedStyle(document.documentElement).getPropertyValue(color).trim();
}
