import '@/assets/tailwind.css';
import '@/assets/style.css';
import App from '@/App.vue';
import i18n from '@/i18n';
import router from '@/router';
import { createPinia } from 'pinia';
import { createApp } from 'vue';

const pinia = createPinia();

const app = createApp(App);

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
app.use(pinia).use(router).use(i18n).mount('#app');
