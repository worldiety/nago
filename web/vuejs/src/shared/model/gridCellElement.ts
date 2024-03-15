export interface GridCellElement {
	type: 'GridCell';
	colSpan: number;
	rowSpan: number;
	views: UiElement[];
}
