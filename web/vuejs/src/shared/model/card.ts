import type { Action } from '@/shared/model/action';
import type { Button } from '@/shared/model/button';
import type { FontIcon } from '@/shared/model/fontIcon';

export interface Card {
	type: 'Card';
	title: string;
	subtitle: string;
	content: any;
	prependIcon: FontIcon;
	appendIcon: FontIcon;
	actions: Button[];
	primaryAction: Action;
}

export interface CardView {
	type: 'CardView';
	cards: Card[];
}
