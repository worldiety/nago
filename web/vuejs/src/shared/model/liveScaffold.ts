import type { PropertyString } from '@/shared/model/propertyString';
import type { ComponentList } from '@/shared/model/componentList';
import type { LiveButton } from '@/shared/model/liveButton';
import type { PropertyComponent } from '@/shared/model/propertyComponent';
import type { LiveComponent } from '@/shared/model/liveComponent';

export interface LiveScaffold {
	type: 'Scaffold'
	id: number
	title: PropertyString
	breadcrumbs: ComponentList<LiveButton> // currently ever of LiveButton
	menu: ComponentList<LiveButton> // currently always of LiveButton
	body: PropertyComponent<LiveComponent>
	topbarLeft: PropertyComponent<LiveComponent>
	topbarMid: PropertyComponent<LiveComponent>
	topbarRight: PropertyComponent<LiveComponent>
}
