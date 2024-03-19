import type { Action } from '@/shared/model/action';

export interface Button {
	type: 'Button';
	caption: string;
	action: Action;
}
