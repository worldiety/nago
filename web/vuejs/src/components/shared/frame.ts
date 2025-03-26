/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { cssLengthValue } from '@/components/shared/length';
import { Frame } from '@/shared/proto/nprotoc_gen';

export function frameCSS(frame?: Frame): string[] {
	const styles: string[] = [];
	if (!frame) {
		return styles;
	}

	if (frame.width) {
		styles.push('width:' + cssLengthValue(frame.width));
	}

	if (frame.minWidth) {
		styles.push('min-width:' + cssLengthValue(frame.minWidth));
	}

	if (frame.maxWidth) {
		styles.push('max-width:' + cssLengthValue(frame.maxWidth));
	}

	if (frame.height) {
		styles.push('height:' + cssLengthValue(frame.height));
	}

	if (frame.minHeight) {
		styles.push('min-height:' + cssLengthValue(frame.minHeight));
	}

	if (frame.maxHeight) {
		styles.push('max-height:' + cssLengthValue(frame.maxHeight));
	}

	return styles;
}

export function frameCSSObject(frame?: Frame): Object | undefined {
	if (!frame) {
		return undefined;
	}

	return {
		width: cssLengthValue(frame.width),
		height: cssLengthValue(frame.height),
	};
}
