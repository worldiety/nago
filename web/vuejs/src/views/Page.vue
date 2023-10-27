<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>

import { useRoute } from "vue-router";
import { PageConfiguration, UiDescription } from "@/shared/model";
import { provide, ref } from "vue";
import GenericUi from "@/components/UiGeneric.vue";
import { useHttp } from "@/shared/http";

enum State {
    Loading,
    ShowUI,
    Error,
}

const route = useRoute();
const page = route.meta.page as PageConfiguration;

const http = useHttp();

const state = ref(State.Loading);
const ui = ref<UiDescription>();

// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide("ui", ui);

async function init() {
    try {
        const response = await http.request("http://localhost:3000" + page.endpoint);
        ui.value = await response.json();
        state.value = State.ShowUI;
    } catch {
        state.value = State.Error;
    }
}

init();

</script>

<template>
    <div>
        <div>Dynamic page information: {{ page }}</div>
        <div v-if="state === State.Loading">Loading UI definition…</div>
        <div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
        <generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui.renderTree" />
        <div v-else>Empty UI</div>
    </div>
</template>