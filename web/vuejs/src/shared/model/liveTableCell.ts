import type { LiveComponent } from '@/shared/model/liveComponent';
import type { PropertyComponent } from '@/shared/model/propertyComponent';

export interface LiveTableCell {
	type: 'TableCell';
	id: number;
	body: PropertyComponent<LiveComponent>;
}
