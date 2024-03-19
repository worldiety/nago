import type { ComponentList } from '@/shared/model/componentList';
import type { LiveComponent } from '@/shared/model/liveComponent';

export interface VBox {
	type: 'VBox';
	children: ComponentList<LiveComponent>;
}
