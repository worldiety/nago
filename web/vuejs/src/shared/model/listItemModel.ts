import type { NavAction } from '@/shared/model/navAction';

export interface ListItemModel {
	type: 'ListItem';
	id: string;
	title: string;
	action: NavAction;
}
