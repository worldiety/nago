import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveImage {
	type: 'Image';
	url: PropertyString;
	downloadToken: PropertyString;
	caption: PropertyString;
}
