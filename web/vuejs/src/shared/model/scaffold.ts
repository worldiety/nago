import type { Breadcrumb } from '@/shared/model/breadcrumb';
import type { NavItem } from '@/shared/model/navItem';

export interface Scaffold {
	type: 'Scaffold';
	children: URL[];
	title: string;
	navigation: NavItem[];
	breadcrumbs: Breadcrumb[];
}
