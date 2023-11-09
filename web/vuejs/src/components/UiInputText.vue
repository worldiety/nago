<script lang="ts" setup>
import type { InputTextElement } from '@/shared/model';

const props = defineProps<{
    ui: InputTextElement;
}>();
</script>

<template>
    <div>
        <label :for="props.ui.name" class="mb-2 block text-sm font-medium text-gray-900">{{ props.ui.label }}</label>
        <input
            :id="props.ui.name"
            :name="props.ui.name"
            :value="props.ui.value"
            class="block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500"
        />
    </div>
</template>
