<script lang="ts" setup>
import type { InputFileElement } from '@/shared/model';
import UiText from '@/components/UiText.vue';

const props = defineProps<{
    ui: InputFileElement;
}>();
</script>

<template>
    <input
        type="file"
        :multiple="props.ui.multiple"
        :name="props.ui.name"
        :accept="props.ui.accept"
        class="border px-2 py-1"
    />
</template>
