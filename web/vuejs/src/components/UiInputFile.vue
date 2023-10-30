<script lang="ts" setup>

import { InputFileElement } from "@/shared/model";
import UiText from "@/components/UiText.vue";

const props = defineProps<{
    ui: InputFileElement,
}>();

</script>

<template>
    <input
        type="file"
        :multiple="props.ui.multiple"
        :name="props.ui.name"
        :accept="props.ui.accept"
        class="px-2 py-1 border" />
</template>