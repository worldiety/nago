import '@/assets/tailwind.css';
import '@/assets/style.css';
import App from '@/App.vue';
import i18n from '@/i18n';
import { createPinia } from 'pinia';
import { createApp } from 'vue';
import { eventBusKey, serviceAdapterKey } from '@/shared/injectionKeys';
import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import EventBus from '@/shared/eventbus/eventBus';

const pinia = createPinia();

const app = createApp(App);

const eventBus = new EventBus();
app.provide(serviceAdapterKey, new WebSocketAdapter(eventBus));
app.provide(eventBusKey, eventBus);

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
