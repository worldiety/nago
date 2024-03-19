import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveChip {
	type: 'Chip';
	caption: PropertyString;
	action: PropertyFunc;
	onClose: PropertyFunc;
	color: PropertyString;
}
