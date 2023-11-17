<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router';
import Page from '@/views/Page.vue';
import { ref } from 'vue';
import type { PagesConfiguration } from '@/shared/model';

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
        const response = await fetch(import.meta.env.VITE_HOST_BACKEND + 'api/v1/ui/application');
        const app: PagesConfiguration = await response.json();

        app.pages.forEach((page) => {
            let anchor = page.anchor.replaceAll("{",":")
            anchor = anchor.replaceAll("}","?")
            anchor = anchor.replaceAll("-","_") //OMG
            router.addRoute({ path: anchor, component: Page, meta: { page } });
            console.log('registered route', anchor);
        });

        // Update router with current route, to load the dynamically configured page.
        await router.replace(route);

        state.value = State.ShowRoutes;
    } catch (e) {
        console.log(e);
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
