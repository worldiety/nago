export interface Scaffold {
	type: 'Scaffold';
	children: URL[];
	title: string;
	navigation: NavItem[];
	breadcrumbs: Breadcrumb[]
}
