import type { ComponentList } from '@/shared/model/componentList';
import type { LiveComponent } from '@/shared/model/liveComponent';

export interface LivePage {
	type: 'Page';
	root: LiveComponent;
	modals: ComponentList<LiveComponent>;
	token: string;
}
