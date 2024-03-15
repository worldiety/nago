import type { UiElement } from '@/shared/model/uiElement';

export interface GridCellElement {
	type: 'GridCell';
	colSpan: number;
	rowSpan: number;
	views: UiElement[];
}
