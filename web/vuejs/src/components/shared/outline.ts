/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { colorValue } from '@/components/shared/colors';
import { Outline } from '@/shared/proto/nprotoc_gen';

export function outlineCSS(outline?: Outline): string[] {
	if (!outline) return [];

	const css: string[] = [];

	if (outline.color) css.push(`outline-color: ${colorValue(outline.color)}`);
	if (outline.width) css.push(`outline-width: ${outline.width}px`);
	if (outline.offset) css.push(`outline-offset: ${outline.offset}px`);

	return css;
}
