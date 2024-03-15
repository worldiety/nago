import type { LiveComponent } from '@/shared/model/liveComponent';
import type { ComponentList } from '@/shared/model/componentList';

export interface LivePage {
	type: 'Page'
	root: LiveComponent
	modals: ComponentList<LiveComponent>
	token: string
}
