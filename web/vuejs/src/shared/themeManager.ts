/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { inject } from 'vue';
import { themeManagerKey } from '@/shared/injectionKeys';
import { Locale, Theme, Themes } from '@/shared/proto/nprotoc_gen';

export default class ThemeManager {
	private readonly localStorageKey = 'color-theme';
	private themes: Themes | null = null;
	public activeLocale: Locale;

	constructor() {
		this.activeLocale = '';
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
		return activeThemeKey ? (activeThemeKey as ThemeKey) : null;
	}

	toggleDarkMode(): void {
		if (!this.themes) {
			return;
		}

		if (localStorage.getItem(this.localStorageKey) === ThemeKey.LIGHT) {
			this.applyDarkmodeTheme();
			return;
		} else if (localStorage.getItem(this.localStorageKey) === ThemeKey.DARK) {
			this.applyLightmodeTheme();
		}
	}

	applyLightmodeTheme(): void {
		if (!this.themes) {
			return;
		}

		this.applyTheme(this.themes.light);
		document.getElementsByTagName('html')[0].classList.remove('darkmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.LIGHT);
	}

	applyDarkmodeTheme(): void {
		if (!this.themes) {
			return;
		}

		this.applyTheme(this.themes.dark);
		document.getElementsByTagName('html')[0].classList.add('darkmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.DARK);
	}

	private applyTheme(theme?: Theme): void {
		if (!theme) {
			return;
		}

		let elem = document.getElementsByTagName('html')[0];

		if (theme.colors) {
			// TODO this is underspecified, because the namespace is not involved in the colorname which break the logic namespacing
			theme.colors.value.forEach((val, key) => {
				val.value.forEach((colorVal, colorName) => {
					elem.style.setProperty(`--${colorName}`, colorVal);
				});
			});
		}

		if (theme.lengths) {
			theme.lengths.value.forEach((lengthVal, lengthName) => {
				elem.style.setProperty(`--${lengthName}`, lengthVal);
			});
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
