import type { ComponentList } from '@/shared/model/componentList';
import type { LiveComponent } from '@/shared/model/liveComponent';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveCard {
	type: 'Card'
	children: ComponentList<LiveComponent>
	action: PropertyFunc
}
