import { randomStr } from '@/components/shared/util';

type PseudoClass = 'hover' | 'focus' | 'active';

export class CssClasses {
	private static classes: Map<string, string> = new Map();
	private static defaultStyleElem: HTMLStyleElement;
	private static hoverStyleElem: HTMLStyleElement;
	private static focusStyleElem: HTMLStyleElement;
	private static activeStyleElem: HTMLStyleElement;

	private static init() {
		if (this.defaultStyleElem && this.hoverStyleElem && this.focusStyleElem && this.activeStyleElem) return;

		this.defaultStyleElem = document.createElement('style');
		this.hoverStyleElem = document.createElement('style');
		this.focusStyleElem = document.createElement('style');
		this.activeStyleElem = document.createElement('style');

		document.body.appendChild(this.defaultStyleElem);
		document.body.appendChild(this.hoverStyleElem);
		document.body.appendChild(this.focusStyleElem);
		document.body.appendChild(this.activeStyleElem);
	}

	public static getOrCreate(styles: string[], pseudoClass?: PseudoClass): string {
		const mapKey = this.createMapKey(styles, pseudoClass);
		const existing = this.classes.get(mapKey);
		if (existing) return existing;
		return this.createCssClass(mapKey, styles, pseudoClass);
	}

	private static createCssClass(mapKey: string, styles: string[], pseudoClass?: PseudoClass): string {
		this.init();
		const stylesStr = styles.join(';\n');
		const className = randomStr(12);
		let selectorStr = `.${className}${pseudoClass ? `:${pseudoClass}` : ''}`;
		if (pseudoClass === 'focus') selectorStr += `,\n.${className}:focus-visible`;
		const classStr = `${selectorStr} {\n${stylesStr}\n}`;
		this.addClassToElem(classStr, pseudoClass);
		this.classes.set(mapKey, className);

		return className;
	}

	private static addClassToElem(classStr: string, pseudoClass?: PseudoClass) {
		switch (pseudoClass) {
			case 'hover':
				this.hoverStyleElem.innerHTML += `\n\n${classStr}`;
				break;
			case 'focus':
				this.focusStyleElem.innerHTML += `\n\n${classStr}`;
				break;
			case 'active':
				this.activeStyleElem.innerHTML += `\n\n${classStr}`;
				break;
			default:
				this.defaultStyleElem.innerHTML += `\n\n${classStr}`;
		}
	}

	private static createMapKey(styles: string[], pseudoClass?: string): string {
		return `${styles.join('-')}-${pseudoClass || 'default'}`;
	}
}
