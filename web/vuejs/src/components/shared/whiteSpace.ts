/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { WhiteSpaceValues } from '@/shared/proto/nprotoc_gen';

export function whiteSpaceCSS(whiteSpace?: WhiteSpaceValues): string[] {
	const css: string[] = [];

	switch (whiteSpace) {
		case WhiteSpaceValues.WhiteSpaceNoWrap:
			css.push(`white-space: nowrap`);
			break;
		default:
			css.push(`white-space: normal`);
	}

	return css;
}
