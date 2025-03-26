/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { Length } from '@/shared/proto/nprotoc_gen';

export function cssLengthValue(l?: Length): string {
	if (!l || l === '') {
		return '';
	}

	// px is just wrong in CSS, they always mean dp
	l = l.replaceAll('dp', 'px');

	if (l.charAt(0) === '-' || (l.charAt(0) >= '0' && l.charAt(0) <= '9')) {
		return l;
	}

	if (l.startsWith('calc')) {
		return l;
	}

	return `var(--${l})`;
}

export function cssLengthValue0Px(l?: Length): string {
	if (!l) {
		return '0px';
	}

	l = l.replaceAll('dp', 'px');
	return l;
}
