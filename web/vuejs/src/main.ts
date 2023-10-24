import "./assets/main.css";

import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./App.vue";
import router from "./router";
import {createRouter, createWebHashHistory} from "vue-router";
import generic from "@/components/generic.vue";


const res = await fetch("http://localhost:8080/api/v1/ui/pages").then(r=>r.json())

const router = createRouter({
    history: createWebHashHistory(),
    routes: [],
})

console.log("hello")

res["pages"].forEach(r=>{
    router.addRoute({ path: r.anchor, component: generic })
    console.log("registered route "+r.anchor)
} )


const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);

app.mount("#app");
