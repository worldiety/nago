<script lang="ts" setup>

import { InputTextElement } from "@/shared/model";

const props = defineProps<{
    ui: InputTextElement,
}>();

</script>

<template>
    <input
        :name="props.ui.name"
        :value="props.ui.value"
        class="px-2 py-1 border" />
</template>