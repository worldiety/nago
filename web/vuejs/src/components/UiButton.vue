<script lang="ts" setup>
import type { ButtonElement, UiDescription } from '@/shared/model';
import { useUiEvents } from '@/shared/uiEvent';
import type { Ref } from 'vue';
import { inject } from 'vue';

const props = defineProps<{
    ui: ButtonElement;
}>();

const ui: Ref<UiDescription> = inject('ui')!;
const events = useUiEvents(ui);

async function onClick() {
    await events.send(props.ui.onClick);
}
</script>

<template>
    <button class="border px-2 py-1" @click="onClick">
        {{ props.ui.title.value }}
    </button>
</template>
