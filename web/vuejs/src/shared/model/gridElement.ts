export interface GridElement {
	type: 'Grid';
	columns: number;
	rows: number;
	gap: number;
	padding: string;
	cells: GridCellElement[];
}
