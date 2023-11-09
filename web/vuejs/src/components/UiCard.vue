<script lang="ts" setup>
import type { CardElement, UiDescription } from '@/shared/model';
import { ButtonElement } from '@/shared/model';
import { useUiEvents } from '@/shared/uiEvent';
import type { Ref } from 'vue';
import { inject } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';

const props = defineProps<{
    ui: CardElement;
}>();

const ui: Ref<UiDescription> = inject('ui')!;
const events = useUiEvents(ui);

async function onClick() {
    await events.send(props.ui.onClick);
}
</script>

<template>
    <div class="block rounded-lg border border-gray-200 bg-white p-6 shadow" @click="onClick">
        <ui-generic v-for="ui in props.ui.views" :ui="ui" />
    </div>
</template>
