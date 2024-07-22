import type {Theme} from '@/shared/protocol/ora/theme';
import {inject} from 'vue';
import {themeManagerKey} from '@/shared/injectionKeys';
import type {Themes} from '@/shared/protocol/ora/themes';

export default class ThemeManager {

	private readonly localStorageKey = 'color-theme';
	private themes: Themes | null = null;

	constructor() {
		if (!localStorage.getItem(this.localStorageKey)) {
			const userPrefersDarkTheme = window.matchMedia('(prefers-color-scheme: dark)').matches;
			localStorage.setItem(this.localStorageKey, userPrefersDarkTheme ? ThemeKey.DARK : ThemeKey.LIGHT);
		}
	}

	setThemes(themes: Themes): void {
		this.themes = themes;

	}

	applyActiveTheme(): void {
		if (!this.themes) {
			return;
		}

		switch (localStorage.getItem(this.localStorageKey)) {
			case ThemeKey.LIGHT:
				this.applyLightmodeTheme();
				break;
			case ThemeKey.DARK:
				this.applyDarkmodeTheme();
				break;
		}
	}

	getActiveThemeKey(): ThemeKey | null {
		const activeThemeKey = localStorage.getItem(this.localStorageKey);
		return activeThemeKey ? activeThemeKey as ThemeKey : null;
	}

	toggleDarkMode(): void {
		if (!this.themes) {
			return;
		}

		if (localStorage.getItem(this.localStorageKey) === ThemeKey.LIGHT) {
			this.applyDarkmodeTheme()
			return;
		} else if (localStorage.getItem(this.localStorageKey) === ThemeKey.DARK) {
			this.applyLightmodeTheme();
		}
	}

	private applyLightmodeTheme(): void {
		if (!this.themes) {
			return;
		}

		this.applyTheme(this.themes.light);
		document.getElementsByTagName('html')[0].classList.remove('darkmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.LIGHT);
	}

	private applyDarkmodeTheme(): void {
		if (!this.themes) {
			return;
		}

		this.applyTheme(this.themes.dark);
		document.getElementsByTagName('html')[0].classList.add('darkmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.DARK);
	}

	private applyTheme(theme: Theme): void {
		let elem = document.getElementsByTagName('html')[0];

		if (theme.colors) {
			for (const [ns, nameValuePairs] of Object.entries(theme.colors)) {
				for (const [colorName, colorValue] of Object.entries(nameValuePairs)) {
					elem.style.setProperty(`--${colorName}`, colorValue)
					console.log(colorName,"=",colorValue)
				}

			}
		}

		if (theme.lengths.customLengths) {
			for (const [key, value] of Object.entries(theme.lengths.customLengths)) {
				elem.style.setProperty(`--${key}`, value)
			}
		}

	}
}

export enum ThemeKey {
	LIGHT = 'light',
	DARK = 'dark',
}

export function useThemeManager(): ThemeManager {
	const themeManager = inject(themeManagerKey);
	if (!themeManager) {
		throw new Error('Could not inject ThemeManager as it is undefined');
	}

	return themeManager;
}
