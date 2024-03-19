import type { ComponentList } from '@/shared/model/componentList';
import type { LiveButton } from '@/shared/model/liveButton';
import type { LiveComponent } from '@/shared/model/liveComponent';
import type { PropertyComponent } from '@/shared/model/propertyComponent';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveDialog {
	type: 'Dialog';
	id: number;
	title: PropertyString;
	body: PropertyComponent<LiveComponent>;
	icon: PropertyString;
	actions: ComponentList<LiveButton>;
}
