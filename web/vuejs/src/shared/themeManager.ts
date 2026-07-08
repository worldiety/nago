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
			localStorage.setItem(this.localStorageKey, ThemeKey.SYSTEM);
		}

		window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', this.applyActiveTheme.bind(this));
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
			default:
				this.applySystemTheme();
				break;
		}
	}

	getActiveThemeKey(): ThemeKey | null {
		const activeThemeKey = localStorage.getItem(this.localStorageKey);
		return activeThemeKey ? (activeThemeKey as ThemeKey) : null;
	}

	applySystemTheme(): void {
		if (!this.themes) {
			return;
		}

		const darkModeMql = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)');

		if (darkModeMql && darkModeMql.matches) {
			this.applyTheme(this.themes.dark);
		} else {
			this.applyTheme(this.themes.light);
		}

		document.getElementsByTagName('html')[0].classList.remove('lightmode');
		document.getElementsByTagName('html')[0].classList.remove('darkmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.SYSTEM);
	}

	applyLightmodeTheme(): void {
		if (!this.themes) {
			return;
		}

		this.applyTheme(this.themes.light);
		document.getElementsByTagName('html')[0].classList.remove('darkmode');
		document.getElementsByTagName('html')[0].classList.add('lightmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.LIGHT);
	}

	applyDarkmodeTheme(): void {
		if (!this.themes) {
			return;
		}

		this.applyTheme(this.themes.dark);
		document.getElementsByTagName('html')[0].classList.remove('lightmode');
		document.getElementsByTagName('html')[0].classList.add('darkmode');
		localStorage.setItem(this.localStorageKey, ThemeKey.DARK);
	}

	private applyTheme(theme?: Theme): void {
		if (!theme) {
			return;
		}

		const elem = document.getElementsByTagName('html')[0];

		if (theme.colors) {
			// TODO this is underspecified, because the namespace is not involved in the colorname which break the logic namespacing
			theme.colors.value.forEach((val) => {
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

		// update the global html document style
		const color = getComputedStyle(document.documentElement).getPropertyValue('--M1').trim();

		document.getElementById('themeColorMeta')!.setAttribute('content', color);
	}
}

export enum ThemeKey {
	LIGHT = 'light',
	DARK = 'dark',
	SYSTEM = 'system',
}

export function useThemeManager(): ThemeManager {
	const themeManager = inject(themeManagerKey);
	if (!themeManager) {
		throw new Error('Could not inject ThemeManager as it is undefined');
	}

	return themeManager;
}
