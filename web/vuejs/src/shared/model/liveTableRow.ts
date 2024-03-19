import type { ComponentList } from '@/shared/model/componentList';
import type { LiveTableCell } from '@/shared/model/liveTableCell';

export interface LiveTableRow {
	type: 'TableRow';
	id: number;
	cells: ComponentList<LiveTableCell>;
}
