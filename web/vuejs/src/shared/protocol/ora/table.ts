/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { Border } from '@/shared/protocol/ora/border';
import type { Color } from '@/shared/protocol/ora/color';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Frame } from '@/shared/protocol/ora/frame';
import type { Padding } from '@/shared/protocol/ora/padding';
import type { TableHeader } from '@/shared/protocol/ora/tableHeader';
import type { TableRow } from '@/shared/protocol/ora/tableRow';

/**
 * Table represents a pre-styled table with limited styling capabilities. Use Grid for maximum flexibility.
 */
export interface Table {
	// Type
	type: 'B' /*ComponentType*/;
	// Header
	h /*omitempty*/? /*Header*/ : TableHeader;
	// Rows
	r /*omitempty*/? /*Rows*/ : TableRow[];
	// Frame
	f /*omitempty*/? /*Frame*/ : Frame;
	// BackgroundColor
	bgc /*omitempty*/? /*BackgroundColor*/ : Color;
	// Border
	b /*omitempty*/? /*Border*/ : Border;
	// DefaultCellPadding
	p /*omitempty*/? /*DefaultCellPadding*/ : Padding;
	// RowDividerColor
	rdc /*omitempty*/? /*RowDividerColor*/ : Color;
	// HeaderDividerColor
	hdc /*omitempty*/? /*HeaderDividerColor*/ : Color;
}
