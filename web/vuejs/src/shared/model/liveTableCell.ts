import type { PropertyComponent } from '@/shared/model/propertyComponent';
import type { LiveComponent } from '@/shared/model/liveComponent';

export interface LiveTableCell {
	type: 'TableCell'
	id: number
	body: PropertyComponent<LiveComponent>
}
