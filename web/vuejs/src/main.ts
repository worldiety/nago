import "./assets/main.css";

import App from "@/App.vue";
import router from "@/router";
import { createPinia } from "pinia";
import { createApp } from "vue";
import {createRouter, createWebHashHistory} from "vue-router";
import generic from "@/components/generic.vue";

const router = createRouter();

// TODO clean up
const res = await fetch("http://localhost:8080/api/v1/ui/pages").then(r=>r.json());
res["pages"].forEach(r=>{
    router.addRoute({ path: r.anchor, component: generic })
    console.log("registered route "+r.anchor)
});

const pinia = createPinia();

createApp(App)
  .use(pinia)
  .use(router)
  .mount("#app");
