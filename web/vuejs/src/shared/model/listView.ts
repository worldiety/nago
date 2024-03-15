import type { ListViewLinks } from '@/shared/model/listViewLinks';

export interface ListView {
	type: 'ListView';
	links: ListViewLinks;
}
