import type { UiElement } from '@/shared/model/uiElement';
import type { Redirection } from '@/shared/model/redirection';

export interface UiDescription {
	renderTree: UiElement;
	viewModel: any;
	redirect: Redirection | null;
}
