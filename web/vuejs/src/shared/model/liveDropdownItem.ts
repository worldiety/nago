import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveDropdownItem {
	type: 'DropdownItem';
	itemIndex: PropertyInt;
	content: PropertyString;
	onSelected: PropertyFunc;
}
