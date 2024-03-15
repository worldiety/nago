import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveChip {
	type: 'Chip'
	caption: PropertyString
	action: PropertyFunc
	onClose: PropertyFunc
	color: PropertyString
}
