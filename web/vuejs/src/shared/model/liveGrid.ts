import type { ComponentList } from '@/shared/model/componentList';
import type { LiveGridCell } from '@/shared/model/liveGridCell';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveGrid {
	type: 'Grid'
	id: number
	cells: ComponentList<LiveGridCell>
	rows: PropertyInt
	columns: PropertyInt
	smColumns: PropertyInt
	mdColumns: PropertyInt
	lgColumns: PropertyInt
	gap: PropertyString
}
