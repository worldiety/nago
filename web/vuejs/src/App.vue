<script setup lang="ts">

import { RouterView, useRoute, useRouter } from "vue-router";
import Page from "@/views/Page.vue";
import { ref } from "vue";
import { PagesConfiguration } from "@/shared/model";

const router = useRouter();
const route = useRoute();

enum State {
    LoadingRoutes,
    ShowRoutes,
    Error,
}

const state = ref(State.LoadingRoutes);

async function init() {
    try {
        const response = await fetch("http://localhost:3000/api/v1/ui/pages");
        const pages: PagesConfiguration = await response.json();

        pages.pages.forEach((page) => {
            router.addRoute({ path: page.anchor, component: Page, meta: { page } });
            console.log("registered route", page.anchor);
        });

        // Update router with current route, to load the dynamically configured page.
        await router.replace(route);

        state.value = State.ShowRoutes;
    } catch {
        state.value = State.Error;
    }
}

init();

</script>

<template>
    <div>
        <div v-if="state === State.LoadingRoutes">Loading…</div>
        <div v-if="state === State.Error">Routes could not be loaded.</div>
        <RouterView v-if="state === State.ShowRoutes" />
    </div>
</template>