import type { NavItem } from '@/shared/model/navItem';
import type { Breadcrumb } from '@/shared/model/breadcrumb';

export interface Scaffold {
	type: 'Scaffold';
	children: URL[];
	title: string;
	navigation: NavItem[];
	breadcrumbs: Breadcrumb[]
}
