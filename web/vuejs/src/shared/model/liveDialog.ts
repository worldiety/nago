import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyComponent } from '@/shared/model/propertyComponent';
import type { LiveComponent } from '@/shared/model/liveComponent';
import type { ComponentList } from '@/shared/model/componentList';
import type { LiveButton } from '@/shared/model/liveButton';

export interface LiveDialog {
	type: 'Dialog'
	id: number
	title: PropertyString
	body: PropertyComponent<LiveComponent>
	icon: PropertyString
	actions: ComponentList<LiveButton>
}
