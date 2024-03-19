import type { Image } from '@/shared/model/image';
import type { NavAction } from '@/shared/model/navAction';

export interface NavItem {
	title: string;
	link: NavAction;
	anchor: string;
	icon: Image;
}
