import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveText {
	type: 'Text';
	id: number;
	value: PropertyString;
	color: PropertyString;
	colorDark: PropertyString;
	size: PropertyString;
	onClick: PropertyFunc;
	onHoverStart: PropertyFunc;
	onHoverEnd: PropertyFunc;
}
