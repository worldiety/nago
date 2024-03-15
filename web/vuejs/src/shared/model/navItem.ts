import type { NavAction } from '@/shared/model/navAction';
import type { Image } from '@/shared/model/image';

export interface NavItem {
	title: string;
	link: NavAction;
	anchor: string,
	icon: Image;
}
