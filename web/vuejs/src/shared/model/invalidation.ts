import type { ComponentList } from '@/shared/model/componentList';
import type { LiveComponent } from '@/shared/model/liveComponent';

export interface Invalidation {
	type: 'Invalidation' | 'HistoryPushState' | 'HistoryBack';
	root: LiveComponent;
	modals: ComponentList<LiveComponent>;
	token: string;
}
