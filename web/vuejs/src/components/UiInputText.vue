<script lang="ts" setup>

import { InputTextElement } from "@/shared/model";

const props = defineProps<{
    ui: InputTextElement,
}>();

</script>

<template>
  <div>
    <label :for="props.ui.name" class="block mb-2 text-sm font-medium text-gray-900">{{props.ui.label}}</label>
    <input
        :id="props.ui.name"
        :name="props.ui.name"
        :value="props.ui.value"
        class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5" />
  </div>
</template>