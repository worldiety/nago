import type { ListViewLinks } from '@/shared/model/components/listViewLinks';

export interface ListView {
	type: 'ListView';
	links: ListViewLinks;
}
