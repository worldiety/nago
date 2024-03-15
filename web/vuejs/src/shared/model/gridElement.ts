import type { GridCellElement } from '@/shared/model/gridCellElement';

export interface GridElement {
	type: 'Grid';
	columns: number;
	rows: number;
	gap: number;
	padding: string;
	cells: GridCellElement[];
}
