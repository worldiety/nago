/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { cssLengthValue } from '@/components/shared/length';
import { Padding } from '@/shared/proto/nprotoc_gen';

// paddingCSS applies the padding length values. Note, that negative paddings are interpreted as negative margins,
// because negative padding values are not allowed and it seems practical to move views around for some nice effects.
export function paddingCSS(padding?: Padding): string[] {
	const styles: string[] = [];

	if (!padding) {
		return styles;
	}

	if (padding.bottom) {
		if (padding.bottom.startsWith('-')) {
			styles.push(`margin-bottom:${cssLengthValue(padding.bottom)}`);
		} else {
			styles.push(`padding-bottom:${cssLengthValue(padding.bottom)}`);
		}
	}

	if (padding.top) {
		if (padding.top.startsWith('-')) {
			styles.push(`margin-top: ${cssLengthValue(padding.top)}`);
		} else {
			styles.push(`padding-top:${cssLengthValue(padding.top)}`);
		}
	}

	if (padding.right) {
		if (padding.right.startsWith('-')) {
			styles.push(`margin-right:${cssLengthValue(padding.right)}`);
		} else {
			styles.push(`padding-right:${cssLengthValue(padding.right)}`);
		}
	}

	if (padding.left) {
		if (padding.left.startsWith('-')) {
			styles.push(`margin-left:${cssLengthValue(padding.left)}`);
		} else {
			styles.push(`padding-left:${cssLengthValue(padding.left)}`);
		}
	}

	return styles;
}

// marginCSS is like padding but interprets all padding lengths as margin length
export function marginCSS(padding?: Padding): string[] {
	const styles: string[] = [];

	if (!padding) {
		return styles;
	}

	if (padding.bottom) {
		styles.push(`margin-bottom:${cssLengthValue(padding.bottom)}`);
	}

	if (padding.top) {
		styles.push(`margin-top: ${cssLengthValue(padding.top)}`);
	}

	if (padding.right) {
		styles.push(`margin-right:${cssLengthValue(padding.right)}`);
	}

	if (padding.left) {
		styles.push(`margin-left:${cssLengthValue(padding.left)}`);
	}

	return styles;
}
