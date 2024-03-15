import type { TextElement } from '@/shared/model/textElement';
import type { ButtonElement } from '@/shared/model/buttonElement';
import type { GridElement } from '@/shared/model/gridElement';
import type { Scaffold } from '@/shared/model/scaffold';
import type { ListView } from '@/shared/model/listView';
import type { FormField } from '@/shared/model/formField';
import type { CardView } from '@/shared/model/card';
import type { LiveComponent } from '@/shared/model/liveComponent';

export type UiElement =
	TextElement
	| ButtonElement
	| GridElement
	| Scaffold
	| ListView
	| FormField
	| CardView
	| LiveComponent
	| SVGElement;
