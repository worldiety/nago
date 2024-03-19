import type { LiveComponent } from '@/shared/model/liveComponent';

export interface PropertyComponent<T extends LiveComponent> {
	id: number;
	name: string;
	value: T;
}
