import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveDropdownItem {
	id: number,
	type: 'DropdownItem',
	content: PropertyString,
	onClicked: PropertyFunc,
}
