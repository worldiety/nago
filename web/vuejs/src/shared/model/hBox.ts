import type { ComponentList } from '@/shared/model/componentList';
import type { LiveComponent } from '@/shared/model/liveComponent';
import type { PropertyString } from '@/shared/model/propertyString';

export interface HBox {
	type: 'HBox'
	children: ComponentList<LiveComponent>
	alignment: PropertyString
}
