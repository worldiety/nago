import type { Redirection } from '@/shared/model/redirection';
import type { UiElement } from '@/shared/model/uiElement';

export interface UiDescription {
	renderTree: UiElement;
	viewModel: any;
	redirect: Redirection | null;
}
