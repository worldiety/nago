import type { TextElement } from '@/shared/model/textElement';
import type { UiEvent } from '@/shared/model/uiEvent';

export interface ButtonElement {
	type: 'Button';
	title: TextElement;
	onClick: UiEvent;
}
