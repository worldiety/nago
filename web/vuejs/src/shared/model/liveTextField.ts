import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveTextField {
	type: 'TextField'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	value: PropertyString
	disabled: PropertyBool
	onTextChanged: PropertyFunc
}
