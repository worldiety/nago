import './assets/main.css';

import App from '@/App.vue';
import router from '@/router';
import '@mdi/font/css/materialdesignicons.css';
import { createPinia } from 'pinia';
import { createApp } from 'vue';
import { md3 } from 'vuetify/blueprints';

import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import 'vuetify/styles';
import { VDataTable } from 'vuetify/labs/VDataTable'

const vuetify = createVuetify({
    components,
    directives,
    blueprint: md3,
    icons: {
        defaultSet: 'mdi',
    },
});

const pinia = createPinia();

createApp(App).use(vuetify).use(pinia).use(router).mount('#app');
