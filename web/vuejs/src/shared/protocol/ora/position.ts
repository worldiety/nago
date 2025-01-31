import type { Length } from '@/shared/protocol/ora/length';

export type PositionType = number;

// PositionDefault is the default and any explicit position value have no effect.
// See also https://developer.mozilla.org/de/docs/Web/CSS/position#static.
export const PositionDefault = 0;

// PositionOffset is like PositionDefault but moves the element by applying the given position values after
// layouting. See also https://developer.mozilla.org/de/docs/Web/CSS/position#relative.
export const PositionOffset = 1;

// PositionAbsolute removes the element from the layout and places it using the given values in an absolute way
// within any of its parent layouted as PositionOffset. If no parent with PositionOffset is found, the viewport
// is used. See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
export const PositionAbsolute = 2;

// PositionFixed removes the element from the layout and places it at a fixed position according to the viewport
// independent of the scroll position. See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
export const PositionFixed = 3;

// PositionSticky is here for completion, and it is unclear which rules to follow on mobile clients.
// See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
export const PositionSticky = 4;

export interface Position {
	/**
	 * Kind   PositionType `json:"k"`
	 */
	k?: PositionType;

	/**
	 * Left   Length       `json:"l"`
	 */
	l?: Length;
	/**
	 * Top    Length       `json:"t"`
	 */
	t?: Length;
	/**
	 * Right  Length       `json:"r"`
	 */
	r?: Length;

	/**
	 * Bottom Length       `json:"b"`
	 */
	b?: Length;
}
