import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyBool } from '@/shared/model/propertyBool';

export interface LiveButton {
	type: 'Button'
	id: number
	caption: PropertyString
	preIcon: PropertyString
	postIcon: PropertyString
	color: PropertyString
	action: PropertyFunc
	disabled: PropertyBool
}
