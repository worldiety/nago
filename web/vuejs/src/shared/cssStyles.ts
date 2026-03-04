export class CssStyles {
	private readonly targetId: string;
	private readonly styleElem: HTMLStyleElement;
	private activeStyles: string[] = [];
	private defaultStyles: string[] = [];
	private focusStyles: string[] = [];
	private hoverStyles: string[] = [];

	constructor(targetId: string) {
		this.targetId = targetId;
		const elem = document.createElement('style') as HTMLStyleElement;
		elem.setAttribute('data-target', targetId);
		document.body.appendChild(elem);
		this.styleElem = elem;
	}

	public setStyles(defaults: string[], hover: string[], focus: string[], active: string[]) {
		this.activeStyles = active;
		this.defaultStyles = defaults;
		this.focusStyles = focus;
		this.hoverStyles = hover;
		this.update();
	}

	public setActiveStyles(styles: string[]) {
		this.activeStyles = styles;
		this.update();
	}

	public setDefaultStyles(styles: string[]) {
		this.defaultStyles = styles;
		this.update();
	}

	public setFocusStyles(styles: string[]) {
		this.focusStyles = styles;
		this.update();
	}

	public setHoverStyles(styles: string[]) {
		this.hoverStyles = styles;
		this.update();
	}

	public remove() {
		this.styleElem.remove();
	}

	private update() {
		const defaultStyles = this.getCssForPseudoClass();
		const activeStyles = this.getCssForPseudoClass('active');
		const focusStyles = this.getCssForPseudoClass('focus');
		const hoverStyles = this.getCssForPseudoClass('hover');
		this.styleElem.innerHTML = `${defaultStyles}\n\n${hoverStyles}\n\n${focusStyles}\n\n${activeStyles}`; // order of styles matters
	}

	private getCssForPseudoClass(pseudoClass?: 'active' | 'focus' | 'hover') {
		let styles: string[];
		switch (pseudoClass) {
			case 'active':
				styles = this.activeStyles;
				break;
			case 'focus':
				styles = this.focusStyles;
				break;
			case 'hover':
				styles = this.hoverStyles;
				break;
			default:
				styles = this.defaultStyles;
		}

		const selectors = [`#${this.targetId}${pseudoClass ? `:${pseudoClass}` : ''}`];
		if (pseudoClass === 'focus') {
			selectors.push(`#${this.targetId}${pseudoClass ? `:focus-visible` : ''}`);
		}

		return `${selectors.join(',\n')} {
			${styles.join(';\n')}
		}`;
	}
}
