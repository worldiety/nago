import type { Theme } from '@/shared/protocol/ora/theme';
import { inject } from 'vue';
import { themeManagerKey } from '@/shared/injectionKeys';
import type { Themes } from '@/shared/protocol/ora/themes';
import {LengthRule} from "@/shared/protocol/ora/lengthRule";
import {Length} from "@/shared/protocol/ora/length";

export default class ThemeManager {

	private readonly localStorageKey = 'color-theme';
	private themes: Themes|null = null;

	constructor() {
		if (!localStorage.getItem(this.localStorageKey)) {
			const userPrefersDarkTheme = window.matchMedia('(prefers-color-scheme: dark)').matches;
			localStorage.setItem(this.localStorageKey, userPrefersDarkTheme ? ThemeKey.DARK : ThemeKey.LIGHT);
		}
	}

	setThemes(themes: Themes): void {
		this.themes = themes;

		// inject our custom color and length rules
		// not sure about this, as it conflicts with the apply theme below
		for (let color of themes.colors) {
			let style = document.createElement('style');
			style.innerHTML = `
								
				@media (prefers-color-scheme: light) {
					:root {
						--${color.name}: ${color.light};
					}
				}
				
					@media (prefers-color-scheme: dark) {
					:root {
						--${color.name}: ${color.dark};
					}
				}
`;
			document.head.appendChild(style);
		}


		//themes.lengths.sort((a, b) => {
		//	return a.minWidth - b.minWidth
		//})
		for (let length of themes.lengths) {
			let style = document.createElement('style');
			style.innerHTML = `
			
					
				@media (min-width: ${length.minWidth}) {
					:root {
						--${length.name}: ${length.value};
					}
				}
`;
			document.head.appendChild(style);
		}
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

	getActiveThemeKey(): ThemeKey|null {
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
		document.getElementsByTagName('html')[0].style.setProperty('--primary', `${theme.colors.primary.h}deg ${theme.colors.primary.s}% ${theme.colors.primary.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-10', `${theme.colors.primary10.h}deg ${theme.colors.primary10.s}% ${theme.colors.primary10.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-12', `${theme.colors.primary12.h}deg ${theme.colors.primary12.s}% ${theme.colors.primary12.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-14', `${theme.colors.primary14.h}deg ${theme.colors.primary14.s}% ${theme.colors.primary14.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-17', `${theme.colors.primary17.h}deg ${theme.colors.primary17.s}% ${theme.colors.primary17.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-22', `${theme.colors.primary22.h}deg ${theme.colors.primary22.s}% ${theme.colors.primary22.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-30', `${theme.colors.primary30.h}deg ${theme.colors.primary30.s}% ${theme.colors.primary30.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-60', `${theme.colors.primary60.h}deg ${theme.colors.primary60.s}% ${theme.colors.primary60.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-70', `${theme.colors.primary70.h}deg ${theme.colors.primary70.s}% ${theme.colors.primary70.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-83', `${theme.colors.primary83.h}deg ${theme.colors.primary83.s}% ${theme.colors.primary83.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-87', `${theme.colors.primary87.h}deg ${theme.colors.primary87.s}% ${theme.colors.primary87.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-90', `${theme.colors.primary90.h}deg ${theme.colors.primary90.s}% ${theme.colors.primary90.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-92', `${theme.colors.primary92.h}deg ${theme.colors.primary92.s}% ${theme.colors.primary92.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-94', `${theme.colors.primary94.h}deg ${theme.colors.primary94.s}% ${theme.colors.primary94.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-96', `${theme.colors.primary96.h}deg ${theme.colors.primary96.s}% ${theme.colors.primary96.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--primary-98', `${theme.colors.primary98.h}deg ${theme.colors.primary98.s}% ${theme.colors.primary98.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--secondary', `${theme.colors.secondary.h}deg ${theme.colors.secondary.s}% ${theme.colors.secondary.l}%`);
		document.getElementsByTagName('html')[0].style.setProperty('--tertiary', `${theme.colors.tertiary.h}deg ${theme.colors.tertiary.s}% ${theme.colors.tertiary.l}%`);

		// TODO apply themes custom colors or move it into each theme?
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
