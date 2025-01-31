import { Color } from '@/shared/protocol/ora/color';

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

	return `var(--${color})`;
}
