/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { cssLengthValue } from '@/components/shared/length';
import { Background, Frame } from '@/shared/proto/nprotoc_gen';

export function backgroundCSS(bg?: Background): string[] {
	const styles: string[] = [];
	if (!bg) {
		return styles;
	}

	if (bg.image) {
		styles.push('background-image:' + bg.image.value.join(', '));
	}

	if (bg.positionX) {
		styles.push('background-position-x:' + bg.positionX + '%;');
	}

	if (bg.positionY) {
		styles.push('background-position-y:' + bg.positionY + '%;');
	}

	if (bg.repeat) {
		styles.push('background-repeat:' + bg.repeat);
	}

	if (bg.size) {
		styles.push('background-size:' + bg.size);
	}

	return styles;
}
