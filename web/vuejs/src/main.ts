import { createApp } from 'vue';
import App from '@/App.vue';
import { UploadRepository } from '@/api/upload/uploadRepository';
import i18n from '@/i18n';
import { createPinia } from 'pinia';
import EventBus from '@/shared/eventbus/eventBus';
import { eventBusKey, serviceAdapterKey, themeManagerKey, uploadRepositoryKey } from '@/shared/injectionKeys';
import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import ThemeManager from '@/shared/themeManager';
import '@/assets/style.css';
import '@/assets/tailwind.css';

const pinia = createPinia();

const app = createApp(App);

const eventBus = new EventBus();
app.provide(serviceAdapterKey, new WebSocketAdapter(eventBus));
app.provide(eventBusKey, eventBus);
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
