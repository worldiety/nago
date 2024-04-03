import type { ButtonElement } from '@/shared/model/buttonElement';
import type { CardView } from '@/shared/model/card';
import type { FormField } from '@/shared/model/formField';
import type { GridElement } from '@/shared/model/gridElement';
import type { ListView } from '@/shared/model/listView';
import type { LiveComponent } from '@/shared/model/liveComponent';
import type { Scaffold } from '@/shared/model/scaffold';
import type { TextElement } from '@/shared/model/textElement';
import type {WebView} from "@/shared/model/webView";

export type UiElement =
	| TextElement
	| ButtonElement
	| GridElement
	| Scaffold
	| ListView
	| FormField
	| CardView
	| LiveComponent
	| WebView
	| SVGElement;
