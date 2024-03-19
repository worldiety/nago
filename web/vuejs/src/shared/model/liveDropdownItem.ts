import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveDropdownItem {
	id: number,
	type: 'DropdownItem',
	content: PropertyString,
	onClicked: PropertyFunc,
}
