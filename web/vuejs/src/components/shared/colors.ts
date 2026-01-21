/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { Color } from '@/shared/proto/nprotoc_gen';

export enum NamedColor {
	// Primary call-to-action intention.
	Primary = 'p',

	// Secondary call-to-action intention.
	Secondary = 's',

	// Tertiary call-to-action intention.
	Tertiary = 't',

	// Error describes a negative or a destructive intention. In Western Europe usually red. Use it, when the
	// user cannot continue normally and has to fix the problem first.
	Error = 'n',

	// Warning describes a critical condition. In Western Europe usually yellow. Use it to warn on situations,
	// which may result in a future error condition.
	Warning = 'c',

	// Positive describes a good condition or a confirming intention. In Western Europe usually green.
	// Use it to symbolize something which has been successfully applied.
	Positive = 'o',

	// Informative shall be used to highlight something, which just changed. E.g. a newly added component or
	// a recommendation from a system. Do not use it to highlight text. In Western Europe usually blue.
	Informative = 'i',

	// Regular shall be used for any default of any UI element which has no special semantic intention.
	// An empty color is always regular.
	Regular = 'r',
}

export function colorValue(color?: Color): string {
	if (!color) {
		return '';
	}

	if (color.startsWith('#')) {
		return color;
	}

	let opacity = 100;
	if (color.includes('/')) {
		opacity = opacityValue(color);
		color = color.split('/')[0];
	}

	return `color-mix(in srgb, var(--${color}) ${opacity}%, rgba(255, 255, 255, 0))`;
}

function opacityValue(color?: Color): number {
	if (color?.startsWith('#')) {
		let split = '';
		if (color?.length === 5) split = color.substring(4);
		if (color?.length === 9) split = color.substring(8);
		if (split) {
			const num = Number(`0x${split}`);
			return num / 255 * 100;
		}

		return 100;
	}

	if (color?.includes('/')) {
		const split = color?.split('/').pop();
		return split ? parseInt(split) / 255 * 100 : 0;
	}

	return 0;
}
