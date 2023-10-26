import "./assets/main.css";

import App from "@/App.vue";
import router from "@/router";
import { createPinia } from "pinia";
import { createApp } from "vue";

const pinia = createPinia();

createApp(App)
  .use(pinia)
  .use(router)
  .mount("#app");
