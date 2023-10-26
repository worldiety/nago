<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>

import { useRoute } from "vue-router";
import { PageConfiguration, UiElement } from "@/shared/model";
import { ref } from "vue";
import GenericUi from "@/components/UiGeneric.vue";

enum State {
    Loading,
    ShowUI,
    Error,
}

const route = useRoute();
const page = route.meta.page as PageConfiguration;

const state = ref(State.Loading);
const ui = ref<UiElement>();

async function init() {
    try {
        const response = await fetch("http://localhost:3000" + page.endpoint);
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
        <div v-if="state === State.Error">Failed to fetch UI definition.</div>
        <generic-ui v-if="state === State.ShowUI" :ui="ui" />
    </div>
</template>