/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { createApp } from 'vue';
import App from '@/App.vue';
import { UploadRepository } from '@/api/upload/uploadRepository';
import i18n from '@/i18n';
import { createPinia } from 'pinia';
import { serviceAdapterKey, themeManagerKey, uploadRepositoryKey } from '@/shared/injectionKeys';
import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import ThemeManager from '@/shared/themeManager';
import '@/assets/style.css';
import '@/assets/tailwind.css';

const pinia = createPinia();

const app = createApp(App);

app.provide(serviceAdapterKey, new WebSocketAdapter());
app.provide(uploadRepositoryKey, new UploadRepository());
app.provide(themeManagerKey, new ThemeManager());

app.directive('inline', (element: HTMLElement) => {
	const parentCss = element.classList;
	for (let i = 0; i < element.children.length; i++) {
		for (let j = 0; j < parentCss.length; j++) {
			const parentCssClass = parentCss.item(j);
			const childElement = element.children.item(i);
			if (childElement && parentCssClass) {
				childElement.classList.add(parentCssClass);
			}
		}
	}
	element.replaceWith(...Object.values(element.children));
});
app.use(pinia).use(i18n).mount('#app');
