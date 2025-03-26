/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { cssLengthValue } from '@/components/shared/length';
import { Position, PositionTypeValues } from '@/shared/proto/nprotoc_gen';

export function positionCSS(position?: Position): string[] {
	const styles: string[] = [];

	if (!position) {
		return styles;
	}

	switch (position.kind) {
		case PositionTypeValues.PositionDefault:
			//styles.push('position:static'); // TODO not sure if we should switch that to change inherit behavior
			break;
		case PositionTypeValues.PositionAbsolute:
			styles.push('position:absolute');
			break;
		case PositionTypeValues.PositionOffset:
			styles.push('position:relative');
			break;
		case PositionTypeValues.PositionSticky:
			styles.push('position:sticky');
			break;
		case PositionTypeValues.PositionFixed:
			styles.push('position:fixed');
			break;
	}

	if (position.left) {
		styles.push('left:' + cssLengthValue(position.left));
	}

	if (position.top) {
		styles.push('top:' + cssLengthValue(position.top));
	}

	if (position.right) {
		styles.push('right:' + cssLengthValue(position.right));
	}

	if (position.bottom) {
		styles.push('bottom:' + cssLengthValue(position.bottom));
	}

	return styles;
}
