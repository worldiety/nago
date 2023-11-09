<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>
import { useRoute } from 'vue-router';
import type { PageConfiguration, Scaffold} from '@/shared/model';
import { UiDescription } from '@/shared/model';
import { provide, ref } from 'vue';
import GenericUi from '@/components/UiGeneric.vue';
import { useHttp } from '@/shared/http';
import router from '@/router';

enum State {
    Loading,
    ShowUI,
    Error,
}

const route = useRoute();
const page = route.meta.page as PageConfiguration;

const http = useHttp();

const state = ref(State.Loading);
const ui = ref<Scaffold>();

// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide('ui', ui);

async function init() {
    try {
        const pageUrl = import.meta.env.VITE_HOST_BACKEND + page.endpoint.slice(1);
        const response = await http.request(pageUrl);
        ui.value = await response.json();
        state.value = State.ShowUI;
        console.log(pageUrl);
        console.log('got value', ui.value);
    } catch {
        state.value = State.Error;
    }
}

init();
</script>

<template>
    <div>
        <!--  <div>Dynamic page information: {{ page }}</div> -->
        <div v-if="state === State.Loading">Loading UI definition…</div>
        <div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
        <generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" />
        <div v-else>Empty UI</div>
    </div>
</template>
