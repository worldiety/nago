<script lang="ts" setup>

import { ButtonElement, UiDescription } from "@/shared/model";
import { useUiEvents } from "@/shared/uiEvent";
import { inject, Ref } from "vue";

const props = defineProps<{
    ui: ButtonElement,
}>();

const ui: Ref<UiDescription> = inject("ui")!;
const events = useUiEvents(ui);

async function onClick() {
    await events.send(props.ui.onClick);
}

</script>

<template>
    <button @click="onClick" class="px-2 py-1 border">
        {{ props.ui.title.value }}
    </button>
</template>