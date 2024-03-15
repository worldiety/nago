import './assets/main.css';

import App from '@/App.vue';
import router from '@/router';
//import '@mdi/font/css/materialdesignicons.css';
import {createPinia} from 'pinia';
import {createApp} from 'vue';
import i18n from "@/i18n";

/*
const vuetify = createVuetify({
    components,
    directives,
    blueprint: md3,
    icons: {
        defaultSet: 'mdi',
    },
});*/


const pinia = createPinia();


const app = createApp(App)
app.directive("inline", (element: HTMLElement) => {
    const parentCss = element.classList
    for (let i = 0; i < element.children.length; i++) {
        for (let j = 0; j < parentCss.length; j++) {
            element.children.item(i).classList.add(parentCss.item(j))
        }

    }
    element.replaceWith(...element.children);
});
app.use(pinia).use(router).use(i18n).mount('#app');
