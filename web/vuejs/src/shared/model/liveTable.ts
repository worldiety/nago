import type { ComponentList } from '@/shared/model/componentList';
import type { LiveTableCell } from '@/shared/model/liveTableCell';
import type { LiveTableRow } from '@/shared/model/liveTableRow';

export interface LiveTable {
	type: 'Table';
	id: number;
	headers: ComponentList<LiveTableCell>;
	rows: ComponentList<LiveTableRow>;
}
