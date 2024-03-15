import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveDropdownItem {
	type: 'DropdownItem',
	itemIndex: PropertyInt,
	content: PropertyString,
	onSelected: PropertyFunc,
}
