import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveText {
	type: 'Text'
	id: number
	value: PropertyString
	color: PropertyString
	colorDark: PropertyString
	size: PropertyString
	onClick: PropertyFunc
	onHoverStart: PropertyFunc
	onHoverEnd: PropertyFunc
}
