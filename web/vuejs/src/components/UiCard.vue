<script lang="ts" setup>

import {ButtonElement, CardElement, UiDescription} from "@/shared/model";
import { useUiEvents } from "@/shared/uiEvent";
import { inject, Ref } from "vue";
import UiGeneric from "@/components/UiGeneric.vue";

const props = defineProps<{
    ui: CardElement,
}>();

const ui: Ref<UiDescription> = inject("ui")!;
const events = useUiEvents(ui);

async function onClick() {
    await events.send(props.ui.onClick);
}

</script>

<template>
    <div @click="onClick" class="block p-6 bg-white border border-gray-200 rounded-lg shadow">
      <ui-generic v-for="ui in props.ui.views" :ui="ui" />
    </div>
</template>